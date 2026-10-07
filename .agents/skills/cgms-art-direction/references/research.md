# Historical research record and decision basis

**Status:** retained provenance and research history, not a current UI brief.
The [current adaptive UI baseline](../../../../docs/design/UI_BASELINE.md) and approved
Flutter app supersede former layout interpretations and concept proposals.
The 20 retained mobile mockups are supporting storyboard evidence. No external
sources were re-fetched for the freeze alignment; dated observations and rights
limits below remain historical claims at their original scope.

Research and visual inspection: **2026-09-27**. This is desk research for an
original CGMS identity, **cgmsart**, not evidence from CGMS user testing.
The user refined background restraint on **2026-09-28**; that historical decision
is recorded below separately from the source observations. The subsequent UI v1
approval is recorded in the baseline, not inferred from this research.

Read this reference when commissioning art, reviewing visual quality, or
revisiting a design decision. Exact tokens belong in [foundations](foundations.md),
interaction contracts in [components](components.md), and game constraints in
[gameplay/layout](gameplay-layout.md). This record explains their basis.

## Evidence labels and method

- **Observation:** what is visible in an inspected image or footage sample.
- **Creator statement:** an attributed first-person explanation; it does not prove a UX outcome.
- **Interpretation:** an original CGMS design decision drawn from those sources.
- **Hypothesis:** a proposed audience or usability outcome that requires testing.

Actual web searches included “site.discoelysium.com devblog Aleksander Rostov
art colour interview Indiegraze”, “Kaspar Tamsalu Disco Elysium Old friends
Artstation”, and “Aleksander Rostov Disco Elysium portfolio artstation”. Source
pages were opened; the credited artwork below was enlarged in a browser, not
inferred from search snippets. The local 2373×5000 image was opened in full and
visually inspected through a scaled preview. It was the principal aesthetic
reference for this research, not evidence of the game's interface.

## Principal local reference: observation → original translation

The former local contact sheet, `docs/design/theme-example.jpeg`, was inspected
before the user removed it from the checkout. It contained 22 portrait panels:
five rows of four and a final centered pair. Small white
uppercase tarot labels occupy black gutters. Its compiler, image permissions
and exact typeface are unverified. Do not ship the sheet or copy its figures.

| Aspect | Direct observation | cgmsart interpretation |
|---|---|---|
| Brushwork | Angular paint planes, broken contours, scraping-like marks and linear hatching; eyes and mouths receive concentrated detail | Apply emotional specificity to King/Queen/Jack portraits only; elsewhere give objects and architecture interrupted silhouettes. Use broad paint masses at card scale and concentrate detail at one focal point |
| Color | Black gutters and cool gray/green shadows sit beside emphatic yellow, orange, blue and cyan; isolated pink/violet passages change the emotional temperature | Give focal artwork a hot/cool conflict, supported by muted broad fields and calm reading areas. Preserve chromatic character in cards and meaningful accents while the backdrop recedes |
| Composition | Tight crops, off-center faces, changing gaze, diagonals, broad vertical strokes and several unusually abstract portraits | Let environmental planes, royal card art, suit emblems and action surfaces have different masses and silhouettes. Guide the eye toward the active seat or decision without allocating portrait space |
| Framing | Strong black intervals separate intense portraits, while the paintings themselves vary sharply | Use negative space to separate meaning. A card frame can be cut, painted or overprinted; repeating identical dashboard panels is not a required translation of this sheet |
| Texture | Visible underlayers, scumbled colors and scraped-looking patches occur throughout the pictures, not only at their margins | Give card faces and action seals a tactile character; keep supporting board texture sparse and low contrast. Values, labels and instructions retain stable opaque masks |
| Typography | Narrow uppercase titles are subordinate to the images; the sheet is not a reading interface | Use striking display typography sparingly for moments and roles, with generously sized functional text. Do not copy tiny condensed captions into gameplay |

These observations identify visible qualities, not the physical medium, source
software, author of every portrait, or any reuse license. The sheet's tarot
names do not create CGMS card roles or mechanics.

## Credited artwork and creator accounts

