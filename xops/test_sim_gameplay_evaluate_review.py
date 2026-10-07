import importlib.util
from pathlib import Path
import unittest
import copy
import tempfile

spec=importlib.util.spec_from_file_location('campaign_review',Path(__file__).with_name('sim_gameplay_evaluate.py'))
campaign=importlib.util.module_from_spec(spec);spec.loader.exec_module(campaign)


def profile(pop, blocks=1000):
    seats=['p'+str(i) for i in range(pop)]
    return {'population':pop,'block_ids':['held-'+str(i) for i in range(blocks)],'rotations':[{'id':'r'+str(i),'seats':seats[i:]+seats[:i]} for i in range(pop)],'games_per_match':2,'candidate':{'participant_id':'p0'},'budgets':{'max_matches':100000},'root_seed':'private-frozen-seed'}


class IndependentCampaignReview(unittest.TestCase):
    def test_balanced_resource_order_and_exact_thousand_block_denominators(self):
        units=campaign.build_units([profile(4),profile(3)],5)
        self.assertEqual([u['id'] for u in units[:4]],['p3-s0000','p4-s0000','p3-s0001','p4-s0001'])
        for pop in (3,4):
            selected=[u for u in units if u['population']==pop]
            self.assertEqual(len(selected),200)
            self.assertEqual(sum(u['planned_matches'] for u in selected),1000*pop*2)
            self.assertEqual(sum(u['planned_games'] for u in selected),1000*pop*2*2)
            self.assertEqual([b for u in selected for b in u['experiment']['block_ids']],profile(pop)['block_ids'])

    def test_derived_shard_seed_tampering_rejected(self):
        unit=campaign.build_units([profile(3,1)],1)[0]
        with tempfile.TemporaryDirectory() as d:
            path=Path(d)/(unit['id']+'-experiment.json')
            campaign.immutable_json(path,unit['experiment'])
            campaign.verify_shard_inputs(Path(d),[unit])
            changed=copy.deepcopy(unit['experiment']);changed['root_seed']='different-seed'
            path.write_bytes(campaign.encoded(changed))
            with self.assertRaises(ValueError):campaign.verify_shard_inputs(Path(d),[unit])

    def test_analysis_rejects_unplanned_rotation_and_treatment(self):
        p=profile(3,1)
        for rotation,treatment in [('unplanned','candidate'),('r0','cherry-picked')]:
            row={'block':p['block_ids'][0],'rotation':rotation,'treatment':treatment,'status':'complete','match_complete':False}
            with self.assertRaises(ValueError):campaign.paired_blocks(p,[row])

    def test_incomplete_whole_block_keeps_all_thousand_blocks_planned(self):
        p=profile(3)
        values,statuses=campaign.paired_blocks(p,[])
        self.assertEqual(values,[])
        self.assertEqual(statuses,{'not-started':6000})

    def test_one_block_cannot_supply_precision(self):
        interval=campaign.cluster_interval(['9'],'frozen-analysis',100)
        self.assertEqual(interval['mean'],'9')
        self.assertIsNone(interval['lower_95'])
        self.assertIsNone(interval['upper_95'])

    def test_invalid_outcomes_retained_as_unknown_without_scores(self):
        p=profile(3,1)
        rows=campaign.verified_analysis_rows(p,[{'block':'held-0','rotation':'r0','treatment':'candidate','scores':['999999'],'population':3,'planned_game_slots':2,'finalized_games':100,'match_complete':True}],True)
        self.assertEqual(len(rows),6)
        self.assertEqual(sum(r['unknown_game_slots'] for r in rows),12)
        self.assertTrue(all('scores' not in r for r in rows))
        values,statuses=campaign.paired_blocks(p,rows)
        self.assertEqual(values,[])
        self.assertEqual(statuses,{'unverified-outcomes':6})

    def test_duplicate_seat_schedules_rejected(self):
        p=profile(3,1)
        for r in p['rotations']:r['seats']=['p0','p1','p2']
        with self.assertRaises(ValueError):campaign.build_units([p],1)

    def test_development_block_overlap_rejected(self):
        p=profile(3,1);p['development_block_ids']=['held-0']
        with self.assertRaises(ValueError):campaign.build_units([p],1)

    def test_resume_cannot_drop_or_reorder_planned_shards(self):
        units=campaign.build_units([profile(3,6),profile(4,6)],5)
        state={'units':copy.deepcopy(units)}
        campaign.verify_state_plan(state,units)
        state['units'].pop()
        with self.assertRaises(ValueError):campaign.verify_state_plan(state,units)
        state={'units':list(reversed(copy.deepcopy(units)))}
        with self.assertRaises(ValueError):campaign.verify_state_plan(state,units)

    def test_unclean_resume_charges_downtime_without_resetting_budget(self):
        state={'elapsed_seconds':7}
        campaign.start_wall_accounting(state,100)
        campaign.checkpoint_wall_accounting(state,110)
        self.assertEqual(state['elapsed_seconds'],17)
        # Crash at the journal, resume after thirty seconds including downtime.
        restored=copy.deepcopy(state)
        campaign.start_wall_accounting(restored,140)
        self.assertEqual(restored['elapsed_seconds'],47)
        campaign.checkpoint_wall_accounting(restored,145)
        self.assertEqual(restored['elapsed_seconds'],52)
        restored['invocation_active']=False
        campaign.start_wall_accounting(restored,200)
        self.assertEqual(restored['elapsed_seconds'],52)

    def test_effective_experiment_only_allows_named_worker_override(self):
        frozen=profile(3,1);frozen['budgets']['workers']=1
        with tempfile.TemporaryDirectory() as d:
            path=Path(d)/'effective-experiment.json'
            path.write_bytes(campaign.encoded(frozen))
            campaign.verify_effective_experiment(Path(d),frozen)
            changed=copy.deepcopy(frozen);changed['budgets']['workers']=4
            path.write_bytes(campaign.encoded(changed))
            campaign.verify_effective_experiment(Path(d),frozen,4)
            with self.assertRaises(ValueError):campaign.verify_effective_experiment(Path(d),frozen)
            for mutate in ('root_seed','candidate'):
                bad=copy.deepcopy(changed);bad[mutate]='different-frozen-value'
                path.write_bytes(campaign.encoded(bad))
                with self.assertRaises(ValueError):campaign.verify_effective_experiment(Path(d),frozen,4)

    def test_public_analysis_rejects_private_status_and_impossible_instances(self):
        p=profile(3,1)
        rows=[{'block':'held-0','rotation':r['id'],'treatment':t,'population':3,'planned_game_slots':2,'status':'budget-exhausted','match_complete':False,'started_games':1,'board_ended_games':0,'finalized_games':0,'started_instances':2,'voided_games':1} for r in p['rotations'] for t in ('baseline','candidate')]
        campaign.validate_outcomes(p,rows,False)
        for field,value in [('status','private-seed-canary'),('voided_games',3)]:
            bad=copy.deepcopy(rows);bad[0][field]=value
            with self.assertRaises(ValueError):campaign.validate_outcomes(p,bad,False)


