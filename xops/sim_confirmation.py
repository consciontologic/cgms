#!/usr/bin/env python3
"""Authenticate frozen campaigns and publish fixed-family confirmation claims.

Run with --scope primary|roster --mixed CAMPAIGN [--economic CAMPAIGN
--pressure CAMPAIGN] --out NEW_PROTECTED_DIRECTORY. Sources must be frozen in
schema-v3, including this wrapper. No simulator or replay is executed here.
"""
import argparse
from fractions import Fraction
import os
from pathlib import Path
import sys

import sim_claims as claims
import sim_gameplay_evaluate as campaign


def validate_roster(profile, lineup):
    campaign.require(profile['games_per_match']==3,'confirmation requires three-game matches')
    campaign.require(profile.get('rules_profile')=='accepted-game-rules','accepted rules required')
    campaign.require(profile.get('output',{}).get('omniscient_comparison') is False,'omniscient comparisons forbidden')
    pop=profile['population']
    expected={f'p{i}':('economic' if i%2==0 else 'pressure') if lineup=='mixed' else lineup for i in range(pop)}
    bots={p['participant_id']:p for p in profile['bots']}
    campaign.require(len(bots)==len(profile['bots']) and set(bots)==set(expected),'fixed roster identities')
    for participant, policy in expected.items():
        bot=bots[participant]
        campaign.require(bot.get('information')=='seat-projection','baseline observation boundary')
        campaign.require(bot['policy']==policy and bot['version']=='v2' and not bot.get('features') and not bot.get('weights'),'fixed corrected baseline roster/parameters')
    candidate=profile['candidate']
    campaign.require(candidate.get('information')=='seat-projection','candidate observation boundary')
    campaign.require(candidate['participant_id']=='p0' and candidate['policy']=='opportunity' and candidate['version']=='v2' and not candidate.get('features') and not candidate.get('weights'),'fixed corrected opportunity candidate')
    seats=[f'p{i}' for i in range(pop)]
    cyclic={tuple(seats[i:]+seats[:i]) for i in range(pop)}
    campaign.require({tuple(r['seats']) for r in profile['rotations']}==cyclic and len(profile['rotations'])==pop,'fixed cyclic schedule')


def population_data(profile, rows, unknown_blocks=()):
    unknown=set(unknown_blocks)
    verified=dict(profile,block_ids=[b for b in profile['block_ids'] if b not in unknown])
    usable=[r for r in rows if r['block'] not in unknown]
    completion=claims.completion_by_treatment(verified,usable) if verified['block_ids'] else {
        t:dict(planned_matches=0,started_matches=0,finalized_matches=0,incomplete_started_matches=0,
               never_started_matches=0,planned_games=0,started_games=0,board_ended_games=0,
               finalized_games=0,never_started_games=0,statuses={}) for t in ('baseline','candidate')}
    unknown_matches=len(unknown)*len(profile['rotations'])
    for value in completion.values():
        value['planned_matches']+=unknown_matches
        value['planned_games']+=unknown_matches*profile['games_per_match']
        value['unknown_matches']=unknown_matches
        value['unknown_game_slots']=unknown_matches*profile['games_per_match']
    values=claims.strength_blocks(verified,usable) if verified['block_ids'] else []
    return {'profile':profile,'verified_profile':verified,'rows':usable,
            'strength_values':values,'completion':completion,
            'all_complete':not unknown and len(values)==len(profile['block_ids'])}


