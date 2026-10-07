# cgmsart — four sibling decks

**Shared match styles, decision 0016 (2026-10-01).** First Light, Rain Glaze,
Ember Glaze and Drypoint are assigned distinctly across players for each entire
match, with the same public assignment for all viewers. Captured cards take
their new controlling player's style. These are automatic appearances, never
sold cosmetics or earned-currency unlocks. First Light remains the master art
edition and standalone/unassigned fallback; its earlier universal use in
[UI baseline v1](../UI_BASELINE.md) is historical.
Rain Glaze is the strongest subtle alternative: cool reflected strokes sit
naturally in the wet, nocturnal world. Ember Glaze is warmer and more theatrical;
Drypoint emphasizes printed material and is easiest to appreciate close up.

These are four cosmetic editions of **one 52-card design**, created on
2026-09-28. CGMS still uses **two decks without jokers**, as specified in the
[current rules](../../project/GAME_RULES.md#11-players-equipment-match-structure).
Editions do not determine physical copy numbers, gameplay ownership, rarity, powers, values or
permanent combo roles. Aces remain labelled **A**; the canonical rules permit a
declared numerical value of 2–10 for one attack or defense, independent of these
drawings.

## Open the artwork

Start with the [local review gallery](index.html), which includes edition and
artwork-only controls, all cards, comparisons, mobile previews and the shared
back. It works as a local HTML file; no external service or account is needed.

| Edition | Complete 52-card sheet | Readable suit sheets |
|---|---|---|
| First Light | [All 52](sheets/first-light-52.png) | [Hearts](sheets/first-light-hearts.png) · [Diamonds](sheets/first-light-diamonds.png) · [Clubs](sheets/first-light-clubs.png) · [Spades](sheets/first-light-spades.png) |
| Ember Glaze | [All 52](sheets/ember-glaze-52.png) | [Hearts](sheets/ember-glaze-hearts.png) · [Diamonds](sheets/ember-glaze-diamonds.png) · [Clubs](sheets/ember-glaze-clubs.png) · [Spades](sheets/ember-glaze-spades.png) |
| Rain Glaze | [All 52](sheets/rain-glaze-52.png) | [Hearts](sheets/rain-glaze-hearts.png) · [Diamonds](sheets/rain-glaze-diamonds.png) · [Clubs](sheets/rain-glaze-clubs.png) · [Spades](sheets/rain-glaze-spades.png) |
| Drypoint | [All 52](sheets/drypoint-52.png) | [Hearts](sheets/drypoint-hearts.png) · [Diamonds](sheets/drypoint-diamonds.png) · [Clubs](sheets/drypoint-clubs.png) · [Spades](sheets/drypoint-spades.png) |

All sheets use **Hearts, Diamonds, Clubs, Spades**, each ordered
**A, 2, 3, 4, 5, 6, 7, 8, 9, 10, J, Q, K**. Full sheets are
5592 × 2700 pixels; each suit sheet is 2800 × 2520 pixels.

- [Same-card comparison](previews/comparison.png): A Hearts, 8 Clubs,
  Q Hearts, K Spades and J Diamonds across all four editions.
- [390 × 844 mobile preview](previews/mobile-390.png).
- [One shared card back](shared/back.png).
- Individual faces and unlabelled native illustration crops are in
  `cards/<edition>/<suit>-<rank>/`; mobile exports are in
  `previews/mobile/<width>/<edition>/`. The gallery links each face.
- Separate rank/suit layers are in `overlays/`; the [manifest](manifest.json)
  maps all 208 cards to their exact source crop and shared overlay/back.

**Resolution disclosure:** the image tool returned each 13-painting atlas at
1448 × 1086, despite a larger requested canvas. Native illustrations are roughly
280 × 350 pixels after seam removal. The 720 × 1008 faces, contact sheets and
enlarged close-ups resample that artwork; their lettering is rendered at export
resolution. They are review exports, not native high-resolution print masters.
Native crops are retained so this distinction is inspectable. The back source
is 1060 × 1484. Mobile samples are 96 × 134 and 120 × 168; larger print or
high-density inspection production would need higher-detail painting masters.

## Variation matrix

| Element | Fixed across all editions | Edition treatment |
|---|---|---|
| Geometry | 720 × 1008; same cut upper-right corner, boundary and outer 20-pixel band | No variation |
| Identity | Same rank, suit, subject, portrait identity, pose and main composition | No intentional variation |
| Typography | Barlow Condensed Bold ranks; fixed corners, reading hierarchy and opaque masks | No variation; the same overlay file is reused |
| Suit marks | Original vector silhouettes; rose Hearts, amber Diamonds, turquoise Clubs, iris Spades | No variation |
| First Light | Original painted composition and pigment balance | Baseline |
| Ember Glaze | Same composition and dominant suit color | Small warm reflected-light strokes on existing object edges or coat folds |
| Rain Glaze | Same composition and dominant suit color | Small cool reflected-light strokes on existing shadow edges or coat seams |
| Drypoint | Same composition, lighting and dominant suit color | Sparse dry-pigment hatching inside object/garment areas |
| Hidden presentation | One shared indigo back and identical exterior | No rank, suit, copy or edition marking |

The intended 90–95% consistency is an art-direction judgment, not a pixel-diff
score. Generative edits can move small paint details; no measured percentage of
unchanged pixels is claimed. Every variant was generated directly from First
Light to avoid accumulating changes through successive edits.

## Subjects and layers

Hearts keep bent shelter roofs and a warm window. Diamonds retain fractured
beacons/glass and scaffolding. Clubs use dark batons/reeds beneath a severed cyan
tram wire. Spades alternate pointed ferry bows and worn press forms with violet
reflections. Each rank has an individual composition; objects are not pip counts
and introduce no rules. Faces appear only on **J, Q and K**.

Generated illustrations contain no rank lettering. Opaque label masks, original
suit vectors and text are separate PNG/SVG overlays. The flattened `face.png`
files are convenient review composites, not the required Flutter integration
format. A future Flutter renderer should use the native illustration and shared
geometry with live Barlow rank text and vector suit marks from authoritative
face-up state. Keep labels available when art is hidden or unavailable.

The supplied concept contained small back depictions but no standalone back
asset. The shared back redraws their indigo/copper aperture vocabulary as one
original reusable source. Existing references and the art-direction skill are
unchanged. Future concealed views must request only `shared/back.png`, use a
generic hidden-card label and omit private identity from URLs, semantics and
preloads. This art gallery is not a network/privacy implementation.

## Provenance and reproduction

[production.json](production.json) records all 17 built-in image-generation
calls, their prompts, reference roles, returned dimensions, output IDs and
source checksums. [atlas-crops.json](atlas-crops.json) pins inspected grid seams.
The [manifest](manifest.json) records per-card source boxes, native dimensions,
derived checksums and label/back references. Assembly crops and resizes the
paintings; it does not recolor or repaint them.

On 2026-09-29 the art identity and local paths were renamed to `cgmsart`.
Historical prompt text is preserved apart from that identity spelling; generation
output IDs, native source checksums and vendor font/license bytes are unchanged.

All illustrations were generated for this exploration using the cgmsart concept
as a style reference. The retained, rebranded [concept study](sources/cgmsart-concept.png)
now lives with source provenance; it is not a frontend layout reference. The
adaptive interface contract is defined by [UI baseline v2](../UI_BASELINE.md).
No third-party character, game image or stock artwork was
incorporated. Original assembly code and vectors follow the repository's
[MIT license](../../../LICENSE). Generated art is exploration material, with
generation-time production acceptance records retained. The later UI approval
selected First Light initially; decision 0016 now uses all four styles. Inclusion here does not claim exclusive
rights or third-party clearance. The unmodified Barlow Condensed and Atkinson
Hyperlegible Next binaries use **SIL OFL 1.1**; pinned upstream sources,
checksums and full notices are in [fonts/PROVENANCE.md](fonts/PROVENANCE.md).

From the repository root, with Pillow available:

```bash
pwd
bash xops/agent/safe-run.sh cgmsart-assemble -- python3 docs/design/cgmsart-decks/assemble.py
pwd
bash xops/agent/safe-run.sh cgmsart-verify -- python3 docs/design/cgmsart-decks/verify.py
```

Verification checks exact 4 × 52 completeness, both rank labels, suit colors and
contrast, overlay reuse, distinct source crops, source checksums, fixed edges,
shared silhouette/back and output dimensions. Visual review checks composition,
royal identity, face restrictions, cropping and mobile readability. These are
static artwork checks, not Flutter execution, user testing or print certification.
The historical unresolved Flutter-analysis breadcrumb concerns absent source and
is preserved; no runtime repair is claimed by this exploration.

**Recorded review:** all 16 atlases were visually inspected; no face on an ace or
number card, missing painting, generated lettering or changed royal identity was
found. Fresh artifact verification passed for all 208 cards. Rank contrast is
13.88:1; suit contrast ranges from 8.04:1 to 10.96:1. Independent static checks
passed for 53 Markdown links, 666 HTML/CSS references, 17 image-source hashes and
nine font/notice hashes. Browser checks exercised all four edition selections
and both directions of the artwork-only toggle successfully. Existing skill,
rules and previously staged gameplay images were confirmed unchanged.