class CanonicalScoreReview(unittest.TestCase):
    def test_actual_cli_rational_shape_is_exact_and_canonical(self):
        self.assertEqual(campaign.decode_score({'numerator':'-17','denominator':'3'}),campaign.Fraction(-17,3))
        self.assertEqual(campaign.decode_score({'numerator':'0','denominator':'1'}),0)
        for bad in ['1/2',1,True,{'numerator':1,'denominator':'2'}, {'numerator':'01','denominator':'2'}, {'numerator':'-0','denominator':'1'}, {'numerator':'2','denominator':'4'}, {'numerator':'0','denominator':'2'}, {'numerator':'1','denominator':'0'}, {'numerator':'1','denominator':'-2'}, {'numerator':'1','denominator':'2','extra':1}]:
            with self.subTest(bad=bad),self.assertRaises(ValueError):campaign.decode_score(bad)

    def test_actual_cli_rows_are_not_discarded_as_unverified(self):
        p=profile(3,1)
        rows=[]
        for r in p['rotations']:
            for treatment in ('baseline','candidate'):
                scores=[{'numerator':'0','denominator':'1'} for _ in range(3)]
                scores[r['seats'].index('p0')]={'numerator':'7' if treatment=='candidate' else '2','denominator':'3'}
                rows.append({'block':'held-0','rotation':r['id'],'treatment':treatment,'population':3,'planned_game_slots':2,'status':'complete','match_complete':True,'started_games':2,'board_ended_games':2,'finalized_games':2,'started_instances':2,'voided_games':0,'seats':r['seats'],'scores':scores})
        verified=campaign.verified_analysis_rows(p,rows,True)
        self.assertEqual(verified,rows)
        values,statuses=campaign.paired_blocks(p,verified)
        self.assertEqual(values,[campaign.Fraction(5,3)])
        self.assertEqual(statuses,{'complete':6})