def read_campaign(directory):
    directory=campaign.safe_path(directory)
    journals=sorted(directory.glob('state-[0-9]*.json'))
    campaign.require(bool(journals),'missing immutable campaign journal')
    state=campaign.load(journals[-1]);plan=campaign.load(directory/'plan.json')
    campaign.require(plan['schema']==campaign.SCHEMA_V3 and state['schema']==campaign.SCHEMA_V3,'confirmation requires source-frozen schema-v3')
    campaign.require(not state.get('invocation_active') and state['status']!='running','campaign still active')
    frozen=campaign.frozen_inputs(Path(state['frozen']['binary']['path']),Path(state['frozen']['plan']['path']),plan)
    campaign.require(frozen==state['frozen'] and state['plan_hash']==campaign.digest(campaign.encoded(plan)),'frozen campaign inputs changed')
    campaign.verify_frozen_files(frozen);campaign.verify_analysis_sources(directory,frozen)
    campaign.require(str(Path(__file__).resolve()) in [f['path'] for k,f in frozen.items() if k.startswith('analysis-source-')],'confirmation wrapper was not frozen')
    profiles=[campaign.load(p) for p in plan['profiles']]
    units=campaign.build_units(profiles,plan['shard_blocks'])
    campaign.replay_selection(plan,units)
    campaign.verify_state_plan(state,units);campaign.verify_shard_inputs(directory,units)
    rows={p['population']:[] for p in profiles};unknown={p['population']:set() for p in profiles}
    verified_replays={p['population']:0 for p in profiles}
    worker_games={p['population']:0 for p in profiles}
    for unit,saved in zip(units,state['units']):
        for name,stage in saved['stages'].items():
            campaign.require(name in ('run','replay','workers','workers-replay'),'unknown stage')
            campaign.require(stage['status'] in ('complete','incomplete'),'unclosed stage')
            campaign.stage_complete(stage)
            if stage.get('manifest_sha256'):
                output=directory/(unit['id']+'-'+name)
                campaign.inspect_manifest(output,stage['exit_code'],frozen['binary']['sha256'],unit['experiment'],plan['worker_check_workers'] if name.startswith('workers') else None)
                campaign.require(campaign.digest((output/'manifest.json').read_bytes())==stage['manifest_sha256'],'manifest changed')
        run=saved['stages'].get('run')
        if not run:continue
        if not run.get('manifest_sha256'):
            unknown[unit['population']].update(unit['experiment']['block_ids']);continue
        outcome=campaign.load(directory/(unit['id']+'-run')/'outcomes.json')
        campaign.validate_outcomes(unit['experiment'],outcome,run['status']=='complete')
        campaign.paired_blocks(unit['experiment'],outcome)
        rows[unit['population']].extend(outcome)
        if campaign.stage_complete(saved['stages'].get('replay',{})):
            verified_replays[unit['population']]+=sum(r['finalized_games'] for r in outcome)
        if saved.get('worker_identical'):
            campaign.require(plan['worker_check_workers']!=unit['experiment']['budgets'].get('workers'),'worker verification requires distinct worker counts')
            campaign.require(all(campaign.stage_complete(saved['stages'].get(n,{})) for n in ('run','workers','workers-replay')),'incomplete worker verification')
            runpath=directory/(unit['id']+'-run');other=directory/(unit['id']+'-workers')
            for artifact in campaign.load(runpath/'manifest.json')['artifacts']:
                rel=artifact['path']
                if rel=='outcomes.json' or rel.startswith('games/') and rel.endswith('/gameplay.json'):
                    campaign.require((runpath/rel).read_bytes()==(other/rel).read_bytes(),'worker-dependent outcomes')
            worker_games[unit['population']]+=sum(r['finalized_games'] for r in outcome)
    return {'provenance':{k:v['sha256'] for k,v in frozen.items()},
            'populations':[dict(population_data(p,rows[p['population']],unknown[p['population']]),
                                replayed_finalized_games=verified_replays[p['population']],
                                worker_checked_finalized_games=worker_games[p['population']]) for p in profiles]}


def labels(interval, admissible, margin=5):
    result={key:'inconclusive' for key in ('positive_effect','practical_strength','noninferiority','equivalence')}
    lo,hi=interval['lower'],interval['upper']
    if not admissible or interval['independent_blocks']<1000 or lo is None or hi is None:return result
    shared=campaign.classify_interval(interval,5,margin)
    result['practical_strength']=shared['strength']
    result['equivalence']=shared['equivalence']
    if lo>0:result['positive_effect']='supported'
    elif hi<0:result['positive_effect']='refuted'
    # The explicit claim is an improvement greater than five, not merely
    # non-negative play. A confidently smaller positive effect refutes it.
    if hi<5:result['practical_strength']='refuted'
    if lo>-5:result['noninferiority']='supported'
    elif hi<-5:result['noninferiority']='refuted'
    return result


