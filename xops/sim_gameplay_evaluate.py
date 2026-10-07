#!/usr/bin/env python3
"""Frozen, resumable protected complete-game campaign; no implicit retries."""
import argparse
import copy
from contextlib import ExitStack
import fcntl
from fractions import Fraction
import hashlib
import json
import math
import os
from pathlib import Path
import random
import re
import signal
import stat
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parent.parent
ARTIFACTS = ROOT / 'sims/artifacts'
SCHEMA = 'cgms-gameplay-campaign-v1'
SCHEMA_V2 = 'cgms-gameplay-campaign-v2'
SCHEMA_V3 = 'cgms-gameplay-campaign-v3'
ANALYSIS = 'python-mt19937-seed-block-percentile-v1'


def require(ok, message):
    if not ok:
        raise ValueError(message)


def digest(data):
    return hashlib.sha256(data).hexdigest()


def encoded(value):
    return json.dumps(value, sort_keys=True, separators=(',', ':'), allow_nan=False).encode()


def load(path):
    def unique(pairs):
        out = {}
        for key, value in pairs:
            require(key not in out, 'duplicate JSON key')
            out[key] = value
        return out
    return json.loads(Path(path).read_bytes(), object_pairs_hook=unique, parse_constant=lambda _: (_ for _ in ()).throw(ValueError('nonfinite JSON')))


def safe_path(path):
    path = Path(path).absolute()
    require('..' not in path.parts and path.is_relative_to(ARTIFACTS), 'artifact path outside protected tree')
    require(not any(p.is_symlink() for p in (path, *path.parents)), 'artifact symlink')
    return path


def immutable_json(path, value):
    with open(path, 'xb') as out:
        out.write(encoded(value))