class AnalysisOnlyReview(unittest.TestCase):
    def test_analysis_only_preserves_campaign_and_never_launches_process(self):
        from unittest.mock import patch
        with tempfile.TemporaryDirectory() as d:
            root=Path(d); art=root/'artifacts';dest=art/'test';dest.mkdir(parents=True)
            binary=root/'sim';binary.write_bytes(b'frozen never executed')
            config=root/'rules.json';config.write_text('{}')
            source=root/'source.json';source.write_text('{}')
            p=profile(3,1);pfile=root/'profile.json';pfile.write_bytes(campaign.encoded(p))
            plan={'schema':campaign.SCHEMA,'config':str(config),'profiles':[str(pfile)],'source_manifest':str(source),'shard_blocks':1,'wall_seconds':1,'storage_bytes':1000000,'worker_check_units':[],'worker_check_workers':2,'analysis_seed':'private-analysis','resamples':100}
            planfile=root/'plan.json';planfile.write_bytes(campaign.encoded(plan))
            units=campaign.build_units([p],1)
            state={'schema':campaign.SCHEMA,'revision':7,'frozen':campaign.frozen_inputs(binary,planfile,plan),'plan_hash':campaign.digest(campaign.encoded(plan)),'status':'wall-budget','elapsed_seconds':2,'invocation_active':False,'units':[dict({k:u[k] for k in ('id','population','planned_matches','planned_games')},stages={}) for u in units]}
            rows=[]
            for rotation in p['rotations']:
                for treatment in ('baseline','candidate'):
                    scores=[{'numerator':'0','denominator':'1'} for _ in range(3)]
                    scores[rotation['seats'].index('p0')]={'numerator':'7' if treatment=='candidate' else '2','denominator':'3'}
                    rows.append({'block':'held-0','rotation':rotation['id'],'treatment':treatment,'population':3,'planned_game_slots':2,'status':'complete','match_complete':True,'started_games':2,'board_ended_games':2,'finalized_games':2,'started_instances':2,'voided_games':0,'seats':rotation['seats'],'scores':scores})
            for stage in ('run','replay'):
                directory=dest/(units[0]['id']+'-'+stage);directory.mkdir()
                artifacts=[]
                for name,value in [('outcomes.json',rows),('effective-experiment.json',units[0]['experiment'])]:
                    raw=campaign.encoded(value);(directory/name).write_bytes(raw)
                    artifacts.append({'path':name,'bytes':len(raw),'sha256':campaign.digest(raw)})
                raw=b'public report';(directory/'report.md').write_bytes(raw)
                artifacts.append({'path':'report.md','bytes':len(raw),'sha256':campaign.digest(raw)})
                manifest={'schema_version':'cgms-manifest-v1','exit_code':0,'privacy':'restricted','provenance':{'binary_sha256':state['frozen']['binary']['sha256']},'artifacts':artifacts}
                campaign.immutable_json(directory/'manifest.json',manifest)
                state['units'][0]['stages'][stage]={'status':'complete','exit_code':0,'manifest_sha256':campaign.digest(campaign.encoded(manifest))}
            for name,value in [('state.json',state),('state-000007.json',state),('plan.json',plan),(units[0]['id']+'-experiment.json',units[0]['experiment']),('analysis-000007.json',[])]:campaign.immutable_json(dest/name,value)
            (dest/'report-000007.md').write_text('original preserved')
            (dest/'campaign.lock').touch()
            before={str(f.relative_to(dest)):f.read_bytes() for f in dest.rglob('*') if f.is_file()}
            argv=['campaign',str(binary),str(planfile),'test','--analyze-only']
            with patch.object(campaign,'ARTIFACTS',art),patch.object(campaign.sys,'argv',argv),patch.object(campaign.subprocess,'Popen',side_effect=AssertionError('analysis launched process')),patch.object(campaign,'save_state',side_effect=AssertionError('analysis changed state')):
                self.assertEqual(campaign.main(),0)
                self.assertEqual(campaign.main(),0)
            for name,data in before.items():self.assertEqual((dest/name).read_bytes(),data)
            corrected=sorted(dest.glob('reanalysis-*-analysis.json'))
            self.assertEqual(len(corrected),2)
            self.assertEqual(campaign.load(corrected[0])[0]['planned_games'],12)
            self.assertEqual(campaign.load(corrected[0])[0]['never_started_games'],0)
            self.assertEqual(campaign.load(corrected[0])[0]['finalized_games'],12)
            self.assertEqual(campaign.load(corrected[0])[0]['replayed_finalized_games'],12)
            self.assertEqual(campaign.load(corrected[0])[0]['mean'],'5/3')
            self.assertEqual(campaign.load(corrected[0])[0]['complete_blocks'],1)
            self.assertNotIn('private-analysis',''.join(f.read_text() for f in dest.glob('reanalysis-*-report.md')))
            # An authentic canceled manifest is not a successful replay merely
            # because a mutable journal flag claims complete.
            replay_manifest=dest/(units[0]['id']+'-replay')/'manifest.json'
            original_manifest=replay_manifest.read_bytes()
            canceled=campaign.load(replay_manifest);canceled['exit_code']=130
            replay_manifest.write_bytes(campaign.encoded(canceled))
            changed=copy.deepcopy(state)
            changed['units'][0]['stages']['replay'].update(exit_code=130,status='complete',manifest_sha256=campaign.digest(replay_manifest.read_bytes()))
            (dest/'state-000007.json').write_bytes(campaign.encoded(changed))
            with patch.object(campaign,'ARTIFACTS',art),patch.object(campaign.sys,'argv',argv):
                with self.assertRaises(ValueError):campaign.main()
            self.assertEqual(len(list(dest.glob('reanalysis-*-analysis.json'))),2)
            replay_manifest.write_bytes(original_manifest)
            (dest/'state-000007.json').write_bytes(campaign.encoded(state))
            replay_report=dest/(units[0]['id']+'-replay')/'report.md'
            replay_report.write_bytes(b'tampered replay')
            with patch.object(campaign,'ARTIFACTS',art),patch.object(campaign.sys,'argv',argv):
                with self.assertRaises(ValueError):campaign.main()
            self.assertEqual(len(list(dest.glob('reanalysis-*-analysis.json'))),2)
            replay_report.write_bytes(b'public report')
            binary.write_bytes(b'tamper')
            with patch.object(campaign,'ARTIFACTS',art),patch.object(campaign.sys,'argv',argv):
                with self.assertRaises(ValueError):campaign.main()


