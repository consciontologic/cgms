import copy
from fractions import Fraction
import unittest

import sim_claims as claims


def fixture():
    p = {'population': 3, 'games_per_match': 3, 'block_ids': ['a', 'b'],
         'rotations': [{'id': f'r{i}', 'seats': [f'p{(j+i)%3}' for j in range(3)]} for i in range(3)],
         'candidate': {'participant_id': 'p0'}, 'budgets': {},
         'bots': [{'participant_id': f'p{i}', 'policy': 'economic', 'version': 'v1'} for i in range(3)]}
    rows = []
    for block in p['block_ids']:
        for rotation in p['rotations']:
            for treatment in ('baseline', 'candidate'):
                scores = [9, 9, 0]
                if treatment == 'candidate':
                    scores[rotation['seats'].index('p0')] += 18
                rows.append({'block': block, 'rotation': rotation['id'], 'treatment': treatment,
                             'population': 3, 'seats': rotation['seats'], 'planned_game_slots': 3,
                             'started_games': 3, 'board_ended_games': 3, 'finalized_games': 3,
                             'match_complete': True, 'status': 'complete',
                             'scores': [{'numerator': str(s), 'denominator': '1'} for s in scores]})
    return p, rows


class ClaimsTests(unittest.TestCase):
    def test_per_game_strength_and_whole_blocks(self):
        p, rows = fixture()
        self.assertEqual(claims.strength_blocks(p, rows), [Fraction(6), Fraction(6)])
        rows[-1].update(match_complete=False, finalized_games=2, status='budget-exhausted')
        self.assertEqual(claims.strength_blocks(p, rows), [Fraction(6)])

    def test_seat_contrasts_use_physical_seats_tied_win_credit_and_baseline_only(self):
        p, rows = fixture()
        result = claims.seat_blocks(p, rows)
        self.assertEqual(result['score_per_game']['0-2'], [Fraction(3), Fraction(3)])
        self.assertEqual(result['score_per_game']['0-1'], [Fraction(0), Fraction(0)])
        self.assertEqual(result['win_share']['0-2'], [Fraction(1, 2), Fraction(1, 2)])
        rows[-1].update(match_complete=False, finalized_games=2, status='budget-exhausted')
        self.assertEqual(claims.seat_blocks(p, rows), result)

    def test_incomplete_baseline_excludes_entire_seat_block(self):
        p, rows = fixture()
        rows[0].update(match_complete=False, finalized_games=2, status='budget-exhausted')
        self.assertEqual(claims.seat_blocks(p, rows)['score_per_game']['0-2'], [Fraction(3)])

    def test_homogeneity_compares_parameters_not_just_policy_names(self):
        p, rows = fixture()
        p['bots'][1]['weights'] = [{'name': 'x', 'value': 1}]
        with self.assertRaises(ValueError):
            claims.seat_blocks(p, rows)

    def test_completion_keeps_never_started_and_incomplete_separate(self):
        p, rows = fixture()
        rows.pop()
        rows[1].update(match_complete=False, finalized_games=1, started_games=2,
                       board_ended_games=1, status='budget-exhausted')
        result = claims.completion_by_treatment(p, rows)['candidate']
        self.assertEqual(result['planned_matches'], 6)
        self.assertEqual(result['started_matches'], 5)
        self.assertEqual(result['finalized_matches'], 4)
        self.assertEqual(result['incomplete_started_matches'], 1)
        self.assertEqual(result['never_started_matches'], 1)
        self.assertEqual(result['planned_games'], 18)
        self.assertEqual(result['started_games'], 14)
        self.assertEqual(result['finalized_games'], 13)
        self.assertEqual(result['never_started_games'], 4)

    def test_duplicate_wrong_seats_and_population_rejected(self):
        p, rows = fixture()
        wrong_seats = copy.deepcopy(rows)
        wrong_seats[0]['seats'] = list(reversed(wrong_seats[0]['seats']))
        for variant in (rows + [rows[0]], wrong_seats):
            with self.assertRaises(ValueError):
                claims.strength_blocks(p, variant)
        rows[0]['population'] = 4
        with self.assertRaises(ValueError):
            claims.completion_by_treatment(p, rows)

    def test_unknown_dispatch_is_not_reclassified_as_never_started(self):
        p, rows = fixture()
        rows[0].update(status='unverified-outcomes', unknown_game_slots=3,
                       match_complete=False, finalized_games=0)
        with self.assertRaises(ValueError):
            claims.completion_by_treatment(p, rows)

    def test_four_player_full_permutations_and_tied_rank(self):
        from itertools import permutations
        p, _ = fixture()
        p.update(population=4, block_ids=['a'],
                 rotations=[{'id': f'r{i}', 'seats': list(seats)}
                            for i, seats in enumerate(permutations(['p0', 'p1', 'p2', 'p3']))])
        p['bots'].append({'participant_id': 'p3', 'policy': 'economic', 'version': 'v1'})
        rows = []
        for r in p['rotations']:
            rows.append({'block': 'a', 'rotation': r['id'], 'treatment': 'baseline',
                         'population': 4, 'seats': r['seats'], 'planned_game_slots': 3,
                         'started_games': 3, 'board_ended_games': 3, 'finalized_games': 3,
                         'match_complete': True, 'status': 'complete',
                         'scores': [{'numerator': str(s), 'denominator': '1'} for s in (12, 12, 12, 0)]})
        result = claims.seat_blocks(p, rows)
        self.assertEqual(len(result['score_per_game']), 6)
        self.assertEqual(result['score_per_game']['0-3'], [Fraction(4)])
        self.assertEqual(result['win_share']['2-3'], [Fraction(1, 3)])
        self.assertEqual(claims.strength_blocks(p, rows), [])


if __name__ == '__main__':
    unittest.main()
