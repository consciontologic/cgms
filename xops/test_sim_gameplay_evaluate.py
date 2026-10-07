import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location('campaign', Path(__file__).with_name('sim_gameplay_evaluate.py'))
campaign = importlib.util.module_from_spec(spec)
spec.loader.exec_module(campaign)


class CampaignTests(unittest.TestCase):
    def test_shards_preserve_every_block_and_rotation(self):
        p = {'population': 3, 'block_ids': ['b'+str(i) for i in range(11)], 'rotations': [{'id': 'r'+str(i), 'seats': (['p0','p1','p2'][i:]+['p0','p1','p2'][:i])} for i in range(3)], 'games_per_match': 2, 'bots': [], 'candidate': {}, 'budgets': {'max_matches': 1000}}
        units = campaign.build_units([p], 5)
        self.assertEqual([len(u['experiment']['block_ids']) for u in units], [5,5,1])
        self.assertEqual(sum(u['planned_matches'] for u in units), 66)
        self.assertEqual(sum(u['planned_games'] for u in units), 132)
        self.assertEqual([b for u in units for b in u['experiment']['block_ids']], p['block_ids'])



class AnalysisTests(unittest.TestCase):
    def test_cluster_unit_and_exact_constant(self):
        result=campaign.cluster_interval(['3/2']*3,'analysis-private',100)
        self.assertEqual((result['mean'],result['lower_95'],result['upper_95']),('3/2','3/2','3/2'))
        self.assertEqual(result['complete_blocks'],3)
        self.assertIsNone(campaign.cluster_interval([], 'a',100)['mean'])

    def test_partial_rotation_excludes_whole_block_without_dropping_units(self):
        p={'block_ids':['a','b'],'candidate':{'participant_id':'p0'},'games_per_match':2,'rotations':[{'id':'r0','seats':['p0','p1','p2']},{'id':'r1','seats':['p2','p0','p1']},{'id':'r2','seats':['p1','p2','p0']}]}
        rows=[]
        for rotation in p['rotations']:
            for treatment in ('baseline','candidate'):
                scores=[{'numerator':'0','denominator':'1'} for _ in range(3)];scores[rotation['seats'].index('p0')]={'numerator':'10' if treatment=='candidate' else '6','denominator':'1'}
                rows.append({'block':'a','rotation':rotation['id'],'treatment':treatment,'status':'complete','match_complete':True,'finalized_games':2,'seats':rotation['seats'],'scores':scores})
        rows.append({'block':'b','rotation':'r0','treatment':'baseline','status':'budget-exhausted','match_complete':False})
        values,statuses=campaign.paired_blocks(p,rows)
        self.assertEqual(values,[campaign.Fraction(4)])
        self.assertEqual(sum(statuses.values()),12)
        self.assertEqual(statuses['not-started'],5)
        self.assertEqual(statuses['budget-exhausted'],1)
        with self.assertRaises(ValueError):campaign.paired_blocks(p,rows+rows[:1])

    def test_duplicate_json_and_no_overwrite(self):
        import tempfile
        with tempfile.TemporaryDirectory() as d:
            path=Path(d)/'x.json';path.write_text('{"x":1,"x":2}')
            with self.assertRaises(ValueError):campaign.load(path)
            with self.assertRaises(FileExistsError):campaign.immutable_json(path,{})

    def test_shard_bound(self):
        with self.assertRaises(ValueError):campaign.build_units([],6)
        with self.assertRaises(ValueError):campaign.build_units([],True)

class IntegrityTests(unittest.TestCase):
    def test_missing_outcomes_cannot_complete_campaign(self):
        profile={'population':3,'games_per_match':2,'block_ids':['b'],'rotations':[{'id':'r0'}]}
        with self.assertRaises(ValueError):campaign.validate_outcomes(profile,[],True)
        rows=[{'population':3,'block':'b','rotation':'r0','treatment':t,'planned_game_slots':2,'status':'complete','match_complete':True,'finalized_games':2,'started_games':2,'board_ended_games':2} for t in ('baseline','candidate')]
        campaign.validate_outcomes(profile,rows,True)
        rows[0]['finalized_games']=1
        with self.assertRaises(ValueError):campaign.validate_outcomes(profile,rows,True)

    def test_journal_and_atomic_pointer(self):
        import tempfile
        with tempfile.TemporaryDirectory() as d:
            dest=Path(d);state={'revision':0,'units':[]}
            campaign.save_state(dest,state)
            self.assertEqual(campaign.load(dest/'state.json')['revision'],1)
            campaign.save_state(dest,state)
            self.assertTrue((dest/'state-000001.json').exists())
            self.assertTrue((dest/'state-000002.json').exists())
            self.assertEqual(campaign.load(dest/'state.json')['revision'],2)

class ResumeTests(unittest.TestCase):
    def test_budget_stop_resume_keeps_plan_and_rejects_changed_binary(self):
        import tempfile
        from unittest.mock import patch
        import json
        with tempfile.TemporaryDirectory() as d:
            root=Path(d);art=root/'artifacts';art.mkdir()
            binary=root/'sim';binary.write_bytes(b'not executed')
            config=root/'rules.json';config.write_text('{}')
            source=root/'source.json';source.write_text('{}')
            profile=root/'profile.json';profile.write_text(json.dumps({'population':3,'block_ids':['b0'],'rotations':[{'id':f'r{i}','seats':(['p0','p1','p2'][i:]+['p0','p1','p2'][:i])} for i in range(3)],'candidate':{'participant_id':'p0'},'games_per_match':2,'budgets':{'max_matches':6}}))
            plan=root/'plan.json';plan.write_text(json.dumps({'schema':campaign.SCHEMA,'config':str(config),'profiles':[str(profile)],'source_manifest':str(source),'shard_blocks':5,'wall_seconds':1,'storage_bytes':1000000,'worker_check_units':[],'worker_check_workers':2,'analysis_seed':'private-analysis','resamples':100}))
            argv=['campaign',str(binary),str(plan),'test']
            with patch.object(campaign,'ARTIFACTS',art),patch.object(campaign.sys,'argv',argv),patch.object(campaign.time,'monotonic',side_effect=[0,2,3]):
                self.assertEqual(campaign.main(),5)
            state=campaign.load(art/'test/state.json')
            self.assertEqual(state['status'],'wall-budget')
            self.assertEqual(state['units'][0]['stages'],{})
            reports=list((art/'test').glob('report-*.md'))
            self.assertNotIn('private-analysis',reports[0].read_text())
            binary.write_bytes(b'changed')
            with patch.object(campaign,'ARTIFACTS',art),patch.object(campaign.sys,'argv',argv+['--resume']):
                with self.assertRaises(ValueError):campaign.main()

if __name__ == '__main__':
    unittest.main()