class StageCompletionReview(unittest.TestCase):
    def test_completion_requires_verified_integer_zero_exit(self):
        for stage in ('run','replay','workers','workers-replay'):
            for code in (130,5,-9,True,False,'0',None):
                with self.subTest(stage=stage,code=code),self.assertRaises(ValueError):
                    campaign.stage_complete({'status':'complete','exit_code':code,'manifest_sha256':'verified'})
        self.assertTrue(campaign.stage_complete({'status':'complete','exit_code':0,'manifest_sha256':'verified'}))
        self.assertFalse(campaign.stage_complete({'status':'incomplete','exit_code':130,'manifest_sha256':'verified'}))
        # A zero process exit can still have failed artifact verification.
        self.assertFalse(campaign.stage_complete({'status':'incomplete','exit_code':0}))


class ProspectiveInferenceReview(unittest.TestCase):
    def test_student_quantile_reference_vectors(self):
        for df, expected in [(1,12.7062047364),(2,4.3026527297),(10,2.2281388520),(999,1.9623414611)]:
            self.assertAlmostEqual(campaign.student_critical(.975,df),expected,places=7)

    def test_exact_block_mean_and_variance(self):
        result=campaign.prospective_interval(['1','3','5'], comparisons=1)
        self.assertEqual(result['mean'],'3')
        self.assertEqual(result['sample_variance'],'4')
        self.assertEqual(result['independent_blocks'],3)
        self.assertAlmostEqual(result['upper']-3,4.3026527297*(4/3)**.5,places=7)

    def test_multiplicity_widens_and_tiny_sample_inconclusive(self):
        a=campaign.prospective_interval([0,1,2],comparisons=1)
        b=campaign.prospective_interval([0,1,2],comparisons=20)
        self.assertGreater(b['upper'],a['upper'])
        self.assertEqual(campaign.classify_interval(a,5,5,min_blocks=1000)['strength'],'inconclusive')
        self.assertIsNone(campaign.prospective_interval([1],comparisons=1)['lower'])

    def test_equivalence_requires_interval_within_margin(self):
        r={'independent_blocks':1000,'mean':'0','lower':-6,'upper':6}
        self.assertEqual(campaign.classify_interval(r,5,5)['equivalence'],'inconclusive')
        r.update(lower=-4,upper=4)
        self.assertEqual(campaign.classify_interval(r,5,5)['equivalence'],'supported')
        r.update(lower=6,upper=8)
        labels=campaign.classify_interval(r,5,5)
        self.assertEqual(labels['equivalence'],'refuted')
        self.assertEqual(labels['strength'],'supported')
        r.update(lower=-8,upper=-6)
        self.assertEqual(campaign.classify_interval(r,5,5)['strength'],'refuted')

    def test_degenerate_and_invalid_inputs(self):
        r=campaign.prospective_interval([0]*1000,comparisons=1)
        self.assertEqual(campaign.classify_interval(r,5,5)['equivalence'],'inconclusive')
        for count in (0,-1,True):
            with self.assertRaises(ValueError):campaign.prospective_interval([1,2],comparisons=count)


