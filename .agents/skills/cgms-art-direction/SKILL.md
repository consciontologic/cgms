---
name: cgms-art-direction
description: Preserve and review the CGMS shared vertical UI, its cgmsart theme, shared Flutter board, illustrated cards, accessible interactions, and rule-faithful presentation.
---

# CGMS art direction — shared vertical board

Use for CGMS frontend implementation, maintenance, art and review. Backend-only
work does not require visual changes. This skill supplies constraints within the
user's task; it does not authorize a redesign or changes to game rules.

## Authority and adaptive revision

Read the [UI baseline](../../../docs/design/UI_BASELINE.md). The owner revision
of 2026-10-02 selects `client/screenshots/cgmsart-phone.png` as the board layout
at all widths. It supersedes the v2 distinct-pane/rail/page requirement and the
later minimum-card-size/no-overlap restriction wherever they conflict with the
selected compact reference. Use the existing `client/` host and `cgms_ui` package.

## Preserve identity while adapting size

Keep one viewport-contained board: compact status, a full-width table split into
four equal public-seat quadrants, the adjacent private hand, bounded action controls and
bottom actions. Never collapse the public seats to one column on narrow screens.
Size from available width AND height; shrinking public artwork must keep the
hand visible without scrolling the page. Detailed content may scroll internally.
Align the table's left/right edges with the hand. The owner's later selection
replaces square-only geometry with wider rectangles. Public cards share the
hand's size scale where space permits; native artwork fits without distortion,
and bounded fan depth prevents dense piles from continually shrinking faces.
Status and tools share one vertically centered header row, including compact
widths: smaller visible type/icons and selector spacing preserve full ordinary-size
labels and 48px targets. Enlarged text may wrap within the status area. Group utilities and
primary actions in an opaque bottom dock; adapt insets, gaps and control sizes,
and reflow compact widths without losing focus or shrinking primary targets.
Auto width thresholds remain 600/1050 logical pixels for sizing. The owner's
follow-up replaces the six-column private hand with uniform 4:5 card frames in
horizontal overlapping suit groups, exposing at least 48 logical pixels per card.
Two suit columns may reflow to one with enlarged text; long groups scroll.
Public overlapping previews need an accessible quadrant control revealing full,
individually usable cards. Preserve live state, drafts, focus, game/command IDs,
privacy-safe projections, keyboard/touch inspection and exact rules.

Retain blue-charcoal surfaces, cgmsart, coral actions, cyan selection, lime focus,
fonts and vector suits. Faces belong only on J/Q/K; no decorative portraits or
casino props. The source mockups remain other-flow references; cgmsart-decks
remains artwork authority. The sole retained screenshot is the selected phone
reference; `client/screenshots` contains only that file, with no JSON/log archive.
New captures are temporary and must be removed after verification.
Do not overwrite that reference with a fresh widget render.

## Rules and implementation limits

Read [the design index](../../../docs/design/README.md), its canonical game rules,
accepted decisions and implementation/test register. Rules remain authoritative
for legality, timing, exact arithmetic and visibility; screenshots and this
skill are not a competing engine specification.

The online route connects to the Go authority; local demonstration routes remain
bounded deterministic scenarios. Browser evidence does not establish production
capacity, accessibility certification or native-device validation.
Keep integration boundaries in the [app contract](../../../client/INTEGRATION.md).

Preserve public hand/Ace counts without concealed identities or eligible-card
breakdowns; immediate Coup timing; action versus response/effect-input stages;
nonblocking proposals versus accepted transfers; success-only alliances; physical
card restrictions and persistent effect ownership; immutable game IDs; and
board closure versus settlement/finalization. Detailed retained safeguards live
in the gameplay and platform references.

## Read only what the task needs

| Work | Maintained reference |
|---|---|
| Theme, typography, card art, rights | [Foundations](references/foundations.md) |
| Public widgets, input and state semantics | [Components](references/components.md) |
| Gameplay presentation and hidden information | [Gameplay contract](references/gameplay-layout.md) |
| Flutter integration, navigation and connectivity | [Platform contract](references/flutter-platform.md) |
| Historical art observations and credited sources | [Research](references/research.md), not a current layout brief |
| Regression evidence and provenance | [Validation](references/validation.md) |

Compare actual rendered screens and interactions with the selected phone reference and shared hierarchy at each width after relevant
changes. Check art-visible and art-off states, exact identities, readable text,
focus and privacy. Report the evidence actually run and its limits. Keep this
entrypoint concise and update existing references rather than adding a parallel
visual specification. Follow repository review/tracking/staging rules.
