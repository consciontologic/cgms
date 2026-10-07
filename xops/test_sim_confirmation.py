import copy
from fractions import Fraction
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import sim_confirmation as confirmation
import sim_gameplay_evaluate as campaign


def fixture(pop=3, lineup='mixed', scope='primary'):
    seats=[f'p{i}' for i in range(pop)]
    names=['economic' if i%2==0 else 'pressure' for i in range(pop)] if lineup=='mixed' else [lineup]*pop
    profile={'id':f'confirmation-v3-{scope}-{lineup}-{pop}p','population':pop,'games_per_match':3,'block_ids':['held-a'],
             'phase':'held-out','rules_profile':'accepted-game-rules','root_seed':f'private-test-root-{pop}',
             'pairing_id':'fixed','rotations':[{'id':f'r{i}','seats':seats[i:]+seats[:i]} for i in range(pop)],
             'bots':[{'participant_id':s,'policy':p,'version':'v2','information':'seat-projection','features':[],'weights':[]} for s,p in zip(seats,names)],
             'candidate':{'participant_id':'p0','policy':'opportunity','version':'v2','information':'seat-projection','features':[],'weights':[]},
             'output':{'omniscient_comparison':False},
             'budgets':{'workers':1,'max_menu_operations_per_decision':1000000}}
    rows=[]
    for r in profile['rotations']:
        for t in ['baseline','candidate']:
            scores=[0]*pop
            if t=='candidate':scores[r['seats'].index('p0')]=18
            rows.append({'block':'held-a','rotation':r['id'],'treatment':t,'population':pop,
                         'planned_game_slots':3,'started_games':3,'board_ended_games':3,
                         'finalized_games':3,'status':'complete','match_complete':True,
                         'seats':r['seats'],'scores':[{'numerator':str(v),'denominator':'1'} for v in scores]})
    return profile,rows