class FrozenReplaySelectionReview(unittest.TestCase):
    def test_v1_keeps_all_and_v2_uses_frozen_ids(self):
        units=campaign.build_units([profile(3,6),profile(4,6)],2)
        self.assertEqual(campaign.replay_selection({'schema':campaign.SCHEMA},units),{u['id'] for u in units})
        plan={'schema':'cgms-gameplay-campaign-v2','replay_units':['p3-s0000','p4-s0000'],'worker_check_units':['p3-s0000']}
        self.assertEqual(campaign.replay_selection(plan,units),{'p3-s0000','p4-s0000'})
        plan['replay_units'].append('not-a-unit')
        with self.assertRaises(ValueError):campaign.replay_selection(plan,units)

    def test_selection_cannot_duplicate_or_omit_worker_verification(self):
        units=campaign.build_units([profile(3,1)],1)
        plan={'schema':'cgms-gameplay-campaign-v2','replay_units':[],'worker_check_units':['p3-s0000']}
        with self.assertRaises(ValueError):campaign.replay_selection(plan,units)
        plan['replay_units']=['p3-s0000','p3-s0000']
        with self.assertRaises(ValueError):campaign.replay_selection(plan,units)

    def test_v2_actual_dispatch_counts_only_selected_replays_and_freezes_resume(self):
        self.dispatch_case(campaign.SCHEMA_V2)

    def test_v3_analysis_sources_freeze_before_dispatch_and_resume(self):
        self.dispatch_case(campaign.SCHEMA_V3)
        self.dispatch_case(campaign.SCHEMA_V3,tamper_during_run=True)

    def dispatch_case(self, schema, tamper_during_run=False):
        from unittest.mock import patch
        with tempfile.TemporaryDirectory() as d:
            root=Path(d); art=root/'artifacts';art.mkdir()
            (root/'xops').mkdir()
            runner=root/'xops/sim_gameplay_evaluate.py';runner.write_text('runner-v1')
            claims=root/'xops/sim_claims.py';claims.write_text('claims-v1')
            binary=root/'sim';binary.write_bytes(b'fake frozen binary')
            config=root/'config';config.write_text('{}')
            source=root/'source';source.write_text('{}')
            p=profile(3,2);p['budgets']['workers']=4
            pfile=root/'profile';pfile.write_bytes(campaign.encoded(p))
            plan={'schema':schema,'config':str(config),'profiles':[str(pfile)],
                  'source_manifest':str(source),'shard_blocks':1,'wall_seconds':60,
                  'storage_bytes':10000000,'worker_check_units':['p3-s0000'],
                  'worker_check_workers':1,'analysis_seed':'test','resamples':100,
                  'replay_units':['p3-s0000']}
            if schema==campaign.SCHEMA_V3:
                plan['analysis_sources']=[str(runner),str(claims)]
            planfile=root/'plan';planfile.write_bytes(campaign.encoded(plan))
            calls=[]

            def child(argv, **kwargs):
                # Exercise protected spec creation, real manifest verification,
                # worker comparison, final analysis and dispatch state journals.
                spec=campaign.load(argv[-1]); args=spec['argv']
                out=Path(args[args.index('--out')+1]);out.mkdir()
                calls.append(out.name)
                if args[1]=='replay':
                    parent=Path(args[args.index('--manifest')+1]).parent
                    experiment=campaign.load(parent/'effective-experiment.json')
                    rows=campaign.load(parent/'outcomes.json')
                else:
                    experiment=campaign.load(args[args.index('--experiment')+1])
                    if '--workers' in args:
                        experiment['budgets']['workers']=int(args[args.index('--workers')+1])
                    rows=[{'block':b,'rotation':r['id'],'treatment':t,'population':3,
                           'planned_game_slots':2,'started_games':2,'board_ended_games':2,
                           'finalized_games':2,'status':'complete','match_complete':True,
                           'seats':r['seats'],'scores':[{'numerator':'0','denominator':'1'}]*3}
                          for b in experiment['block_ids'] for r in experiment['rotations']
                          for t in ('baseline','candidate')]
                artifacts=[]
                for name,raw in [('report.md',b'public'),('effective-experiment.json',campaign.encoded(experiment)),('outcomes.json',campaign.encoded(rows))]:
                    (out/name).write_bytes(raw)
                    artifacts.append({'path':name,'bytes':len(raw),'sha256':campaign.digest(raw)})
                campaign.immutable_json(out/'manifest.json',{'schema_version':'cgms-manifest-v1',
                    'exit_code':0,'privacy':'restricted','planned':len(rows),
                    'provenance':{'binary_sha256':campaign.digest(binary.read_bytes())},'artifacts':artifacts})
                if tamper_during_run:
                    claims.write_text('changed analysis after first process')
                class Done:
                    returncode=0
                    def poll(self):return 0
                return Done()

            argv=['campaign',str(binary),str(planfile),'test']
            with patch.object(campaign,'ROOT',root),patch.object(campaign,'__file__',str(runner)),patch.object(campaign,'ARTIFACTS',art),patch.object(campaign.sys,'argv',argv),patch.object(campaign.subprocess,'Popen',side_effect=child):
                if tamper_during_run:
                    with self.assertRaises(ValueError):campaign.main()
                else:
                    self.assertEqual(campaign.main(),0)
            if tamper_during_run:
                self.assertEqual(calls,['p3-s0000-run'])
                self.assertEqual((art/'test/analysis-source-1.py').read_text(),'claims-v1')
                self.assertEqual(campaign.load(art/'test/state.json')['status'],'incomplete')
                self.assertEqual(list((art/'test').glob('analysis-*.json')),[])
                return
            self.assertEqual(calls,['p3-s0000-run','p3-s0000-replay','p3-s0000-workers',
                                    'p3-s0000-workers-replay','p3-s0001-run'])
            state=campaign.load(art/'test/state.json')
            analysis=campaign.load(art/'test'/f'analysis-{state["revision"]:06d}.json')[0]
            self.assertEqual(analysis['complete_blocks'],2)
            self.assertEqual(analysis['finalized_games'],24)
            self.assertEqual(analysis['replayed_finalized_games'],12)
            self.assertEqual(analysis['worker_checked_finalized_games'],12)
            self.assertEqual(state['units'][1]['stages'].keys(),{'run'})
            # The selection itself is frozen, even if a later choice would still
            # be syntactically valid and include all worker-check units.
            if schema==campaign.SCHEMA_V3:
                claims.write_text('changed analysis before resume')
            else:
                plan['replay_units'].append('p3-s0001')
                planfile.write_bytes(campaign.encoded(plan))
            with patch.object(campaign,'ROOT',root),patch.object(campaign,'__file__',str(runner)),patch.object(campaign,'ARTIFACTS',art),patch.object(campaign.sys,'argv',argv+['--resume']),patch.object(campaign.subprocess,'Popen',side_effect=AssertionError('mutated plan launched process')):
                with self.assertRaises(ValueError):campaign.main()


