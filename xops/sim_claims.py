"""Exact new-cohort claim vectors; callers must verify immutable artifacts first.

No game execution or inference happens here. One value is one complete seed
block; financial scores are divided by the fixed match horizon, while match-win
shares retain whole-match rank semantics. Historical estimators are unchanged.
"""
from fractions import Fraction
from itertools import combinations

from sim_gameplay_evaluate import (build_units, decode_score, encoded,
                                   paired_blocks, require, validate_outcomes)


def _rows(profile, outcomes):
    """Validate observed rows and explicitly materialize never-started units.

    Missing input rows mean never dispatched, so callers must not omit dispatched
    units with unverified artifacts. Those require a separate unknown-work report
    and cannot be passed to this verified-outcome extractor.
    """
    build_units([profile], 1)
    keys = {(b, r['id'], t) for b in profile['block_ids']
            for r in profile['rotations'] for t in ('baseline', 'candidate')}
    rows = {}
    schedules = {r['id']: r['seats'] for r in profile['rotations']}
    for row in outcomes:
        key = (row['block'], row['rotation'], row['treatment'])
        require(key in keys and key not in rows, 'unplanned or duplicate outcome')
        require(row['seats'] == schedules[row['rotation']], 'seat schedule mismatch')
        rows[key] = row
    horizon = profile['games_per_match']
    require(type(horizon) is int and horizon > 0, 'positive match horizon')
    for b, r, t in sorted(keys - rows.keys()):
        rows[(b, r, t)] = {'block': b, 'rotation': r, 'treatment': t,
                            'population': profile['population'], 'seats': schedules[r],
                            'planned_game_slots': horizon, 'started_games': 0,
                            'board_ended_games': 0, 'finalized_games': 0,
                            'match_complete': False, 'status': 'not-started'}
    result = list(rows.values())
    validate_outcomes(profile, result, False)
    paired_blocks(profile, result)
    return rows


def strength_blocks(profile, outcomes):
    """Candidate-minus-baseline focal score per game, all paired rotations."""
    rows = _rows(profile, outcomes)
    values, _ = paired_blocks(profile, rows.values())
    return [value / profile['games_per_match'] for value in values]


def seat_blocks(profile, outcomes):
    """Homogeneous baseline physical-seat contrasts; candidate rows excluded.

    Return separate score/game and tied match-win-share vectors for each pair.
    Every baseline rotation in a block must complete to contribute any contrast.
    """
    rows = _rows(profile, outcomes)
    bots = profile['bots']
    participants = set(profile['rotations'][0]['seats'])
    require(len(bots) == len(participants) and
            {b['participant_id'] for b in bots} == participants, 'baseline roster mismatch')
    identities = {encoded({k: v for k, v in b.items() if k != 'participant_id'}) for b in bots}
    require(len(identities) == 1, 'seat fairness requires identical baseline policies and parameters')
    pairs = list(combinations(range(profile['population']), 2))
    result = {metric: {f'{i}-{j}': [] for i, j in pairs}
              for metric in ('score_per_game', 'win_share')}
    for block in profile['block_ids']:
        block_rows = [rows[(block, r['id'], 'baseline')] for r in profile['rotations']]
        if not all(row['match_complete'] for row in block_rows):
            continue
        score_diffs = {pair: Fraction() for pair in pairs}
        win_diffs = {pair: Fraction() for pair in pairs}
        for row in block_rows:
            scores = [decode_score(s) for s in row['scores']]
            best = max(scores)
            credit = Fraction(1, scores.count(best))
            wins = [credit if s == best else Fraction() for s in scores]
            for i, j in pairs:
                score_diffs[(i, j)] += (scores[i] - scores[j]) / profile['games_per_match']
                win_diffs[(i, j)] += wins[i] - wins[j]
        for i, j in pairs:
            result['score_per_game'][f'{i}-{j}'].append(score_diffs[(i, j)] / len(block_rows))
            result['win_share'][f'{i}-{j}'].append(win_diffs[(i, j)] / len(block_rows))
    return result


def completion_by_treatment(profile, outcomes):
    """Exact planned/started/finalized counts, including missing planned rows."""
    rows = _rows(profile, outcomes)
    result = {}
    for treatment in ('baseline', 'candidate'):
        selected = [row for (_, _, t), row in rows.items() if t == treatment]
        n = len(selected)
        started = sum(row['started_games'] > 0 for row in selected)
        finalized = sum(row['match_complete'] for row in selected)
        statuses = {}
        for row in selected:
            statuses[row['status']] = statuses.get(row['status'], 0) + 1
        result[treatment] = {
            'planned_matches': n, 'started_matches': started,
            'finalized_matches': finalized, 'incomplete_started_matches': started - finalized,
            'never_started_matches': n - started,
            'planned_games': n * profile['games_per_match'],
            **{k: sum(row[k] for row in selected)
               for k in ('started_games', 'board_ended_games', 'finalized_games')},
            'never_started_games': n * profile['games_per_match'] - sum(row['started_games'] for row in selected),
            'statuses': statuses}
    return result
