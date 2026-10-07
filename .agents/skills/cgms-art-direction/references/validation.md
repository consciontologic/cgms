# Frozen UI v1 validation and provenance

The [UI baseline](../../../../docs/design/UI_BASELINE.md) records the user's
2026-09-29 approval of the final Flutter web preview and the freeze contract.
This section records historical approval; the owner-selected 2026-10-02 phone
reference in the current baseline supersedes its layout restrictions. The
[20 main mockups](../../../../docs/design/mockup/README.md) remain supporting
storyboard evidence with their [manifest](../../../../docs/design/mockup/manifest.json).
Older concept layouts, worked screen arrangements, speculative widget dimensions
and archived generation prompts are retired from active guidance.

## Evidence and limits

Use the [app validation instructions](../../../../client/README.md#validate),
[current render evidence](../../../../client/README.md#rendered-board-evidence)
and source tests. At UI v1 approval, the implementation had 84 passing host tests,
50 passing package tests, clean analyzers, asset checks and a successful release
web build. Browser checks exercised phone/tablet/desktop flows. These are recorded
acceptance results, not tests rerun by this documentation alignment. The freeze
follow-up added a retired-navigation regression and passed 85 host tests and
50 package tests, as recorded in the app validation instructions linked above.

The real-host geometry regressions check the complete initial 18-card and opened
11-card hands with public cards at 390×844, 768×1024 and 1440×900, using actual regular/bold
fonts. Package captures use bundled card art/fonts/MaterialIcons and are real
Flutter widget renders; they are not full-host or native-device screenshots.
Do not claim every state or enlarged text setting fits without scrolling.

The local deterministic demo covers selected branches and interaction contracts.
UI approval does not certify all engine rules, production multiplayer/recovery,
accounts, screen readers/native devices, offline play or audience preference.
Use the app [integration contract](../../../../client/INTEGRATION.md)
and project roadmap for those remaining gates. Historical research measurements
and old concept contrast/media checks are not current app acceptance evidence.

## Regression criteria

- Preserve the baseline's theme, card art, typography, and implement the current shared vertical reference contract.
- Keep all physical identities present through opening, attack, transfer,
  purchase, result and reconnect. Do not duplicate combo constituents or conceal
  authorized identities from all reachable views.
- Check initial/opened phone/tablet/desktop fit with real host padding, action
  controls, seat status text and bundled fonts. Check short windows and enlarged
  text for reachable content and absence of horizontal clipping.
- Exercise card/pile inspection, physical-copy selection, keyboard I versus
  activation, focus restoration, buttons, menus, navigation, scores/history,
  disabled reasons, pending responses and uncertain reconnection.
- Compare exact legality, arithmetic, response/effect stages and privacy with
  canonical rules. Keep immediate Coup, immutable game IDs, public hand/Ace
  counts, nonblocking offers, success-only alliances and settlement boundaries.
- Inspect rendered art at intended size: quiet canvas, crisp opaque rank/suit
  masks, distinguishable ownership/selection, faces only on K/Q/J. Then inspect
  art-off fallback separately; it does not replace illustrated acceptance.
- Preserve reduced motion and semantics, evaluate final composited contrast,
  and report screen-reader/native-device checks only when actually performed.
- Keep original fonts/license notices, asset/source manifests and storyboard
  hashes. No design freeze changes third-party rights or production clearance.

Relevant tests remain in [the host](../../../../client/test) and
[the presentation package](../../../../client/packages/cgms_ui/test).
A future authorized UI revision updates the baseline version and its evidence;
v2 adaptive work preserves cgmsart while changing class-specific presentation.

## Retired concept provenance

The former skill concept PNG was moved byte-identically to
`docs/design/cgmsart-decks/sources/cgmsart-concept.png` solely because deck
production records depend on that source. It is not a frontend reference,
wireframe, shipping asset or alternate UI. Agents using this skill need not
inspect it. Source linkage and deck records are maintained in the
[deck provenance](../../../../docs/design/cgmsart-decks/README.md).

The concept was originally generated with the built-in imagegen tool on
2026-09-27, replacing a room study with a flat board; a correction removed
invented slogans/scenic margins and brought cards into one illustrated family.
No third-party image or local moodboard was supplied as generation input.
A 2026-09-28 edit reduced background saturation/texture while preserving the
foreground. A 2026-09-29 branding edit changed the label to `CGMS / cgmsart`,
generation output `exec-ac4876c3-787a-48b7-a26c-644e7c4ba062`. None of these edits
approved the concept layout or created additional asset rights.

| Historical source | SHA-256 |
|---|---|
| Retained concept after branding | `b783e11fde30a813d2c3c0e11e16630031e23e5d9a6d589f0147f7730268cfa1` |
| Branding-edit input | `3eeeca7e7f7774f16f0dd4cc1e8d15a55aecf165f128a833097c8bfe9e617d82` |
| Background-edit input | `15c5e19ea0df5884f319eb8e2807cb597e81b83ff23bf179da3972547e2aa681` |
| Earlier room/object study | `2b1558dca128d221742141bf57fdb982f2114b5b4ee85aded5f633689f781f3d` |

These hashes preserve provenance only; missing predecessors and the removed
moodboard/image-free specimen must not be reconstructed as active UI guidance.
Original prompt history remains in version history; imperative prompt wording
is not a present requirement. For credited external research, observation versus
interpretation and audience hypotheses, retain [research](research.md).
For asset grants, author credits, font licensing and production-status limits,
use [foundations](foundations.md) and the manifests/notices it links.

P04 browser acceptance follows the complete [v2 matrix](../../../../docs/design/UI_BASELINE.md#browser-testing-and-evidence),
including class transitions, touch/keyboard, RTL and browser screen readers.
Archive class-specific web appearance/journeys for native comparison. R06N retains
all native platform/device acceptance; this reference is not a second roadmap.