class FrozenAnalysisSourcesReview(unittest.TestCase):
    def test_v3_freezes_retained_analysis_bytes_and_detects_tampering(self):
        from unittest.mock import patch
        with tempfile.TemporaryDirectory() as d:
            root=Path(d);(root/'xops').mkdir();dest=root/'artifacts';dest.mkdir()
            runner=root/'xops/sim_gameplay_evaluate.py';runner.write_text('runner-v1')
            claims=root/'xops/sim_claims.py';claims.write_text('claims-v1')
            files={}
            for name in ('binary','config','source','profile'):
                files[name]=root/name;files[name].write_text('{}')
            plan={'schema':'cgms-gameplay-campaign-v3','config':str(files['config']),
                  'profiles':[str(files['profile'])],'source_manifest':str(files['source']),
                  'shard_blocks':1,'wall_seconds':1,'storage_bytes':1000000,
                  'worker_check_units':[],'worker_check_workers':1,'analysis_seed':'test',
                  'resamples':100,'replay_units':[],
                  'analysis_sources':[str(runner),str(claims)]}
            path=root/'plan';path.write_bytes(campaign.encoded(plan))
            with patch.object(campaign,'ROOT',root),patch.object(campaign,'__file__',str(runner)):
                frozen=campaign.frozen_inputs(files['binary'],path,plan)
                campaign.retain_analysis_sources(dest,frozen)
                campaign.verify_analysis_sources(dest,frozen)
                self.assertEqual((dest/'analysis-source-0.py').read_text(),'runner-v1')
                self.assertEqual((dest/'analysis-source-1.py').read_text(),'claims-v1')
                runner.write_text('runner-v2')
                with self.assertRaises(ValueError):campaign.verify_frozen_files(frozen)
                self.assertNotEqual(campaign.frozen_inputs(files['binary'],path,plan),frozen)
                runner.write_text('runner-v1')
                (dest/'analysis-source-1.py').write_text('tampered copy')
                with self.assertRaises(ValueError):campaign.verify_analysis_sources(dest,frozen)

    def test_v3_requires_both_analysis_implementations_and_distinct_sources(self):
        from unittest.mock import patch
        with tempfile.TemporaryDirectory() as d:
            root=Path(d);(root/'xops').mkdir()
            runner=root/'xops/sim_gameplay_evaluate.py';runner.write_text('runner')
            claims=root/'xops/sim_claims.py';claims.write_text('claims')
            with patch.object(campaign,'ROOT',root),patch.object(campaign,'__file__',str(runner)):
                for sources in ([],[str(runner)],[str(runner),str(runner)]):
                    with self.assertRaises(ValueError):campaign.analysis_source_paths({'analysis_sources':sources})


class SeparatePracticalClaims(unittest.TestCase):
    def test_positive_but_too_small_is_not_practically_meaningful(self):
        r={'independent_blocks':1000,'mean':'2','lower':1,'upper':4}
        out=campaign.classify_interval(r,5,5)
        self.assertEqual(out['positive_improvement'],'supported')
        self.assertEqual(out['strength'],'refuted')
        self.assertEqual(out['noninferiority'],'supported')

    def test_negative_but_tolerable_remains_noninferior(self):
        r={'independent_blocks':1000,'mean':'-2','lower':-4,'upper':-1}
        out=campaign.classify_interval(r,5,5)
        self.assertEqual(out['positive_improvement'],'refuted')
        self.assertEqual(out['noninferiority'],'supported')
        r.update(lower=-8,upper=-6)
        self.assertEqual(campaign.classify_interval(r,5,5)['noninferiority'],'refuted')
