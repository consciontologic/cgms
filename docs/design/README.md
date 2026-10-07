# 🎨 `docs/design/` — design docs & ADRs

The approved UI baseline, design proposals and ADRs (Architecture Decision
Records — accepted decisions with rationale).

## Files

- [UI baseline v2](UI_BASELINE.md) — owner-authorized adaptive desktop/phone/tablet contract (2026-09-30), preserved cgmsart and explicit browser acceptance.
- [Flutter web client](../../client/README.md) — the authoritative online UI in `client/`, distinct adaptive desktop/phone/tablet compositions, retained local v1 scenarios and reusable `cgms_ui` package. See the [integration instructions](../../client/INTEGRATION.md) and dated browser evidence for verified scope.
- [Mobile flow mockups 01–20](mockup/README.md) — primary layout, hierarchy, navigation and interaction references, paired with the cgmsart deck library for current artwork. Adapt static arrangements to current gameplay/privacy and the adaptive baseline.

- [Go architecture](../code/ARCHITECTURE.md) and [game API](../code/API-game.md) — architecture and implemented local online authority contract; see roadmap evidence.
- [`DESIGN-go-state-and-delivery.md`](DESIGN-go-state-and-delivery.md) — proposed transaction, idempotency, delivery and concurrency contracts.
- [Economy and store boundaries](DESIGN-economy-and-stores.md) — adopted UTC/allowance/pass policy, dated Apple/Google applicability research and remaining P05/P06 decisions.
- [`ADR-0001-shared-go-simulation-engine.md`](ADR-0001-shared-go-simulation-engine.md) — proposed shared engine for server and [CLI simulation](../../sims/README.md).
- [`RULES-IMPLEMENTATION-QUESTIONS.md`](RULES-IMPLEMENTATION-QUESTIONS.md) — resolved Q01–Q12 decisions, adopted F01–F10/N01–N07/T01–T03 amendments and accepted rule/test obligations.

- [CGMS art-direction skill](../../.agents/skills/cgms-art-direction/SKILL.md) — guidance for applying the adaptive baseline while preserving game, visibility, accessibility and platform constraints.
- [`IDEA.docx`](IDEA.docx) — original rules source document, retained as historical material.
- [GAME_RULES.md](../project/GAME_RULES.md) — sole current game authority, finalized with Q01–Q12/F01–F10/N01–N07/T01–T03: decision and cost rules, combat algebra, persistent Aces, restrictions, game identity, financial provenance, visibility, proposals, draw order and purchases.
- [Project documents](../project/README.md) — completed charter, decision log and glossary alongside the canonical rules.
- [Third frontend/backend assessment](../reports/2026-09-28-final-draft-third-assessment.md) — historical source of the now-adopted T01–T03 recommendations. The [first](../reports/2026-09-28-final-draft-implementation-review.md) and [second](../reports/2026-09-28-final-draft-second-assessment.md) reports are preserved too; none certifies runtime behavior.
- [Four cgmsart decks](cgmsart-decks/README.md) — First Light, Ember Glaze, Rain Glaze and Drypoint: four cosmetic 52-card editions, labelled contact sheets, comparisons, mobile previews and separate artwork/label layers; two-deck rules unchanged.
- [Change notes](../../CHANGELOG.md) — agreed decisions, rationale, and historical proposals.
- [Draft archive](drafts/README.md) — all four historical rule drafts, including [FINAL_DRAFT.md](drafts/FINAL_DRAFT.md) before T01–T03; use GAME_RULES for current play.
- [`THIRD_DRAFT.md`](drafts/THIRD_DRAFT.md) — previous working rules, retained as history with their original wording.
- [`SECOND_DRAFT.md`](drafts/SECOND_DRAFT.md) — previous consolidated rules, retained for reference.
- [`FIRST_DRAFT.md`](drafts/FIRST_DRAFT.md) — previous rules and amendment history.

UI v1 remains a local presentation reference with retained evidence. V2 adaptive
online presentation is implemented and browser-tested under P04. Preserve theme,
rules, privacy, deck library and mockups when extending its distinct device-specific
navigation and components. Native packaging/device/lifecycle/performance and
platform acceptance are deferred to R06N before release, not waived.

## When to write

- **Design doc**: before a change > a few days of work, before a public API, before a cross-module refactor. Reviewed by humans + agents, signed off before code lands.
- **ADR**: after a meaningful architectural decision, so future-you can ask "why is it this way?" and get an answer.

Both live alongside the code they shape — file name embeds the topic
(`DESIGN-auth-rework.md`, `ADR-0007-use-postgres-not-mongo.md`).