def build_units(profiles, shard_blocks):
    require(type(shard_blocks) is int and 1 <= shard_blocks <= 5, 'shard bound')
    require(isinstance(profiles,list) and 1<=len(profiles)<=2, 'one or two population profiles required')
    out = []
    populations = set()
    for profile in profiles:
        pop = profile['population']
        require(pop in (3,4) and pop not in populations, 'population must occur once')
        populations.add(pop)
        blocks = profile['block_ids']
        require(blocks and len(set(blocks)) == len(blocks), 'unique nonempty seed blocks required')
        rotations = profile['rotations']
        require(len(rotations) >= pop and len({r['id'] for r in rotations}) == len(rotations), 'complete planned rotations required')
        schedules=[tuple(r['seats']) for r in rotations]
        participants=set(schedules[0])
        require(len(participants)==pop and len(set(schedules))==len(schedules), 'distinct seat schedules required')
        require(all(len(a)==pop and set(a)==participants for a in schedules), 'rotation participants mismatch')
        require(len(schedules)%pop==0 and all(sum(a[seat]==participant for a in schedules)==len(schedules)//pop for seat in range(pop) for participant in participants), 'complete balanced seat coverage required')
        require(not set(blocks).intersection(profile.get('development_block_ids',[])), 'development and evaluation blocks overlap')
        require(not profile.get('grid'), 'campaign does not combine parameter grids')
        for start in range(0, len(blocks), shard_blocks):
            experiment = copy.deepcopy(profile)
            experiment['block_ids'] = blocks[start:start+shard_blocks]
            matches = len(experiment['block_ids']) * len(rotations) * 2
            experiment['budgets']['max_matches'] = matches
            out.append({'id': f'p{pop}-s{start//shard_blocks:04d}', 'population': pop, 'experiment': experiment, 'planned_matches': matches, 'planned_games': matches * profile['games_per_match']})
    # Fixed population-interleaved resource order is independent of outcomes.
    out.sort(key=lambda unit: (int(unit['id'].split('-s')[1]), unit['population']))
    return out


def replay_selection(plan, units):
    """Select by immutable logical IDs, never by observed score or completion."""
    all_ids={u['id'] for u in units}
    if plan['schema']==SCHEMA: return all_ids
    require(plan['schema'] in (SCHEMA_V2,SCHEMA_V3),'campaign schema')
    chosen=plan['replay_units']
    require(type(chosen) is list and all(type(v) is str for v in chosen),'replay selection list')
    require(len(chosen)==len(set(chosen)) and set(chosen)<=all_ids,'invalid frozen replay units')
    require(set(plan['worker_check_units'])<=set(chosen),'worker checks require selected replay')
    return set(chosen)


def verify_shard_inputs(dest, units):
    for unit in units:
        require(load(dest/(unit['id']+'-experiment.json')) == unit['experiment'], 'derived shard differs from frozen profile')


def verify_state_plan(state, units):
    fields=('id','population','planned_matches','planned_games')
    expected=[{k:u[k] for k in fields} for u in units]
    actual=[{k:u[k] for k in fields} for u in state['units']]
    require(actual==expected,'resume unit identity/order/denominators changed')


def verify_effective_experiment(directory, frozen_experiment, worker_override=None):
    expected=copy.deepcopy(frozen_experiment)
    if worker_override is not None:
        expected['budgets']['workers']=worker_override
    require(load(directory/'effective-experiment.json')==expected,'effective experiment differs from frozen shard')


def inspect_manifest(directory, expected_code, binary_hash, frozen_experiment=None, worker_override=None):
    directory = safe_path(directory)
    m = load(directory / 'manifest.json')
    require(m.get('schema_version') == 'cgms-manifest-v1' and m.get('exit_code') == expected_code, 'manifest/process disagreement')
    require(m.get('privacy') == 'restricted', 'manifest privacy')
    require(m.get('provenance', {}).get('binary_sha256') == binary_hash, 'binary provenance mismatch')
    seen = set()
    for item in m['artifacts']:
        rel = Path(item['path'])
        require(not rel.is_absolute() and '..' not in rel.parts and item['path'] not in seen, 'artifact path or duplicate')
        seen.add(item['path'])
        path = safe_path(directory / rel)
        raw = path.read_bytes()
        require(len(raw) == item['bytes'] and digest(raw) == item['sha256'], 'artifact integrity mismatch')
    require('report.md' in seen, 'missing report')
    if frozen_experiment is not None:
        require('effective-experiment.json' in seen,'missing frozen experiment artifact')
        verify_effective_experiment(directory,frozen_experiment,worker_override)
    return m


def cluster_interval(values, seed, resamples):
    """Each value is one whole paired seed-block mean; never bootstrap games."""
    require(type(resamples) is int and 100 <= resamples <= 100000, 'bootstrap resample bound')
    if not values:
        return {'complete_blocks': 0, 'mean': None, 'lower_95': None, 'upper_95': None, 'analysis': ANALYSIS}
    values = [Fraction(v) for v in values]
    if len(values)<2:
        return {'complete_blocks':len(values),'mean':str(values[0]),'lower_95':None,'upper_95':None,'analysis':ANALYSIS,'precision_note':'At least two independent seed blocks required; few-block intervals remain unreliable.'}
    rng = random.Random(str(seed))
    means = sorted(sum((values[rng.randrange(len(values))] for _ in values), Fraction()) / len(values) for _ in range(resamples))
    return {'complete_blocks': len(values), 'mean': str(sum(values, Fraction())/len(values)), 'lower_95': str(means[math.floor(.025*(resamples-1))]), 'upper_95': str(means[math.ceil(.975*(resamples-1))]), 'analysis': ANALYSIS}


def _beta_fraction(a, b, x):
    """Modified Lentz continued fraction for regularized incomplete beta."""
    floor = 1e-300
    qab, qap, qam = a+b, a+1, a-1
    c, d = 1.0, 1-qab*x/qap
    if abs(d)<floor: d=floor
    d=1/d
    h=d
    for m in range(1, 1001):
        aa=m*(b-m)*x/((qam+2*m)*(a+2*m))
        d=1+aa*d; c=1+aa/c
        if abs(d)<floor: d=floor
        if abs(c)<floor: c=floor
        d=1/d; h*=d*c
        aa=-(a+m)*(qab+m)*x/((a+2*m)*(qap+2*m))
        d=1+aa*d; c=1+aa/c
        if abs(d)<floor: d=floor
        if abs(c)<floor: c=floor
        d=1/d; change=d*c; h*=change
        if abs(change-1)<3e-14: return h
    raise ValueError('incomplete beta did not converge')


def _regularized_beta(x, a, b):
    if x<=0: return 0.0
    if x>=1: return 1.0
    factor=math.exp(math.lgamma(a+b)-math.lgamma(a)-math.lgamma(b)+a*math.log(x)+b*math.log1p(-x))
    if x<(a+1)/(a+b+2): return factor*_beta_fraction(a,b,x)/a
    return 1-factor*_beta_fraction(b,a,1-x)/b


def student_critical(probability, degrees):
    """Upper Student t quantile; deterministic bisection of the beta CDF."""
    require(type(degrees) is int and degrees>0 and .5<probability<1, 'Student parameters')
    def cdf(t):
        return 1-.5*_regularized_beta(degrees/(degrees+t*t), degrees/2, .5)
    lower,upper=0.0,1.0
    while cdf(upper)<probability:
        upper*=2
        require(upper<1e12,'Student quantile outside numerical bound')
    for _ in range(80):
        mid=(lower+upper)/2
        if cdf(mid)<probability: lower=mid
        else: upper=mid
    return (lower+upper)/2


def prospective_interval(values, comparisons, alpha=.05):
    """New-cohort inference only. One exact per-game mean per independent block.

    Bonferroni simultaneous two-sided intervals. Student inference is exact for
    normal block means, asymptotic otherwise; 1,000 blocks is the confirmation
    floor, not a normality guarantee. Historical bootstrap output is unchanged.
    """
    require(type(comparisons) is int and 1<=comparisons<=10000,'comparison family bound')
    require(0<alpha<1,'alpha bound')
    values=[Fraction(v) for v in values]
    n=len(values)
    mean=sum(values,Fraction())/n if n else None
    result={'analysis':'seed-block-student-bonferroni-v1','independent_blocks':n,
            'mean':str(mean) if mean is not None else None,'sample_variance':None,
            'lower':None,'upper':None,'family_comparisons':comparisons,'family_alpha':alpha}
    if n<2: return result
    variance=sum(((v-mean)**2 for v in values),Fraction())/(n-1)
    result['sample_variance']=str(variance)
    if variance==0:
        result['degenerate']=True
        return result
    se=math.sqrt(float(variance)/n)
    critical=student_critical(1-alpha/(2*comparisons),n-1)
    result.update(standard_error=se,critical=critical,lower=float(mean)-critical*se,upper=float(mean)+critical*se)
    return result


def classify_interval(result, improvement, equivalence_margin, min_blocks=1000):
    """Conservative interval inclusion equivalence (stricter than ordinary TOST).

    With family K this is TOST at alpha/(2K), controlling all directional
    claims together. Zero-overlap/non-significance never establishes equivalence.
    """
    require(improvement>0 and equivalence_margin>0 and min_blocks>=2,'decision thresholds')
    out={key:'inconclusive' for key in ('strength','positive_improvement','noninferiority','equivalence')}
    lo,hi=result['lower'],result['upper']
    if result['independent_blocks']<min_blocks or lo is None or hi is None: return out
    if lo>improvement: out['strength']='supported'
    elif hi<improvement: out['strength']='refuted'
    if lo>0: out['positive_improvement']='supported'
    elif hi<0: out['positive_improvement']='refuted'
    if lo>-equivalence_margin: out['noninferiority']='supported'
    elif hi<-equivalence_margin: out['noninferiority']='refuted'
    if lo>-equivalence_margin and hi<equivalence_margin: out['equivalence']='supported'
    elif lo>equivalence_margin or hi<-equivalence_margin: out['equivalence']='refuted'
    return out


def decode_score(value):
    """Decode the canonical game.Amount JSON contract, without rounding."""
    require(type(value) is dict and set(value)=={'numerator','denominator'}, 'rational object required')
    numerator,denominator=value['numerator'],value['denominator']
    integer=r'(0|-[1-9][0-9]*|[1-9][0-9]*)'
    require(type(numerator) is str and type(denominator) is str and re.fullmatch(integer,numerator) and re.fullmatch(integer,denominator), 'invalid rational integer')
    n,d=int(numerator),int(denominator)
    require(d>0, 'denominator must be positive')
    result=Fraction(n,d)
    require(str(result.numerator)==numerator and str(result.denominator)==denominator, 'noncanonical rational')
    return result


def paired_blocks(profile, outcomes):
    rows = {}
    for row in outcomes:
        key = (row['block'], row['rotation'], row['treatment'])
        require(key not in rows, 'duplicate evaluation unit')
        require(row['block'] in profile['block_ids'] and row['rotation'] in {r['id'] for r in profile['rotations']} and row['treatment'] in ('baseline','candidate'), 'unplanned evaluation unit')
        rows[key] = row
    values, statuses = [], {}
    focal = profile['candidate']['participant_id']
    for block in profile['block_ids']:
        diffs = []
        for rotation in profile['rotations']:
            pair = []
            for treatment in ('baseline', 'candidate'):
                row = rows.get((block, rotation['id'], treatment))
                status = 'not-started' if row is None else row['status']
                statuses[status] = statuses.get(status, 0) + 1
                if row is None or not row['match_complete']:
                    pair.append(None)
                else:
                    require(row['finalized_games'] == profile['games_per_match'], 'false finalized match')
                    require(row['seats'] == rotation['seats'], 'seat schedule mismatch')
                    require(type(row['scores']) is list and len(row['scores'])==len(rotation['seats']), 'score vector length')
                    scores=[decode_score(value) for value in row['scores']]
                    pair.append(scores[row['seats'].index(focal)])
            if None not in pair:
                diffs.append(pair[1]-pair[0])
        if len(diffs) == len(profile['rotations']):
            values.append(sum(diffs, Fraction())/len(diffs))
    return values, statuses


def validate_outcomes(profile, rows, complete):
    expected={(b,r['id'],t) for b in profile['block_ids'] for r in profile['rotations'] for t in ('baseline','candidate')}
    actual=set()
    for row in rows:
        key=(row['block'],row['rotation'],row['treatment'])
        require(key in expected and key not in actual,'unplanned or duplicate outcome unit')
        actual.add(key)
        require(row['population']==profile['population'] and row['planned_game_slots']==profile['games_per_match'],'outcome population/game plan')
        require(row.get('status') in {'complete','not-started','invalid-input','unsupported','replay-invariant-failure','budget-exhausted','io-failure','canceled'},'unknown public status')
        instances=row.get('started_instances',row.get('started_games',0));voided=row.get('voided_games',0)
        require(type(instances) is int and type(voided) is int and 0<=voided<=instances and instances>=row.get('started_games',0),'invalid physical instance denominator')
        counts=[row.get(k,0) for k in ('finalized_games','board_ended_games','started_games')]
        require(all(type(v) is int for v in counts) and 0<=counts[0]<=counts[1]<=counts[2]<=profile['games_per_match'],'invalid game denominators')
        require(type(row['match_complete']) is bool and row['match_complete']==(counts[0]==profile['games_per_match']),'match completion mismatch')
        require(not complete or row['match_complete'],'complete shard has incomplete match')
    require(actual==expected,'missing planned outcomes')


def verified_analysis_rows(profile, rows, complete):
    try:
        validate_outcomes(profile,rows,complete)
        # Validate score/seat access before admitting any row from this shard.
        paired_blocks(profile,rows)
        return rows
    except (ValueError,KeyError,TypeError,IndexError,ZeroDivisionError):
        return [{'block':b,'rotation':r['id'],'treatment':t,'status':'unverified-outcomes','match_complete':False,'unknown_game_slots':profile['games_per_match']} for b in profile['block_ids'] for r in profile['rotations'] for t in ('baseline','candidate')]


def analysis_source_paths(plan):
    sources=plan['analysis_sources']
    require(type(sources) is list and 2<=len(sources)<=16 and all(type(s) is str for s in sources),'analysis source list')
    paths=[Path(s).absolute() for s in sources]
    require(len(set(paths))==len(paths),'duplicate analysis source')
    required={Path(__file__).resolve(),ROOT/'xops/sim_claims.py'}
    require(required<=set(paths),'runner and claim extraction sources required')
    for path in paths:
        require(path.is_relative_to(ROOT) and '..' not in path.parts and path.suffix=='.py','analysis source outside workspace Python')
        require(not any(p.is_symlink() for p in (path,*path.parents)),'analysis source symlink')
        require(path.is_file() and path.stat().st_size<=4*1024*1024,'analysis source size/type')
    return paths


def verify_frozen_files(frozen):
    for entry in frozen.values():
        require(digest(Path(entry['path']).read_bytes())==entry['sha256'],'frozen input changed during campaign')


def retain_analysis_sources(dest, frozen):
    for key, entry in frozen.items():
        if key.startswith('analysis-source-'):
            raw=Path(entry['path']).read_bytes()
            require(digest(raw)==entry['sha256'],'analysis source changed before retention')
            with open(dest/(key+'.py'),'xb') as out:out.write(raw)


def verify_analysis_sources(dest, frozen):
    for key, entry in frozen.items():
        if key.startswith('analysis-source-'):
            require(digest(Path(entry['path']).read_bytes())==entry['sha256'],'analysis source changed')
            require(digest((dest/(key+'.py')).read_bytes())==entry['sha256'],'retained analysis source changed')


def frozen_inputs(binary, plan_path, plan):
    required = {'schema', 'config', 'profiles', 'source_manifest', 'shard_blocks', 'wall_seconds', 'storage_bytes', 'worker_check_units', 'worker_check_workers', 'analysis_seed', 'resamples'}
    if plan.get('schema') in (SCHEMA_V2,SCHEMA_V3): required.add('replay_units')
    if plan.get('schema') == SCHEMA_V3: required.add('analysis_sources')
    require(set(plan) == required and plan['schema'] in (SCHEMA, SCHEMA_V2, SCHEMA_V3), 'campaign plan fields')
    require(type(plan['wall_seconds']) is int and plan['wall_seconds'] > 0 and type(plan['storage_bytes']) is int and plan['storage_bytes'] > 0, 'resource limits')
    require(type(plan['worker_check_workers']) is int and 1 <= plan['worker_check_workers'] <= 64, 'worker bound')
    files = {'plan': str(plan_path), 'binary': str(binary), 'config': str(Path(plan['config']).resolve()), 'source_manifest': str(Path(plan['source_manifest']).resolve())}
    for i, path in enumerate(plan['profiles']):
        files[f'profile-{i}'] = str(Path(path).resolve())
    if plan['schema']==SCHEMA_V3:
        for i,path in enumerate(analysis_source_paths(plan)):
            files[f'analysis-source-{i}']=str(path)
    return {key: {'path': path, 'sha256': digest(Path(path).read_bytes())} for key, path in files.items()}


def checkpoint_wall_accounting(state, now=None):
    if not state.get('invocation_active'):
        return
    now=time.time() if now is None else now
    previous=state['wall_checkpoint_epoch']
    state['elapsed_seconds']+=max(0,now-previous)
    state['wall_checkpoint_epoch']=max(previous,now)


def start_wall_accounting(state, now=None):
    now=time.time() if now is None else now
    # An unclean invocation is charged through resume, including downtime.
    # Conservatively overcharging is preferable to silently renewing the budget.
    checkpoint_wall_accounting(state,now)
    state['invocation_active']=True
    state['invocation_started_epoch']=now
    state['wall_checkpoint_epoch']=now


def save_state(dest, state):
    checkpoint_wall_accounting(state)
    state['revision'] += 1
    immutable_json(dest / f'state-{state["revision"]:06d}.json', state)
    temporary = dest / f'state-{state["revision"]:06d}.pending'
    with open(temporary, 'xb') as out:
        out.write(encoded(state)); out.flush(); os.fsync(out.fileno())
    os.replace(temporary, dest/'state.json')


def storage_bytes(dest):
    # One metadata query per entry preserves readable-tree logical byte totals,
    # including each hardlink path. A concurrent atomic publication can remove
    # an enumerated path: discard the partial count and rescan once on ENOENT.
    # Persistent churn and all other errors fail closed, never undercount silently.
    for attempt in range(2):
        total = 0
        pending = [dest]
        try:
            while pending:
                with os.scandir(pending.pop()) as entries:
                    for entry in entries:
                        info = entry.stat(follow_symlinks=False)
                        require(not stat.S_ISLNK(info.st_mode), 'campaign symlink')
                        if stat.S_ISDIR(info.st_mode):
                            pending.append(entry.path)
                        elif stat.S_ISREG(info.st_mode):
                            total += info.st_size
            return total
        except FileNotFoundError:
            if attempt:
                raise


def stage_complete(stage):
    """Only a verified zero-exit stage may earn completion credit."""
    if not stage:
        return False
    require(stage.get('status') in ('complete','incomplete','running'), 'unknown stage status')
    if stage['status']!='complete':
        return False
    require(type(stage.get('exit_code')) is int and stage['exit_code']==0 and bool(stage.get('manifest_sha256')), 'complete stage requires verified zero exit')
    return True


def write_analysis(dest, state, profiles, units, plan, analysis_name, report_name):
    verify_analysis_sources(dest,state['frozen'])
    binary_hash=state['frozen']['binary']['sha256']
    for unit in state['units']:
        for stage in unit['stages'].values():stage_complete(stage)
        if unit.get('worker_identical'):
            require(all(stage_complete(unit['stages'].get(name,{})) for name in ('run','workers','workers-replay')), 'worker verification stage missing')
    analyses=[]
    for p in profiles:
        outcomes=[];replayed_games=0;worker_games=0
        for unit in state['units']:
            if unit['population']==p['population']:
                f=dest/(unit['id']+'-run')/'outcomes.json'
                stage=unit['stages'].get('run',{})
                if f.exists() and stage.get('manifest_sha256'):
                    inspect_manifest(f.parent,stage['exit_code'],binary_hash,next(u['experiment'] for u in units if u['id']==unit['id']))
                    require(digest((f.parent/'manifest.json').read_bytes())==stage['manifest_sha256'],'run manifest changed')
                    rows=load(f)
                    shard=next(u['experiment'] for u in units if u['id']==unit['id'])
                    rows=verified_analysis_rows(shard,rows,stage['status']=='complete')
                    outcomes.extend(rows)
                    if stage_complete(unit['stages'].get('replay',{})):
                        replayed_games+=sum(r.get('finalized_games',0) for r in rows)
                    if unit.get('worker_identical'):worker_games+=sum(r.get('finalized_games',0) for r in rows)
                elif stage:
                    # A dispatched process without verified artifacts has unknown
                    # progress; it is never relabeled as never started.
                    shard=next(u['experiment'] for u in units if u['id']==unit['id'])
                    outcomes.extend(verified_analysis_rows(shard,[],False))
        values,statuses=paired_blocks(p,outcomes)
        interval=cluster_interval(values,str(plan['analysis_seed'])+'/'+str(p['population']),plan['resamples'])
        planned_matches=len(p['block_ids'])*len(p['rotations'])*2
        planned_games=planned_matches*p['games_per_match']
        started_games=sum(r.get('started_games',0) for r in outcomes)
        analyses.append({'population':p['population'],'planned_blocks':len(p['block_ids']),'planned_matches':planned_matches,'planned_games':planned_games,'started_games':started_games,'started_instances':sum(r.get('started_instances',r.get('started_games',0)) for r in outcomes),'voided_games':sum(r.get('voided_games',0) for r in outcomes),'unknown_game_slots':sum(r.get('unknown_game_slots',0) for r in outcomes),'never_started_games':planned_games-started_games-sum(r.get('unknown_game_slots',0) for r in outcomes),'board_ended_games':sum(r.get('board_ended_games',0) for r in outcomes),'finalized_games':sum(r.get('finalized_games',0) for r in outcomes),'finalized_matches':sum(bool(r['match_complete']) for r in outcomes),'replayed_finalized_games':replayed_games,'worker_checked_finalized_games':worker_games,'unit_statuses':statuses,**interval})
    immutable_json(dest/analysis_name,analyses)
    report=['# Complete-game campaign','',f'Status: {state["status"]}.','', 'Observer: public. Estimates condition on complete paired seed blocks; incomplete selection may bias them. No strength or balance conclusion follows automatically.','']
    for a in analyses:
        report.append(f"Population {a['population']} games: planned {a['planned_games']}, started {a['started_games']}, board-ended {a['board_ended_games']}, financially finalized {a['finalized_games']}, never-started {a['never_started_games']}, unknown game slots {a['unknown_game_slots']}; physical instances {a['started_instances']}, voided instances {a['voided_games']}; finalized matches {a['finalized_matches']}/{a['planned_matches']}; replayed finalized games {a['replayed_finalized_games']}; worker-checked finalized games {a['worker_checked_finalized_games']}.")
        report.append(f"Population {a['population']}: {a['complete_blocks']} / {a['planned_blocks']} complete blocks; mean {a['mean']}; clustered 95% interval [{a['lower_95']}, {a['upper_95']}]. Planned-unit statuses: {a['unit_statuses']}.")
    with open(dest/report_name,'x') as out:out.write('\n'.join(report)+'\n')
    return analyses


def analyze_only(dest, state, frozen, plan, profiles, units):
    require(state['frozen']==frozen and state['plan_hash']==digest(encoded(plan)), 'analysis inputs changed')
    require(not state.get('invocation_active') and state['status']!='running', 'analysis requires inactive campaign')
    verify_state_plan(state,units)
    verify_shard_inputs(dest,units)
    require(load(dest/'plan.json')==plan, 'saved plan changed')
    verify_analysis_sources(dest,frozen)
    # Verify every published stage, including replay and worker checks, before
    # crediting any saved completion flag. No process or gameplay is executed.
    for unit in state['units']:
        spec=next(u['experiment'] for u in units if u['id']==unit['id'])
        for name,stage in unit['stages'].items():
            require(name in ('run','replay','workers','workers-replay'), 'unknown stage')
            require(stage.get('status') in ('complete','incomplete'), 'analysis requires closed stage')
            stage_complete(stage)
            if stage.get('manifest_sha256'):
                directory=dest/(unit['id']+'-'+name)
                inspect_manifest(directory,stage['exit_code'],frozen['binary']['sha256'],spec,plan['worker_check_workers'] if name.startswith('workers') else None)
                require(digest((directory/'manifest.json').read_bytes())==stage['manifest_sha256'], 'saved manifest changed')
            else:
                require(stage['status']!='complete', 'complete stage without verified manifest')
        if unit.get('worker_identical'):
            require(all(stage_complete(unit['stages'].get(name,{})) for name in ('run','workers','workers-replay')), 'worker verification stage missing')
            run=dest/(unit['id']+'-run');other=dest/(unit['id']+'-workers')
            for entry in load(run/'manifest.json')['artifacts']:
                rel=entry['path']
                if rel=='outcomes.json' or (rel.startswith('games/') and rel.endswith('/gameplay.json')):
                    require((run/rel).read_bytes()==(other/rel).read_bytes(), 'worker-dependent gameplay')
    # Separate immutable correction namespace; original analysis, journals,
    # elapsed budgets and publication pointer remain byte-for-byte unchanged.
    revision=1
    while any(dest.glob(f'reanalysis-{state["revision"]:06d}-{revision:04d}-*')):
        revision+=1
    prefix=f'reanalysis-{state["revision"]:06d}-{revision:04d}'
    analyses=write_analysis(dest,state,profiles,units,plan,prefix+'-analysis.json',prefix+'-report.md')
    immutable_json(dest/(prefix+'-provenance.json'), {'schema':'cgms-reanalysis-v1','campaign_revision':state['revision'],'campaign_state_sha256':digest(encoded(state)),'plan_sha256':state['plan_hash'],'evaluator_sha256':digest(Path(__file__).read_bytes()),'analysis_sha256':digest(encoded(analyses)),'analysis':ANALYSIS,'correction':'canonical-rational-score-decoding-v1','gameplay_rerun':False})
    print('Analysis-only correction published; protected artifacts:',dest/prefix)
    return 0


def main():
    with ExitStack() as stack:
        return run_campaign(stack)


def run_campaign(stack):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('binary'); parser.add_argument('plan'); parser.add_argument('output_id')
    mode=parser.add_mutually_exclusive_group()
    mode.add_argument('--resume', action='store_true')
    mode.add_argument('--analyze-only', action='store_true', help='verify saved artifacts and publish immutable corrected analysis; never run games or alter campaign state')
    args = parser.parse_args()
    os.umask(0o077)
    require(re.fullmatch(r'[A-Za-z0-9][A-Za-z0-9_-]{0,79}', args.output_id), 'invalid output id')
    binary = Path(args.binary).resolve(strict=True); plan_path = Path(args.plan).resolve(strict=True)
    plan = load(plan_path); frozen = frozen_inputs(binary, plan_path, plan)
    profiles = [load(path) for path in plan['profiles']]; units = build_units(profiles, plan['shard_blocks'])
    selected_replays = replay_selection(plan, units)
    require(set(plan['worker_check_units']).issubset({u['id'] for u in units}), 'unplanned worker check')
    dest = safe_path(ARTIFACTS/args.output_id)
    if args.analyze_only:
        lock=stack.enter_context(open(dest/'campaign.lock','rb'))
        fcntl.flock(lock,fcntl.LOCK_EX | fcntl.LOCK_NB)
        journals=sorted(dest.glob('state-[0-9]*.json'))
        state=load(journals[-1] if journals else dest/'state.json')
        return analyze_only(dest,state,frozen,plan,profiles,units)
    if args.resume:
        lock = stack.enter_context(open(dest/'campaign.lock', 'a+b'))
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        journals=sorted(dest.glob('state-[0-9]*.json'))
        state = load(journals[-1] if journals else dest/'state.json')
        require(state['frozen'] == frozen and state['plan_hash'] == digest(encoded(plan)), 'resume inputs changed')
        for row in state['units']:
            for name, stage in row['stages'].items():
                if stage['status'] == 'running':
                    output=dest/(row['id']+'-'+name)
                    try:
                        m=load(output/'manifest.json');inspect_manifest(output,m['exit_code'],frozen['binary']['sha256'],next(u['experiment'] for u in units if u['id']==row['id']),plan['worker_check_workers'] if name.startswith('workers') else None)
                        stage.update(status='complete' if m['exit_code']==0 else 'incomplete',exit_code=m['exit_code'],reason='recovered-published-manifest',manifest_sha256=digest((output/'manifest.json').read_bytes()))
                    except (OSError,ValueError,KeyError):
                        stage.update(status='incomplete',reason='interrupted-unpublished-output')
    else:
        dest.mkdir(parents=True, exist_ok=False)
        lock = stack.enter_context(open(dest/'campaign.lock', 'a+b'))
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        immutable_json(dest/'plan.json', plan)
        retain_analysis_sources(dest,frozen)
        state = {'schema': plan['schema'], 'plan_hash': digest(encoded(plan)), 'frozen': frozen, 'status': 'planned', 'elapsed_seconds': 0, 'revision': 0, 'units': [{'id': u['id'], 'population': u['population'], 'planned_matches': u['planned_matches'], 'planned_games': u['planned_games'], 'stages': {}} for u in units]}
        for u in units:
            immutable_json(dest/(u['id']+'-experiment.json'), u['experiment'])
        save_state(dest,state)
    verify_state_plan(state,units)
    verify_analysis_sources(dest,frozen)
    verify_shard_inputs(dest,units)
    interrupted = False
    def cancel(_signum, _frame):
        nonlocal interrupted
        interrupted = True
    signal.signal(signal.SIGINT, cancel); signal.signal(signal.SIGTERM, cancel)
    start_wall_accounting(state)
    started = time.monotonic(); previous_elapsed = state['elapsed_seconds']; binary_hash = frozen['binary']['sha256']
    def limit():
        if interrupted: return 'canceled'
        if previous_elapsed + time.monotonic()-started >= plan['wall_seconds']: return 'wall-budget'
        try:
            if storage_bytes(dest) >= plan['storage_bytes']: return 'storage-budget'
        except (OSError, ValueError):
            return 'storage-scan-failed'
        return ''
    def invoke(unit, name, argv):
        verify_shard_inputs(dest,[next(u for u in units if u['id']==unit['id'])])
        stage = unit['stages'].get(name)
        output = dest/(unit['id']+'-'+name)
        if stage:
            require(stage['status'] == 'complete', 'failed or interrupted stage retained; never silently rerun')
            m=inspect_manifest(output,0,binary_hash,next(u['experiment'] for u in units if u['id']==unit['id']),plan['worker_check_workers'] if name.startswith('workers') else None)
            require(stage['manifest_sha256']==digest((output/'manifest.json').read_bytes()),'completed manifest changed')
            return m
        require(not output.exists(), 'refuse to overwrite orphan output')
        verify_frozen_files(frozen)
        verify_analysis_sources(dest,frozen)
        spec_path = dest/(unit['id']+'-'+name+'-spec.json')
        immutable_json(spec_path, {'argv':[str(binary),*argv,'--review-hold','--out',str(output)],'cwd':str(ROOT),'capture':str(dest/(unit['id']+'-'+name+'-private'))})
        unit['stages'][name]={'status':'running'};save_state(dest,state)
        child=subprocess.Popen([sys.executable,str(ROOT/'xops/sim_private.py'),str(spec_path)],cwd=ROOT,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL,start_new_session=True)
        reason='';sent=None
        while child.poll() is None:
            reason=reason or limit()
            if reason and sent is None: child.send_signal(signal.SIGINT);sent=time.monotonic()
            if sent is not None and time.monotonic()-sent>60: os.killpg(child.pid,signal.SIGKILL);reason='cancellation-timeout'
            time.sleep(.2)
        code=child.returncode;stage=unit['stages'][name];stage.update(exit_code=code,status='complete' if code==0 else 'incomplete',reason=reason or ('process-failure' if code else ''))
        try:
            m=inspect_manifest(output,code,binary_hash,next(u['experiment'] for u in units if u['id']==unit['id']),plan['worker_check_workers'] if name.startswith('workers') else None);stage['manifest_sha256']=digest((output/'manifest.json').read_bytes())
        except (ValueError,OSError,KeyError):
            stage['status']='incomplete';stage['reason']=reason if reason=='storage-scan-failed' else 'manifest-verification-failed';m=None
        if m is not None and name in ('run','workers'):
            try:
                profile=load(dest/(unit['id']+'-experiment.json'))
                require(m['planned']==unit['planned_matches'],'manifest planned denominator')
                validate_outcomes(profile,load(output/'outcomes.json'),code==0)
            except (OSError,ValueError,KeyError):
                stage['status']='incomplete';stage['reason']='outcome-verification-failed'
        save_state(dest,state)
        require(stage['status']=='complete','campaign stage incomplete; remaining units preserved')
        return m
    try:
        state['status']='running';save_state(dest,state)
        for spec,unit in zip(units,state['units']):
            if any(stage['status']=='incomplete' for stage in unit['stages'].values()):
                # Explicit resume proceeds in the frozen order without rerunning or
                # deleting failed units; they remain in every planned denominator.
                continue
            why=limit()
            if why: state['status']=why;break
            run=dest/(unit['id']+'-run')
            argv=['tournament','--seed',spec['experiment']['root_seed'],'--experiment',str(dest/(unit['id']+'-experiment.json')),'--config',frozen['config']['path']]
            invoke(unit,'run',argv)
            if unit['id'] in selected_replays:
                invoke(unit,'replay',['replay','--manifest',str(run/'manifest.json')])
            if unit['id'] in plan['worker_check_units']:
                invoke(unit,'workers',[*argv,'--workers',str(plan['worker_check_workers'])])
                other=dest/(unit['id']+'-workers')
                a=load(run/'manifest.json');b=load(other/'manifest.json')
                for entry in a['artifacts']:
                    rel=entry['path']
                    if rel.startswith('games/') and rel.endswith('/gameplay.json') or rel=='outcomes.json':
                        require((run/rel).read_bytes()==(other/rel).read_bytes(),'worker-dependent gameplay')
                invoke(unit,'workers-replay',['replay','--manifest',str(other/'manifest.json')])
                unit['worker_identical']=True;save_state(dest,state)
        else: state['status']='incomplete' if any(any(stage['status']=='incomplete' for stage in unit['stages'].values()) for unit in state['units']) else 'complete'
    except (ValueError,OSError,KeyError):
        state['status']='incomplete'
    finally:
        checkpoint_wall_accounting(state)
        state['elapsed_seconds']=max(state['elapsed_seconds'],previous_elapsed+time.monotonic()-started)
        state['invocation_active']=False
        try:
            state['storage_bytes']=storage_bytes(dest)
            state.pop('storage_error',None)
        except (OSError, ValueError):
            state['storage_bytes']=None
            state['storage_error']='unavailable'
            if state['status']=='complete': state['status']='storage-scan-failed'
        save_state(dest,state)
        write_analysis(dest,state,profiles,units,plan,f'analysis-{state["revision"]:06d}.json',f'report-{state["revision"]:06d}.md')
    print('Campaign status:',state['status'],'; protected artifacts:',dest)
    return 0 if state['status']=='complete' else 130 if interrupted else 5


if __name__=='__main__':
    try: sys.exit(main())
    except (ValueError,OSError,KeyError):
        print('Campaign input or resume validation failed; private details withheld.',file=sys.stderr)
        sys.exit(2)
