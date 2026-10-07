# 📋 `docs/project/` — project meta-docs

The project's gameplay authority, product brief and supporting project metadata.

## Files

- [GAME_RULES.md](GAME_RULES.md) — canonical current gameplay, incorporating the accepted Q01–Q12, F01–F10, N01–N07 and T01–T03 rulings; all four earlier drafts remain in [the archive](../design/drafts/).
- [INITIATE.md](INITIATE.md) — owner-supplied product constraints and broader documentation brief; it does not certify implementation or supersede repository operating rules.
- [CHARTER.md](CHARTER.md) — vision, scope, audience, success criteria and explicit trade-offs.
- [DECISION_LOG.md](DECISION_LOG.md) — append-only project meta-decisions, with approval boundaries and unassigned owner identity stated explicitly; technical ADRs live in [design](../design/README.md).
- [GLOSSARY.md](GLOSSARY.md) — alphabetic game and project terminology.

The completed charter, decision log and glossary above are the maintained
project documents; no parallel template copies are present in this folder.

The [roadmap](../planning/ROADMAP.md) is the sole sequenced implementation
checklist. The [architecture](../code/ARCHITECTURE.md),
[API](../code/API-game.md) and [simulation contract](../../sims/README.md)
separate implemented local systems from planned product/deployment work.
Decision 0012 and [UI baseline v2](../design/UI_BASELINE.md) govern web-first
adaptive development and final native acceptance. Completing documents does not complete their
runtime gates or approve unmeasured capacity, economy values or store compliance.

## When to write project docs

- A new contributor or agent keeps asking "why are we doing X this way?" → CHARTER section.
- A meta-decision was made (scope cut, target audience shift, license change) → DECISION_LOG row.
- A domain term keeps being misunderstood → GLOSSARY entry.