| Source opened | What was actually inspected | Evidence and useful decision |
|---|---|---|
| [Kaspar Tamsalu — Old friends](https://kaspartamsalu.artstation.com/projects/nELn11?album_id=82471) | Artist's attribution text and enlarged René portrait | **Creator statement:** Tamsalu designed René/Gaston and their portraits; Rostov adjusted them before use. **Observation:** cool one-sided facial shadow, bright ochre/teal uniform planes, violet-gray background fragments. **Interpretation:** character contrast should emerge from silhouette, attitude and color relationships, not palette swaps |
| [Aleksander Rostov — Disco Elysium Archetypes](https://rostovjanka.artstation.com/projects/6aAL8x) | Artist attribution, process caption and enlarged three-portrait composite | **Observation:** curled thinking pose against a cyan circular field, reclined figure amid violet shards, forward-driving figure with hot red diagonal strokes. **Interpretation:** vary gesture and directional marks across CGMS roles; reserve energetic shape changes for meaningful focal differences |
| [Aleksander Rostov — Disco Elysium Village](https://rostovjanka.artstation.com/projects/QERm4) | Enlarged environment concept | **Observation:** structures cluster left/center; open cold water occupies the right; jetty and shore provide diagonals; small blue, red and yellow material accents interrupt gray-green surroundings. **Interpretation:** atmosphere needs large calm masses and selective concentrated detail, not uniform noise |
| [ZA/UM — Indiegraze interview](https://discoelysium.com/devblog/2017/08/02/intriguing-indigraze-interview) | Read the interview, published 2017 with input from Rostov and Kurvitz | **Creator statement:** worldbuilding and psychology inform fragmented color, including complementary hues inside apparently simple surfaces. **Interpretation:** let original CGMS artwork express tension and conflicting motives through chromatic variation; keep semantic UI colors consistent |
| [Rostov — Village Concept Art](https://discoelysium.com/devblog/2016/06/29/fishing-village-concept-art) | Read the 2016 process account; the separate portfolio above supplied the visually inspected final image | **Creator statement:** thumbnails connect composition and movement; asymmetric balance creates tension; shape layers constrain textured brushwork. **Interpretation:** thumbnail the gameplay focal hierarchy together with the illustration, then preserve readable shapes while painting |

The portfolios identify Rostov as Disco Elysium's art director and Tamsalu as a
contributing character artist through his own work account. This records credits
on the inspected pages, not a claim about anyone's present employment. Their
portfolio pages state **all rights reserved**. Their artwork is a research
reference and supplies no permission to reproduce it.

## Actual game interface: separate evidence

The coordinating researcher inspected the official Steam screenshot gallery
and played the trailer, then separately inspected a player recording. These
are observations of the actual game interface; portfolio paintings and the
official website's navigation were not treated as interface evidence.

| Source and inspected sample | Direct observation | Evidence boundary |
|---|---|---|
| [Official Steam gallery](https://store.steampowered.com/app/632470/Disco_Elysium__The_Final_Cut/) — world screenshot, gallery #0 | The isometric environment dominates; painted party portraits sit lower-left, with small bottom navigation | A still establishes visible composition, not how navigation behaves |
| Same official gallery — skills screenshot, #1 | Expressive dark painted tiles carry selective purple/red/yellow; white selection frame, large left attribute numbers, focused detail at right and orange Accept strip | Shows an illustrated information system with a legible selected state; no measured contrast or input result |
| Same official gallery — inventory screenshot, #2 | A central full figure sits on a mustard/cream painted field; clothing slots surround it, with compact item grid/detail right and dark stats left | Figure, equipment and information have distinct spatial roles; does not justify copying the layout |
| Same official page — 59-second trailer | Environment and title frames inspected during playback | Marketing footage sample, not a full playthrough or proof of interaction timing; gallery order may change |
| [Indie James player recording](https://www.youtube.com/watch?v=H_pS14_v6po) — 00:15 | Near-black world, narrow right text column, speaker text above an orange Continue bar, and a small character icon at right | Third-party player footage, **not** an official/developer source. YouTube ads interrupted later seeking; playback subsequently resumed |
| Same player recording — 08:47 and 08:54 | A huge painterly mirror portrait dominates the left beside a bounded dark dialogue column. At 08:47 a green CHECK SUCCESS strip and dice occupy the foreground; at 08:54 that overlay is gone while result prose and a numbered orange Okay choice remain | A sampled transition, not a full walkthrough or measured animation duration; no inference about unseen intervening frames |

**CGMS interpretation:** give the next decision a bounded, readable text zone
within a composition whose illustrated objects carry identity and state. A
selected card can expand into an inspector, but its ownership, rank and legal
action must already be readable. Quiet text fields can coexist with active
painted silhouettes; they do not require uniform panels across the screen.
Expressive feedback can recede while its result stays readable. Preserve the
CGMS result in text after an effect ends; do not import dice, random checks or
the mirror composition from the footage.

These samples do not establish Disco Elysium's font family, touch behavior,
keyboard support, accessibility compliance or exact motion durations. They
also do not validate a right-side reading column for CGMS: its card table,
simultaneous public/private information and immediate declarations need their
own responsive composition. No copied UI arrangement is prescribed here.

## Historical identity decisions and interpretations

**cgmsart** expresses CGMS's visible power, private intentions, risky deals
and surviving obligations through painted board territories and card objects,
fractured suit marks, unequal frames and saturated traces. This is an original
art direction, not new game lore or a change to its rules. Read the current game
document from [the design index](../../../../docs/design/README.md); the third
draft was current during this research.

**User-approved refinements, 2026-09-27:** retain the card design, painted
backgrounds, colors and overall cgmsart approach. Human faces belong only on
King, Queen and Jack cards and their inspectors. Replace the proposed player
portraits with compact functional seat identity. The subsequent refinement also
rejects the physical room, table, chairs and decorative props: use a common 2D
game board where gameplay, scores, deals, alliances, combos and flow are visible.
Transfer the approved painterly material and colors into board territories,
card art, relationship brackets, event marks and controls themselves.
This is a product decision from the user, not an inference from Disco Elysium or
an audience study. The source observations above remain accurate research records;
portrait-specific interpretations apply only to royal cards. Objects must not
inherit the old portrait footprint or imply possessions, abilities or resources.
The board is an information canvas, not a map with invented movement mechanics.

**User-directed refinement, 2026-09-28:** the reference image's broad background
color is overused and makes objects harder to discern. Retain cgmsart's
painted character, card motifs and expressive controls while reducing backdrop
saturation, luminance variation and texture density. Cards, items, functional
board groups and live letters/numbers must stand forward. Keep rich color in
focal card art and meaningful action/state accents; use soft blue-gray/indigo
fields with faint iris undertones behind them. This supersedes any earlier
instruction to spread vivid pigment or strong texture across atmospheric areas.
It was a direct design request, not a measured usability result. That reference
remained an art study; its arrangement was never the approved UI. The frozen
Flutter baseline supersedes the layout interpretations below.

- **Illustration is structural.** A selected card's art, silhouette and suit mark
  should lead the eye into its readable rank and action state. An active
  seat's public state, a trade's opposing sides, and an impending declaration
  deserve compositions shaped for that purpose, not one background behind
  conventional controls.
- **The interface carries the painting.** Buttons are tactile action seals,
  counters are distinctive inset instruments, card frames have controlled cuts
  and overprints, and icons have family resemblance to the suit marks. Their
  text, hit areas and focus states remain dependable. This is a design
  interpretation, not an imitation of Disco Elysium's controls.
- **Background restraint supports foreground character.** Large muted fields
  make concentrated coral, cyan and violet in cards and meaningful controls
  easier to distinguish. Selective background desaturation is intentional;
  preserve strong card silhouettes and readable values. Saturated accents need
  a compositional or semantic job; outlining everything in neon removes focus.
- **Complexity is organized, not removed.** Current suit openings, opening
  history, physical-card identity, ownership, available points, recorded score
  and debt need separate readable representations because the CGMS rules make
  them different. An expressive illustration cannot substitute for a label or
  reveal an opponent's hidden information.
- **Readable fallback is a separate gate.** Hide art to check essential state,
  ownership and legal actions, then restore art to judge the intended work.
  Passing the hidden-art check does not establish artistic success and is not
  a reason to make illustration incidental. A generic dashboard still fails.

Do not trace the source portraits, reproduce signature characters/costumes,
reuse logos or tarot labels, recreate the game's exact interface compositions,
or present third-party art as a CGMS asset. Commission original art or use
assets with a verified license suitable for the intended distribution; record
creator, source, license, attribution and modifications before shipping.
Specific font licensing is recorded separately in [foundations](foundations.md).
No outside image was added to the project during this research.

## Audience and usability hypotheses, not demographic claims

**Untested audience hypothesis:** expressive, strange, emotionally specific
imagery and understandable decisions may make CGMS memorable and appealing to
some Gen Z players. No consulted source establishes this for CGMS or for a whole
generation. Age alone does not imply a preference for dark imagery, slang,
irony, neon or rapid motion. Avoid pressure loops, false scarcity and rewards
that obscure the rules.

For a future evaluation, recruit players with varied card-game experience,
including people in the intended age range, keyboard-only players and people
who use large text. Show art-visible compositions first: ask what feels
memorable, what emotional tone they perceive, what they would interact with
next, and whether this feels like a distinctive illustrated game. Do not lead
with the Disco Elysium comparison. A beautiful portfolio image does not prove a
usable game screen, and preference ratings alone do not establish comprehension.

Then test concrete tasks: identify attack eligibility; select one of two
identical clubs without confusing physical-card state; explain available points
versus debt; respond in seat order; find an immediately legal Coup; and recover
from a disconnect. Record correctness, mistaken actions, time-to-find and the
participant's explanation. Use the hidden-art version only for the semantic
fallback tasks. Check whether focal illustration improves or competes with the
next legal decision before claiming it does either.

No participant study, Flutter runtime evaluation, input test or PWA offline test
was conducted during this historical research. Later Flutter validation is
recorded separately in [validation](validation.md) and the UI baseline; it does
not retroactively turn these observations into user-study evidence.
