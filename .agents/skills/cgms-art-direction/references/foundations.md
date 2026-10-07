# cgmsart foundations — shared vertical board

The [UI baseline](../../../../docs/design/UI_BASELINE.md) and current
[theme source](../../../../client/packages/cgms_ui/lib/src/theme.dart)
are authoritative. The former palette extensions, cut-ribbon geometry tables,
minimum card dimensions and speculative widget silhouettes are retired. Do not
apply them over the approved Flutter preview.

## Composition and material

Preserve the flat shared board, quiet blue-charcoal ground, distinct public
territories and vividly illustrated cards. Keep live rank/suit identities and meaningful state readable. Compact public
previews may overlap as in the owner-selected phone reference; private cards and
full inspectors retain individually accessible identities and artwork. Copy numbers remain semantic only. Coral identifies a primary action,
cyan selection and lime keyboard focus; these meanings also have text/shape
signals. Suit color never substitutes for rank, emblem or ownership.

Human faces belong only within K/Q/J cards, including their inspectors. Number
cards and Aces use environments, objects or abstract marks. Seats remain compact
names and numbers; no decorative portrait footprint, furniture, room perspective,
casino props or invented map movement. Preserve the existing card motifs rather
than commissioning a new deck merely to extend engine behavior.

## Implemented color tokens

These values mirror `CgmsColors`; change the source only within an authorized
baseline revision. `CgmsTheme` also defines its error/divider colors and concrete
control states. There is no separate proposed palette to implement.

| Token | Hex | Purpose |
|---|---|---|
| canvas | #101322 | Continuous quiet board background |
| surface | #1D2335 | Opaque reading areas and dialogs |
| raised | #2B344B | Neutral raised/disabled surfaces |
| text | #F5F1EA | Primary text |
| muted | #BAC4D6 | Secondary text |
| action | #FF795E | Primary action |
| cyan | #69D8E4 | Selection and secondary emphasis |
| lime | #DFFF7A | Keyboard focus |
| border | #8595B0 | Meaningful surface boundaries |
| heart | #F5A4BE | Heart suit |
| diamond | #FFD275 | Diamond suit |
| club | #83DCD3 | Club suit |
| spade | #B6B2FF | Spade suit |

Keep labels opaque and readable; never dim the entire scene to soften art.
Background texture must recede behind cards, ranks and controls. Test composited
text contrast and meaningful boundaries, rather than inferring accessibility
from swatches. The retained review targets are text >=4.5:1 and meaningful
boundaries/icons/focus >=3:1; see [WCAG text contrast](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html)
and [non-text contrast](https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html).
Visual approval is not accessibility certification.

## Typography and responsive geometry

The app bundles **Barlow Condensed Bold** for headings/ranks and **Atkinson
Hyperlegible Next Regular/Bold** for readable UI text. Use the existing theme,
not the former proposed weight/size table. Compact card ranks remain live 17px
text; other exact sizes and text growth are defined by the approved components.
Current sizes are historical geometry, not restrictions on v2 adaptive components.
The 2026-10-02 owner revision selects the compact phone reference at every width,
superseding the previous 100px minimum/no-overlap board requirement. Tablet and
desktop enlarge the same items rather than introducing separate panes. Public
piles expose a >=48px group control with full-card selection/inspection inside;
private hand cards preserve independent controls. Larger text may reflow. Keep
source artwork unchanged and full inspectors at native image ratios; validate
the 208 bundled PNGs against their source manifest.

Inherit text scaling, use growing/reflowing content and keep identities reachable.
Do not rasterize instructions or squeeze essential labels with `FittedBox`.
Preserve exact fractions/signs and physical copies. Test long names, Turkish
İ/ı/Ş/ş/Ğ/ğ/Ç/ç/Ö/ö/Ü/ü and localization when those surfaces change. Suit emblems
are vectors with semantic names, not dependence on a platform's suit glyph font.

## Shared match card styles

Decision 0016 assigns the four existing cgmsart styles distinctly across seats
for an entire match. All authorized viewers resolve visible cards from the same
public player assignment and authoritative current controller, so captures
adopt the capturing player's style. A new match gets fresh assignments; reload,
reconnect and subsequent games within a match preserve them. No style is sold
or unlocked. Keep one identical concealed back; First Light is the unassigned
or standalone-demo fallback. Appearance never determines card identity or rules.

## Asset provenance and rights

Use the [runtime asset manifest](../../../../client/packages/cgms_ui/assets/manifest.json),
[asset notice](../../../../client/packages/cgms_ui/assets/NOTICE.md),
and [font provenance](../../../../client/packages/cgms_ui/assets/fonts/PROVENANCE.md).
Font binaries, pinned sources, checksums, original author/contributor credits and
SIL OFL 1.1 notices are retained with the bundled fonts. The repository's MIT
license does not replace their licenses. Retain names because these binaries
are unchanged; honor Reserved Font Names if authorized modifications occur.
No system font installation or runtime font service is needed.

The deck motifs are original cgmsart interpretations: amber fractured beacon for
Diamonds, rain-dark batons/cable for Clubs, bent shelter roofs for Hearts,
violet press/ferry-bow form for Spades and expressive original royal portraits.
All hidden cards share the same identity-neutral back. Art cannot encode hidden
rank, suit, copy, role or eligibility through its asset choice or variations.

For any authorized new asset, record creator/generation method, source, usage
grant, modifications, attribution, checksum and shipping status in the existing
manifest. Visual acceptance does not create third-party rights or by itself
clear production distribution. Preserve original masks/foreground sources when
available and test the actual display crop. Failed art keeps live identity and
state; decorative art is excluded from semantics when adjacent labels duplicate it.

[Credited research](research.md) supplies inspiration only. Public portfolios,
source-game screenshots and the retired moodboard grant no reuse permission.
Do not trace portraits, characters, logos, tarot labels or another game's UI.
The obsolete generated concept is retained only as deck-generation source
provenance. Its historical hashes/method are recorded in
[validation](validation.md), not as a shipping asset or layout brief.
