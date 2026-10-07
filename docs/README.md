# 📚 `docs/`

This folder is the human-readable side of the project. Source of truth for
agents on **what the project is and where it's going**; source of truth for
humans on **the project's design and history**.

## Layout

| Path | Purpose | Audience |
|---|---|---|
| [`code/`](code/) | Module-level documentation (architecture, modules, APIs). | Devs joining the codebase. |
| [`project/`](project/) | Canonical game rules, product brief, charter, decision log and glossary. | Players and contributors. |
| [`design/`](design/) | Approved UI baseline, design docs and ADRs. | Reviewers + future-you. |
| [`planning/`](planning/) | The **ROADMAP** — single source of truth for sequenced work. | Agents + humans. |
| [`tracking/`](tracking/) | How the `docs/tracking/tracking.csv` workflow is used. | Agents. |
| [`guides/`](guides/) | Cross-cutting how-tos: agent operating model, model profiles, MCP usage. | Agents + ops. |
| [`reports/`](reports/) | Generated reports (audit, status snapshots). | Reviewers. |
| [`.agents/skills/`](../.agents/skills/) | The **skill library** — load on demand. | Agents. |

## Phase 0 preparation

The [Astra handoff](../sims/README.md#phase-0-preparation-and-handoff) lists the exact
loading order for simulation research, repository skills/instructions, preparation
schemas/examples and the S01–S10 execution guide. These artifacts are preparation;
that historical preparation delivered no runtime. Current runtime gate status is
recorded in the [roadmap](planning/ROADMAP.md), with
[accepted gameplay evidence](../sims/docs/gameplay-coverage.md) and
[measured gameplay findings](../sims/docs/gameplay-findings.md).

## Discoverability rule

For backend planning, start with the [architecture proposal](code/ARCHITECTURE.md),
[implemented game API](code/API-game.md), [OpenAPI schema](api/online.openapi.json), [state/delivery design](design/DESIGN-go-state-and-delivery.md)
and [simulation contract](../sims/README.md). The
[roadmap](planning/ROADMAP.md) is the only sequenced implementation checklist;
the [rule register](design/RULES-IMPLEMENTATION-QUESTIONS.md) maps the accepted
[canonical game rules](project/GAME_RULES.md) to implementation and test obligations.
Q01–Q12, F01–F10, N01–N07 and T01–T03 recommendations are adopted.
The [third assessment](reports/2026-09-28-final-draft-third-assessment.md)
is the historical source of the final adoption; all four [drafts](design/drafts/README.md)
are archived. Completed [project metadata](project/README.md) supplies scope and
terminology. Architecture mixes implemented local services with planned product
and deployment work; the API/module records and roadmap identify the evidence
and remaining gates. Adaptive web has recorded acceptance after relocation;
native acceptance remains a separate final release gate.

For frontend work, use [UI baseline v2](design/UI_BASELINE.md) and the
[existing Flutter host](../client/README.md) with its `packages/cgms_ui` package.
Flutter web is the current development/testing platform. Distinct desktop,
phone and tablet presentations are implemented P04 work; phone/tablet web defines
native appearance and journeys. The baseline contains the browser test matrix;
R06N retains separate final native acceptance. Existing UI v1 evidence remains
historical. The [20 mockups](design/mockup/README.md) are the primary layout/UX references, paired
with [cgmsart-decks](design/cgmsart-decks/README.md) for current artwork.
[GAME_RULES.md](project/GAME_RULES.md) still governs gameplay and visibility.

Before creating a new doc, search for an existing one. Follow the relevant
folder's README and existing documents, such as the
[shared game engine module](code/MODULE-game.md).