class ConfirmationTests(unittest.TestCase):
    def test_fixed_family_counts_and_per_game_units(self):
        cohorts={}
        for lineup in ('mixed','economic','pressure'):
            cohorts[lineup]=[confirmation.population_data(*fixture(pop,lineup,'roster')) for pop in (3,4)]
        report=confirmation.aggregate('roster',cohorts)
        self.assertEqual(report['family_comparisons'],42)
        self.assertEqual(len(report['contrasts']),42)
        strength=[r for r in report['contrasts'] if r['kind']=='strength']
        self.assertEqual(len(strength),6)
        self.assertTrue(all(r['interval']['mean']=='6' for r in strength))
        self.assertTrue(all(r['labels']['practical_strength']=='inconclusive' for r in strength))
        primary={'mixed':[confirmation.population_data(*fixture(pop)) for pop in (3,4)]}
        self.assertEqual(confirmation.aggregate('primary',primary)['family_comparisons'],2)

    def test_pressure_focal_must_also_be_pressure(self):
        p,rows=fixture(4,'pressure');p['bots'][0]['policy']='economic'
        with self.assertRaises(ValueError):confirmation.validate_roster(p,'pressure')

    def test_omniscient_policy_or_report_rejected(self):
        p,_=fixture()
        for location in ('baseline','candidate','output'):
            bad=copy.deepcopy(p)
            if location=='baseline':bad['bots'][1]['information']='omniscient'
            elif location=='candidate':bad['candidate']['information']='omniscient'
            else:bad['output']['omniscient_comparison']=True
            with self.assertRaises(ValueError):confirmation.validate_roster(bad,'mixed')

    def test_legacy_or_mixed_tie_versions_are_not_confirmation(self):
        p,_=fixture()
        for key in ('baseline','candidate'):
            bad=copy.deepcopy(p)
            if key=='baseline':bad['bots'][0]['version']='v1'
            else:bad['candidate']['version']='v1'
            with self.assertRaises(ValueError):confirmation.validate_roster(bad,'mixed')

    def test_incomplete_block_and_missing_slot_stay_visible(self):
        p,rows=fixture()
        rows[-1].update(match_complete=False,started_games=2,board_ended_games=2,finalized_games=2,status='budget-exhausted')
        data=confirmation.population_data(p,rows)
        self.assertEqual(data['strength_values'],[])
        self.assertEqual(data['completion']['candidate']['never_started_games'],1)
        self.assertEqual(data['completion']['candidate']['incomplete_started_matches'],1)

    def test_interval_labels_are_distinct_and_completion_gated(self):
        interval={'independent_blocks':1000,'lower':1,'upper':4}
        labels=confirmation.labels(interval,True)
        self.assertEqual(labels['positive_effect'],'supported')
        self.assertEqual(labels['practical_strength'],'refuted')
        self.assertEqual(labels['noninferiority'],'supported')
        self.assertEqual(labels['equivalence'],'supported')
        self.assertTrue(all(v=='inconclusive' for v in confirmation.labels(interval,False).values()))

    def test_public_report_omits_private_inputs(self):
        cohorts={'mixed':[confirmation.population_data(*fixture(pop)) for pop in (3,4)]}
        text=confirmation.markdown(confirmation.aggregate('primary',cohorts))
        self.assertNotIn('private-test-root',text)
        self.assertNotIn('held-a',text)
        self.assertIn('inconclusive',text)

    def test_missing_replay_verification_withholds_large_sample_claims(self):
        cohorts={'mixed':[confirmation.population_data(*fixture(pop)) for pop in (3,4)]}
        for data in cohorts['mixed']:
            data['strength_values']=[Fraction(20+i%2) for i in range(1000)]
        report=confirmation.aggregate('primary',cohorts)
        self.assertTrue(all(c['labels']['practical_strength']=='inconclusive' for c in report['contrasts']))

    def test_scope_and_population_root_freeze(self):
        cohorts={'mixed':[confirmation.population_data(*fixture(pop,'mixed','roster')) for pop in (3,4)]}
        with self.assertRaises(ValueError):confirmation.aggregate('primary',cohorts)
        cohorts={'mixed':[confirmation.population_data(*fixture(pop)) for pop in (3,4)]}
        cohorts['mixed'][1]['profile']['root_seed']=cohorts['mixed'][0]['profile']['root_seed']
        with self.assertRaises(ValueError):confirmation.aggregate('primary',cohorts)

    def test_authenticated_campaign_rejects_artifact_and_retained_source_tampering(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary);art=root/'artifacts';dest=art/'campaign';dest.mkdir(parents=True)
            binary=root/'binary';binary.write_text('pinned')
            config=root/'config';config.write_text('{}')
            source=root/'source';source.write_text('{}')
            p,rows=fixture();pfile=root/'profile';pfile.write_bytes(campaign.encoded(p))
            sources=[str(Path(campaign.__file__).resolve()),str(Path(confirmation.__file__).with_name('sim_claims.py').resolve()),str(Path(confirmation.__file__).resolve())]
            plan={'schema':campaign.SCHEMA_V3,'config':str(config),'source_manifest':str(source),
                  'profiles':[str(pfile)],'analysis_sources':sources,'shard_blocks':1,
                  'wall_seconds':60,'storage_bytes':1000000,'worker_check_units':[],
                  'worker_check_workers':1,'analysis_seed':'private-analysis','resamples':100,'replay_units':[]}
            planfile=root/'plan';planfile.write_bytes(campaign.encoded(plan))
            frozen=campaign.frozen_inputs(binary,planfile,plan);units=campaign.build_units([p],1)
            campaign.retain_analysis_sources(dest,frozen)
            campaign.immutable_json(dest/'plan.json',plan)
            campaign.immutable_json(dest/'p3-s0000-experiment.json',units[0]['experiment'])
            run=dest/'p3-s0000-run';run.mkdir()
            artifacts=[]
            for name,raw in [('report.md',b'public'),('outcomes.json',campaign.encoded(rows)),('effective-experiment.json',campaign.encoded(units[0]['experiment']))]:
                (run/name).write_bytes(raw);artifacts.append({'path':name,'bytes':len(raw),'sha256':campaign.digest(raw)})
            manifest={'schema_version':'cgms-manifest-v1','exit_code':0,'privacy':'restricted',
                      'provenance':{'binary_sha256':frozen['binary']['sha256']},'artifacts':artifacts}
            campaign.immutable_json(run/'manifest.json',manifest)
            state={'schema':campaign.SCHEMA_V3,'status':'complete','invocation_active':False,
                   'plan_hash':campaign.digest(campaign.encoded(plan)),'frozen':frozen,
                   'units':[dict({k:units[0][k] for k in ('id','population','planned_matches','planned_games')},stages={'run':{'status':'complete','exit_code':0,'manifest_sha256':campaign.digest(campaign.encoded(manifest))}})]}
            campaign.immutable_json(dest/'state-000001.json',state)
            with patch.object(campaign,'ARTIFACTS',art):
                loaded=confirmation.read_campaign(dest)
                self.assertEqual(loaded['populations'][0]['strength_values'],[Fraction(6)])
                changed=copy.deepcopy(state);changed['units'][0]['worker_identical']=True
                (dest/'state-000001.json').write_bytes(campaign.encoded(changed))
                with self.assertRaisesRegex(ValueError,'distinct worker counts'):
                    confirmation.read_campaign(dest)
                (dest/'state-000001.json').write_bytes(campaign.encoded(state))
                (run/'outcomes.json').write_text('tampered')
                with self.assertRaises(ValueError):confirmation.read_campaign(dest)
                (run/'outcomes.json').write_bytes(campaign.encoded(rows))
                (dest/'analysis-source-2.py').write_text('changed wrapper')
                with self.assertRaises(ValueError):confirmation.read_campaign(dest)


if __name__=='__main__':unittest.main()
