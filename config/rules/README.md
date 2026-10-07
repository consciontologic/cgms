# Canonical rules configuration — preparation specification

The canonical runtime defaults are `game-rules.json`, validated by
`../schemas/rules.schema.json` and `game.LoadRules`. The loader checks the exact
accepted content hash, merges partial overrides, rejects unknown fields and
changed accepted constants, and permits a positive fixed `match_games` count.
This document supplies the extraction contract, not a second defaults file.
[GAME_RULES](../../docs/project/GAME_RULES.md) is authoritative. Use the
[format specification](../../sims/docs/phase-0-format-spec.md) and keep schema
versions distinct from gameplay and engine versions.

Extract named constants with unit, clause and compatibility impact. At minimum:

| Fields to specify in S03 | Accepted value / allowed meaning | Source |
|---|---|---|
| players, decks, initial hand, rounds | 3 or 4; 2 decks / 104 distinct cards; 20 cards/seat; 13 rounds/game | §§1.1, 1.4 |
| match game count | Positive integer fixed before start; default 3; run/ordinary batch explicitly choose 1 | §§0, 1.1 |
| ordinary thresholds / trigger bonus | 13 number Diamonds or 15 physical royals plus four-suit history; +50 points | §1.2 |
| Coup award / obligation | +50 / −50 points; departure exemptions and no ordinary scoring | §§1.2, 4.4 |
| diamond bonus, purchase price | Default 5 points/card each; Code chooses diamond bonus 1–10, price 5–10 in live state | §§2.16, 3.2, 3.5 |
| Spade royal / club award | 10 points/royal at ordinary end; 3 points/physical attacking Club on successful ordinary Club elimination | §§3.5, 4.1 |
| ace numerical value / allowances | Integer 2–10; shared 1/round; Compensation 1/game | §§2.12, 3.4 |
| Inflation / removal | Exact half of number Spade value; one completed voluntary transaction spends ≥20 printed points | §3.4 |
| Compensation / restrictions | One draw per 5 qualifying losses, remainder retained; Barricade availability next round | §§2.9, 2.14, 3.4 |
| alliance sharing | Exactly two equal shares, exclusive pairs; not an adjustable partner count | §§2.4, 4.2 |
| remaining named abilities | Extract the complete §2.12 schedule and §§3.2–3.3 parameters when used; explicitly unsupported otherwise | §§2.12, 3.2–3.3 |

Accepted constants are `const`/bounded enums in the strict rules schema, not
arbitrary tuning knobs. Full catalog extraction does not implement every ability.
The sole effective rule-default source will be this canonical JSON; do not mirror
it in CLI flags, bot code or experiments. Schema `default` annotations do not fill
missing values. Normalize using that file, validate and hash the effective result.
Code's chosen economic settings belong to state, never hot-reloaded config.

Operational bounds (file bytes, matches, workers, rational-input digits) are
resource safeguards in an experiment; they do not cap legal score/debt or rounds
in an accepted game. A reached work limit yields incomplete execution, not a rule
ending. Changes to gameplay require an explicit versioned interpretation naming
replaced clauses; no such replacement is enabled by the examples. An empty sweep
grid is the accepted-profile baseline; a future grid can vary recorded bot
parameters or explicitly authorized interpretation parameters, never hidden rules.

## Runtime coverage boundary

The `ability_catalog` extracts every row of GAME_RULES §2.12, including timing,
support, quotas and committed versus resolution-only costs. Its presence is not
transition support. Engine coverage and scenarios must explicitly identify the
implemented rows; every unimplemented activation returns unsupported before
changing the last committed state. No config override can enable an ability.
The `AcceptedRulesHash` is a compatibility identity, not a duplicate defaults table.
Normalization and integer-only canonical hashing are tested in
`backend/internal/game/rules_test.go` and `backend/internal/canonical/`.
