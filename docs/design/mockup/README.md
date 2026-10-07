# cgmsart mobile flow mockups

[UI baseline v2](../UI_BASELINE.md) governs adaptive UI/UX. The owner reaffirmed
this unchanged set as the primary layout, hierarchy, navigation and interaction
reference on 2026-10-01. Use [cgmsart-decks](../cgmsart-decks/README.md) as the
primary card-art reference: its illustrations replace the art shown here while
the mockups retain their layout and UX direction. Deliberately adapt static
arrangements to current rules, privacy and purpose-built device requirements.

These 20 mobile mockups were approved on 2026-09-29 as the visual basis for the
[Flutter demonstration](../../../client/README.md). The main action board keeps all four
players' public cards and the complete local hand together. Each public number
series or combo forms a compact vertical stack with every rank and suit exposed;
the private hand uses a flat grid. Tablet and desktop layouts retain the same
information and art, with more space for cards and action details.

The PNGs were imported unchanged from the approved generation outputs.
[`manifest.json`](manifest.json) records their generation IDs and SHA-256 hashes.
They are visual references, not a rules engine or runtime accessibility evidence.
The [current rules](../../project/GAME_RULES.md) govern interactions where a
static image cannot express timing, visibility or all possible states.

| Screen | Flow |
|---|---|
| [01](01-room-and-match-setup.png) | Room and match setup |
| [02](02-initial-deal-and-turn-order.png) | Initial deal and fixed seat order |
| [03](03-compact-main-board.png) | Full table and 18-card hand |
| [04](04-inspect-underground.png) | Inspect an opened Underground combo |
| [05](05-select-hearts-to-open.png) | Select all seven number Hearts |
| [06](06-hearts-opened.png) | Open Hearts; hand reduces to 11 |
| [07](07-select-an-exact-attack.png) | Select 8 Clubs against Bob's 8 Hearts |
| [08](08-wait-for-attack-responses.png) | Wait for explicit seat-ordered responses |
| [09](09-attack-resolved.png) | Resolve the attack; used Club remains public |
| [10](10-draft-a-private-trade.png) | Draft a private card trade |
| [11](11-offer-pending-during-play.png) | Keep playing while an offer is pending |
| [12](12-trade-completed.png) | Complete the accepted trade |
| [13](13-choose-a-card-purchase.png) | Choose quantity and Spade payment |
| [14](14-purchase-resolved.png) | Resolve the purchase and show the new hand |
| [15](15-successful-combo-loan-and-alliance.png) | Successful combo loan and alliance |
| [16](16-use-confinement-to-stop-ponzi.png) | Confinement cancels Ponzi |
| [17](17-immediate-coup-declaration.png) | Declare an eligible Coup immediately |
| [18](18-coup-result-and-settlement.png) | Board closed; settlement still pending |
| [19](19-finalized-game-and-match-standings.png) | Finalized match standings and next game |
| [20](20-reconnect-without-replaying-an-action.png) | Reconcile an unknown operation outcome |

Screens 03–14 follow a common opening, attack, trade and purchase sequence.
Screens 15–19 demonstrate independent loan, Ace and Coup branches rather than
additional actions in that same deal. Screen 20 branches at an attack submission
whose acknowledgement was lost. Opponent hand faces remain concealed throughout.

Preserve the board/hand relationship through readily accessible public context,
turn status and actions. V2 authorizes phone Table/Hand/Decisions navigation and
the tablet companion pane; the original one-board arrangement does not override
those device requirements. At larger text sizes or unusually dense states,
readable scrolling takes precedence over clipping cards or text.