def aggregate(scope, cohorts):
    expected={'mixed'} if scope=='primary' else {'mixed','economic','pressure'}
    campaign.require(scope in ('primary','roster') and set(cohorts)==expected,'fixed confirmation scope/lineups')
    family=2 if scope=='primary' else 42
    result={'schema':'cgms-confirmation-report-v1','scope':scope,'family_comparisons':family,
            'confidence':'Bonferroni simultaneous Student; block normality exact, otherwise asymptotic',
            'contrasts':[],'completion':[],'limits':['Fixed cyclic orders only','Completion-only estimates conditional','No human-strength or enjoyment claim']}
    verification={pop:{'replayed_finalized_games':sum(d.get('replayed_finalized_games',0) for data in cohorts.values() for d in data if d['profile']['population']==pop),
                       'worker_checked_finalized_games':sum(d.get('worker_checked_finalized_games',0) for data in cohorts.values() for d in data if d['profile']['population']==pop)} for pop in (3,4)}
    result['verification']=verification
    reference={}
    for lineup,populations in cohorts.items():
        campaign.require(len(populations)==2 and {d['profile']['population'] for d in populations}=={3,4},'separate complete population plans required')
        campaign.require(len({d['profile']['root_seed'] for d in populations})==2,'population roots must be distinct')
        for data in populations:
            p=data['profile'];pop=p['population'];validate_roster(p,lineup)
            campaign.require(p.get('id')==f'confirmation-v3-{scope}-{lineup}-{pop}p','scope/lineup must match frozen profile identity')
            signature={k:p.get(k) for k in ('root_seed','pairing_id','block_ids','rotations','games_per_match','rules_profile','engine_version')}
            signature['budgets']={k:v for k,v in p['budgets'].items() if k not in ('max_matches','workers')}
            campaign.require(pop not in reference or reference[pop]==signature,'cohort chance/schedule/rules/budgets mismatch')
            reference[pop]=signature
            admissible=(data['all_complete'] and p['phase']=='held-out' and
                        verification[pop]['replayed_finalized_games']>=1000 and
                        verification[pop]['worker_checked_finalized_games']>0)
            interval=campaign.prospective_interval(data['strength_values'],family)
            result['contrasts'].append({'population':pop,'lineup':lineup,'kind':'strength','unit':'points/game',
                'interval':interval,'labels':labels(interval,admissible)})
            result['completion'].append({'population':pop,'lineup':lineup,'planned_blocks':len(p['block_ids']),
                'complete_paired_blocks':len(data['strength_values']),'treatments':data['completion'],
                'replayed_finalized_games':data.get('replayed_finalized_games',0)})
            if lineup!='mixed':
                vp=data['verified_profile']
                vectors=claims.seat_blocks(vp,data['rows']) if vp['block_ids'] else {
                    metric:{f'{i}-{j}':[] for i in range(pop) for j in range(i+1,pop)} for metric in ('score_per_game','win_share')}
                for metric,pairs in vectors.items():
                    for pair,values in pairs.items():
                        interval=campaign.prospective_interval(values,family)
                        result['contrasts'].append({'population':pop,'lineup':lineup,'kind':'seat','seats':pair,
                            'unit':metric,'interval':interval,'labels':{'equivalence':labels(interval,admissible,5 if metric=='score_per_game' else .05)['equivalence']}})
    campaign.require(len(result['contrasts'])==family,'comparison family size differs from freeze')
    return result


def markdown(report):
    lines=['# Frozen confirmation analysis','',f"Scope: {report['scope']}; simultaneous family K={report['family_comparisons']}.",
           'Units are independent seed blocks. Results remain conditional when completion is incomplete.',
           'Claims require at least 1,000 blocks, every planned block complete, at least 1,000 completed-game replays per population and worker verification.','',
           '| Population | Lineup | Contrast | Block count | Mean | Simultaneous interval | Labels |',
           '|---|---|---|---:|---:|---|---|']
    for row in report['contrasts']:
        i=row['interval'];desc=row['kind']+(' '+row['seats'] if 'seats' in row else '')+' '+row['unit']
        lines.append(f"| {row['population']} | {row['lineup']} | {desc} | {i['independent_blocks']} | {i['mean']} | [{i['lower']}, {i['upper']}] | "+'; '.join(f'{k}: {v}' for k,v in row['labels'].items())+' |')
    lines+=['','## Completion','']
    for row in report['completion']:
        lines.append(f"{row['population']} players, {row['lineup']}: {row['complete_paired_blocks']}/{row['planned_blocks']} complete blocks; {row['replayed_finalized_games']} completed-game replays.")
        for treatment,c in row['treatments'].items():
            lines.append(f"{treatment}: games planned {c['planned_games']}, started {c['started_games']}, finalized {c['finalized_games']}, never started {c['never_started_games']}, unknown {c['unknown_game_slots']}; complete matches {c['finalized_matches']}/{c['planned_matches']}, incomplete started {c['incomplete_started_matches']}; statuses {c['statuses']}.")
    return '\n'.join(lines)+'\n'


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--scope',choices=['primary','roster'],required=True)
    for name in ('mixed','economic','pressure'):parser.add_argument('--'+name,type=Path)
    parser.add_argument('--out',type=Path,required=True);args=parser.parse_args()
    os.umask(0o077)
    loaded={k:read_campaign(getattr(args,k)) for k in ('mixed','economic','pressure') if getattr(args,k)}
    signatures=[{k:v for k,v in x['provenance'].items() if k in ('binary','config','source_manifest') or k.startswith('analysis-source-')} for x in loaded.values()]
    campaign.require(bool(signatures) and all(v==signatures[0] for v in signatures),'cohort implementation provenance mismatch')
    report=aggregate(args.scope,{k:v['populations'] for k,v in loaded.items()})
    report['provenance']={k:v['provenance'] for k,v in loaded.items()}
    dest=campaign.safe_path(args.out);dest.mkdir(parents=True,exist_ok=False)
    campaign.immutable_json(dest/'report.json',report)
    with open(dest/'report.md','x') as out:out.write(markdown(report))
    print('Verified aggregate report published in restricted output directory.')
    return 0


if __name__=='__main__':
    try:sys.exit(main())
    except (ValueError,OSError,KeyError,TypeError):
        print('Confirmation validation failed; private inputs withheld.',file=sys.stderr);sys.exit(2)
