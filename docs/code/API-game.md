# API: CGMS online game v1

## Overview

Implemented JSON HTTP commands/queries and authenticated WebSocket delivery in
`backend/internal/transport`, using the shared engine and PostgreSQL matchstore.
The machine-readable [OpenAPI contract](../api/online.openapi.json) defines the
wire names; examples below use those names. This is a single-process local
runtime, not a deployed or production-capacity claim. The [Flutter client](../../client/README.md) integrates these routes.
Economy activation and its current local evidence are described below.

## Authentication and retention decision (2026-09-30)

`POST /v1/guests` issues a guest identity. `/v1/register` and `/v1/login` use
local lowercase ASCII usernames (3–32 letters/digits/underscore) and passwords
(15–128 UTF-8 bytes, at least 15 characters). Passwords use salted Argon2id
(19 MiB, two iterations, one lane), with at most two concurrent hash jobs.
No external identity provider, account merge or password recovery is implemented.
`POST /v1/upgrade` preserves the guest's actor ID, assigns credentials and
revokes its old sessions. `POST /v1/logout` revokes one session;
`DELETE /v1/account` revokes all sessions and removes credentials/profile data.
Deletion retains the pseudonymous actor, rooms, shared match history, exact
financial obligations and private authoritative journal indefinitely. It never
manufactures an in-game departure, refusal or timeout. This resolves the runtime
retention conflict with untimed games; it is not a legal/privacy-release assessment.

Opaque 256-bit session tokens expire after 24 hours and are stored only as
hashes. Native requests send `Authorization: Bearer …`; browser requests with
the configured exact Origin receive an HttpOnly, Secure, SameSite=Strict cookie
and a CSRF token. Cookie mutations require that Origin and `X-CSRF-Token`.
Bearer+cookie ambiguity rejects. `/v1/session` returns the authenticated account
and CSRF token. WebSocket upgrades authorize membership and recheck session and
membership before every bounded read. Revocation affects subsequent checks;
a command already admitted to an atomic transaction may finish. Tokens never
belong in URLs, event bodies or logs.

## Running the local authority

From the repository root, set `CGMS_DATABASE_URL` through the local secret
environment and `CGMS_ORIGIN` to the exact browser origin, then run:

```bash
CGMS_RULES_FILE="$PWD/config/rules/game-rules.json" go -C backend run ./cmd/server
```

The standalone binary defaults are loopback
`127.0.0.1:8080` and `config/rules/game-rules.json`; `CGMS_LISTEN` and
`CGMS_RULES_FILE` override them. The rules file must match the accepted canonical
hash. `CGMS_REDIS_URL` optionally configures private Redis hints.
`CGMS_INSECURE_LOCAL=1` permits HTTP origins and non-Secure cookies. It must
only be used with loopback binds/origins; the flag does not enforce that boundary. TLS/reverse-proxy deployment is a later release gate.
Startup runs additive migrations and loads a persistent private handle key from
`online_settings`. Backups must include it; changing it invalidates outstanding
handles. A held PostgreSQL advisory lock rejects accidental second startup;
connection loss stops the server. This is not distributed fencing or hot failover.
Run one authority process and stop the old process before restarting elsewhere.

## Endpoints / methods

These routes are implemented. Session routes are described above; room invitations
are issued separately by the owner and expire after 24 hours. Issuing another
invitation rotates the prior token; only its hash is stored.

| Method | Contract |
|---|---|
| `POST /v1/rooms` | Create a room; `(authenticated_account_id, command_id)` identifies the operation before the room exists. Persist its body fingerprint, generated room and result atomically; identical retries return that room. |
| `GET /v1/rooms/{room}` | Read room/membership as a member. |
| `POST /v1/rooms/{room}/invitations` | Owner issues/rotates an invitation; response is private. |
| `POST /v1/rooms/{room}/join` | Accept an invitation/seat under room capacity and membership constraints. |
| `POST /v1/rooms/{room}/matches` | Start a pinned-rules match; validate three/four seats and a visible positive fixed game count (default three); the technical maximum is 100. Creation uses `games_per_match`, room responses expose `games`. Keep this count immutable after start. Enabled economy reserves each human seat's allowance in the same transaction; exhaustion rejects without a partial match or charge. |
| `GET /v1/matches/{match}` | Authorized seat snapshot and matching event cursor. Never return an opponent's hidden cards. |
| `POST /v1/matches/{match}/commands` | Submit one typed intent; durable result includes `version`, also its resulting event cursor. |
| `GET /v1/matches/{match}/commands/{id}` | Reconcile the authenticated actor's operation on the primary; absence is not proof an earlier request cannot still commit. |
| `GET /v1/matches/{match}/events?after={seq}` | Bounded catch-up of authorized envelopes; expired cursor requires a fresh snapshot. |
| `GET /v1/matches/{match}/stream?after={seq}` | WebSocket upgrade; client presents last applied cursor, receives replay then live envelopes. |
| `GET /v1/economy` | Authenticated caller's benefits, free starts remaining and next UTC reset; returns only `enabled: false` when disabled. |
| `POST /v1/economy/passes` | Spend earned dirt for Daily (`days: 1`) or Weekly (`days: 7`) using immutable `{operation_id, days}`; returns the original cost/absolute expiry on retry. Retired two–six-day receipts remain recoverable, but cannot be newly purchased. |
| `GET /v1/economy/passes/{operation}` | Retrieve only the authenticated caller's pass receipt; 404 means currently absent, not proof an in-flight request cannot commit. |

### Online bot opponents

Room creation accepts optional `bot_difficulty`: `beginner`, `standard` or
`advanced`. Omission retains a human-only room. A bot room places its human owner
at seat 1 and fills the other seats with server-owned identities. The room
response includes `bot_difficulty`, and bot members have `bot: true`; the owner
still starts through the existing idempotent match-start operation.
Authorized match snapshots expose public `projection.online.bot_seats` and
`bot_difficulty` for opponent labels. These fields contain no hidden cards or
policy input; human-only matches omit them.

Bot decisions reuse the existing fair policy tiers and pass through the same
authoritative command, journal and projection boundaries. A bot may see its own
private cards and authorized public information. It cannot act for the human or
resolve a human response by timeout. Bot identities have no credentials or
sessions and do not receive spendable economy rewards; human admission and
reward rules remain unchanged. Durable membership and command identities allow
the runtime to resume bot work after reconnect or server restart.

Online suggestions exclude purchases whose drawn quantity is no greater than
the number of payment cards. This prevents the observed empty-deck cycle of
buying back and reopening the same cards, using only the bot's authorized menu.
It changes neither human purchase legality nor the frozen offline/evaluation
policies, and it is not a general proof of game termination.

Online planning ranks candidates before lifecycle validation, validating the
selected candidate and trying the next ranked choice only if necessary.
Ambiguous legacy ordering keys retain eager validation to preserve tied choices.
Planning observes cancellation; the worker's two-second operation deadline is
unchanged. Frozen offline menu generation and policy behavior remain intact.

At round closure, `projection.online.round_closed` reports the public boundary.
The viewer-specific `departure_pending` flag indicates that viewer still needs
to submit `departure-choice` (`accept: false` to stay, `true` to leave). The
client opens this decision and waits for an explicit human choice; other
players' departure choices remain private. Absent flags mean false.
In a bot match, automatic Stay choices wait for the human's explicit answer,
keeping the version stable while the human responds. All-human rooms retain
their simultaneous choice flow; immediate Coup eligibility remains authoritative.
`projection.online.turn_started` records whether the scheduled turn's draw
attempt has completed. A contiguous live false-to-true transition for the same
game/turn/actor, accompanied by exactly one added public hand or concealed-Ace
count, can drive ephemeral draw feedback. An exhausted attempt adds no card.
Snapshots and reconnect establish a baseline; they do not replay feedback.
The field exposes no physical identity or remaining supply and is optional on
historical projections. It does not grant a client draw command.

`projection.board.own_bindings` preserves fixed Doppelganger roles after a
formation is taken back or its kings separate. Each entry contains `kings`
(one or two current own-card handles) and `slot` (`rank`, `suit`). A counterpart
controlled elsewhere is omitted, including its identity and location. The
entry never mints a capability or exposes another player's concealed card;
its handles reuse the viewer's authorized current cards. One visible king
still cannot be attached, reassigned or paired with a different king. A full
owned pair can be reopened only with its original slot. Formation commands
continue to require exactly two substitute kings. This additive metadata is
omitted when empty and may be absent from older projections. Historical
verification tolerates only absence; a present changed pair, slot, null or
duplicate field remains invalid. Engine binding rules and state are unchanged.

`projection.online.server_pending` is true when the persisted state has a
server-owned transition ready, such as deal, initiative, turn start or automatic
continuation. The client waits for it to clear before enabling ordinary commands;
a dealt hand alone is not turn readiness. `automatic_pending` keeps its existing
financial-work meaning, and eligible immediate Coup remains governed by authority.
Upgraded servers derive missing round and server-readiness flags for an authorized
snapshot from its matching persisted checkpoint. Historical receipts remain
immutable: verification tolerates only joint absence of `round_closed` and
`departure_pending` and, independently, absence of `server_pending` or
`turn_started`; all existing
content and present flag values remain exact.

These difficulty names identify existing policies, not verified strength levels.
P03 evaluation and native offline acceptance remain separate gates.

### Shared match card styles

Online `projection.board.players[]` also includes public `card_style`, one of
`first-light`, `rain-glaze`, `ember-glaze` or `drypoint`. Styles are distinct
across the three/four seats and stable for the entire immutable match ID,
including later games, journal replay and reconnect. The v1 assignment hashes
only the public match ID and edition in a separate namespace; it never uses
private random seeds or viewer/card handles. Clients use the current card
controller's style for authorized visible cards, including after capture.
Stored gameplay observations, checkpoint schemas and replay digests are unchanged.

### Local economy

`CGMS_ECONOMY_ENABLED=1`, or `economy_enabled: true` in the dedicated server
configuration, enables decision 0014's 10 dirt per finalized online game and
30 dirt per pass day. Omission disables economy for new matches. Existing
enabled matches retain their pinned policy on restart. No live store purchase,
promotion-redemption or client-reported outcome route is installed.

The account response is `{enabled, benefits, free_starts_remaining, resets_at,
reward, pass_day_price}`. `benefits` contains `dirt`, `pass_until`,
`premium_until`, `unlimited`, `no_ads`, `ad_free_until` and `cosmetics`; it never includes provider
receipts, another account's holdings or promotion inventory. Free starts reset
at 00:00 UTC. Pass/premium-covered starts retain receipts but consume no free
allowance. Decision 0015 separates unlimited access from ad removal: new earned
Daily/Weekly access retains ads unless a separate ad-free benefit is active.
The private `ad_free_until` reports the effective expiry of verified paid or
grandfathered ad-free access; standalone Ad-free does not grant unlimited starts.
Decision 0018 sets standalone Ad-free to 30 days and paid Daily to a non-renewing
24-hour unlimited/ad-free pass from confirmed purchase. Paid Weekly, Monthly
and Yearly also grant both benefits. Trusted confirmation metadata remains
inside provider reconciliation; it is not a public request field. Restores and
notifications preserve the original fixed-term expiry rather than starting a
new duration when processed.
When there is no ad-free source, the value is epoch. Clients use the authority's `unlimited` and
`no_ads` flags separately rather than deriving eligibility from local clocks.
Older responses may omit the new expiry; clients omit that explanation without
inventing an expiry or changing either flag.
The additive `legacy_ad_free_until` benefit preserves the original promised
expiry and is never extended by new dirt redemption. No cosmetics can be newly
granted or bought. `PRODUCT_UNAVAILABLE` (409) rejects a retired duration after
checking for an existing receipt with that operation identity.

Pass receipts contain `operation_id`, `days`, `cost` and `expires_at`. Persist
the operation ID/duration before sending. On `OUTCOME_UNKNOWN`, reconcile by GET
or retry exactly the same body; a changed duration yields `ECONOMY_CONFLICT`.
`INSUFFICIENT_DIRT`, `ECONOMY_DISABLED` and `ALLOWANCE_EXHAUSTED` are distinct
409 errors. A successful pass wakes automatic next-game admission; an exhausted
allowance never changes the completed game, awards unfinished play or produces
a timeout. Native/store acceptance remains R06N/R06.

Snapshot responses may include top-level `admission_waiting: true` when an
economy-enabled completed/nullified game has a successor but at least one seat
cannot start yet. This reveals no account identity or balance. It is a current
admission check, so refresh snapshots after redemption or UTC reset even when
the durable game cursor has not changed; historical event payloads omit it.

For browser QA only, `python3 xops/postgres_integration.py --webfixture-economy`
runs the loopback fixture with approved amounts. Every synthetic fixture account
receives three prior-day finalized-ledger reward fixtures through the economy
transaction hooks (30 dirt), plus its current game's allowance charge. This is
seeded QA evidence, not a natural full-game playthrough or production reward API.

Command request shape (all five outer fields are required):

```json
{
  "command_id": "opaque-client-operation-id",
  "game_id": "immutable-origin-game-id",
  "expected_version": 42,
  "type": "decision",
  "payload": {
    "pending_action_id": "opaque-action-id",
    "decision_id": "opaque-decision-id",
    "selection": ["authorized-card-handle"]
  }
}
```

Every game-scoped intent, including victory, response/effect input and financial
settlement, requires the immutable `game_id` created by the authority. It is
part of the command body fingerprint. Authenticate/authorize and look up the
durable command result first: an identical committed command returns its original
result even after its game closes. A new command must name the current game
instance before any current-state eligibility check; otherwise return
`STATE_CONFLICT`. Never substitute the current game for an old queued
intent. Bind turn, response-window, action, effect and decision IDs where that
command depends on them, all belonging to the named game. A next-game lifecycle
command names its finalized predecessor; a successful transition returns a fresh
server-generated game ID. Nullifying a game also gives its replacement a fresh
ID even though the configured match slot is reused. The match-start operation
creates the first game ID; room/match-management commands have their own scope.

Possible command families: open/add series, open/reconfigure/withdraw combo,
attack, activate card, respond/pass, create/revise/withdraw/decline/accept a trade
or loan proposal, purchase cards, declare victory,
end turn and request permitted quit/resignation. `decision` supplies an
input to an accepted effect; it is not a second response opportunity. A pending
record identifies effect, decision, required actor, kind, stage and that actor's
authorized options. Derive the actor from authentication, validate every handle
against the stored decision, and reject stale or wrong-actor inputs. Accepted
activations cannot be withdrawn; reconnect returns their same decision and any
already committed random result. Current selected-card handles bind the view revision; pending/decision IDs bind responses.
Offer acceptance names its exact revision; the server validates its stored origin turn.
Each command has an explicit schema,
phase/support/usage checks and a rule reference. A trade/loan acceptance validates
its exact revision and starts an accepted action; the eventual transfer has its
own final revalidation and atomic resolution under the contracts below. Social
reward-sharing promises remain voluntary transfers. Record an explicitly accepted
recipient, exact amount or reward percentage, origin `game_id`, named `award_id`
and typed condition. In v1 the condition is occurrence of that named award for
that payer in that game; distinguish threshold bonus, Coup, Ringleader and
ordinary-ending awards. Free text is commentary, never a server-interpreted
condition. Calculate mutually exclusive trust/anyhoo/scam outcomes under [§4.3](../project/GAME_RULES.md#43-promises-and-reputation)
and Q12, with the 40% threshold measured against the promise. Keep one aggregate
outcome per ordered payer/recipient pair per game; disconnection leaves an
unresolved promise pending, and nullification produces no reputation outcome.
Loan acceptance must reject any new alliance involving an already-allied
player; valid alliances are exclusive pairs that survive card returns. Accept
an intact loan/return only at a clear action boundary on either party's turn;
it stays exposed and consumes no new opening allowance. Receiving an intact loan
needs no ability support; support, timing and quotas apply when using it.
An alliance or loan-derived Underground privilege appears only after successful
transfer resolution, never on proposal or acceptance alone. Returning surviving separate royals
leaves them exposed/unassigned; return cannot reclaim cards held elsewhere.

The acquisition and visibility tables in
[§2.13](../project/GAME_RULES.md#213-acquisition-and-visibility) govern every
snapshot, event, history entry, accessibility label and bot observation. Hand
size and concealed-Ace count are public; faces and hidden royal/eligible-card
breakdowns are private. Captured number Diamonds are exposed even when Justice
took them from a hidden zone; ordinary captured numbers/royals otherwise enter
the hand, and acquired Aces enter the concealed Ace area. Named transfer rules
override those capture defaults. Random selection never discloses unselected
cards. Keep already-public identity history without exposing new hidden draws
or another hidden king's location through Doppelganger binding metadata.
Snapshots distinguish cumulative match score, origin-game score, that game's
available points, debts, phase, lasting-effect owner and physical-card custodian;
fractional values use `{"numerator":"11","denominator":"2"}`.
`online.rules_context` supplies the observer's opening/Ace/Compensation use,
own quota markers and Underground flag, applicable purchase unit price and
Inflation status. `quotas` values are engine interval markers, not remaining
counts. `quota_used` interprets supported ability counters for the current
round, observer's own-turn cycle or game lifetime, including shared Ace limits;
a false value means the quota is unconsumed, not that the action is legal.
Its optional `combat` contains public physical attack/defense totals, modifiers,
effective totals and the comparison operator (`equal`, strict `greater` for
Baron, `at_least` for Bomb). It is omitted if an operand is concealed; no physical
card identifier is added. This context does not assert that a particular intent
is legal, and the existing narrow board menu is not a full action catalogue.
Include the current game ID and immutable IDs for historical game results, and
attribute game events and pending operations to their originating instance.
Expose public active-effect records with source, target, custodian and expiry;
their physical Aces remain in the effect zone through Fate, outside its reset
count. Separate card custody from reusable gameplay control. Record physical
`available_from_round` restrictions under [§2.14](../project/GAME_RULES.md#214-restricted-cards)
and filter concealed metadata: an observer must not locate a restricted card
through a hidden transfer or reshuffle.

Combat previews and committed results carry authorized physical selections,
effective card values, applicable modifiers and comparison outcome under §2.5.
Keep effective totals distinct from the physical Clubs returned in an exchange;
modifiers never create payment cards. Publish per-seat Code economics and round
penalty eligibility separately: the declaration is a non-attacking global
economic action, while its recurring hostile penalty excludes protected/confined
opponents and requires the active, non-confined declarer's two current series.
The rules-context fields above implement committed combat calculations and the
observer's purchase economics. The client explains the selected action's timing,
declaration and resolution costs, cancellation consequences, and visible own-Club
usage. Arbitrary pre-submission previews and exhaustive penalty-eligibility
calculations remain client-integration work; the server decides legality.

## Trade and loan proposal contracts

[§2.15](../project/GAME_RULES.md#215-trade-and-loan-proposals) separates an
unanswered proposal from an accepted transfer action. Persist `offer_id`,
revision, parties, exact terms/card handles, origin `game_id` and `turn_id`,
status and any accepted action ID. Return offer terms only to their authorized
parties; other seats receive only already-public knowledge and permitted effects.

| Operation / state | Contract |
|---|---|
| Create/revise → `pending` | At a legal idle boundary; both parties active and non-confined. Ordinary trade creation uses maker's turn; loans may use either party's turn. No cards, opening quota or Ace use are reserved. Revision requires fresh acceptance and cannot extend the original turn expiry. |
| Withdraw/decline → `withdrawn`/`declined` | Maker withdraws or recipient declines an unanswered offer. These record operations open no response window and serialize against acceptance; they cannot undo an accepted action. |
| Turn end/board closure/party departure → `expired` | Expire unanswered offers atomically with that boundary. Silence never blocks End turn, accepts, or produces a reputation refusal. Never retarget into another turn/game. |
| Validate acceptance | Require exact offer ID/revision and current origin game/turn; both parties eligible and no pending action/response/effect/automatic work. Check ownership, availability, trade restrictions, transfer quotas and alliance compatibility. An unfulfillable offer remains invalid for acceptance until revised/closed; a rejected acceptance spends/reserves nothing. |
| Legal acceptance → `accepted` | Accepting player is action actor/declarer; response order starts to their left and Confinement targets them. Ordinary out-of-turn acceptance grants no unrelated action. Each supplier of an Ace spends their applicable Ace allowance at legal acceptance. Other transfer/opening costs are absent. No withdrawal or revision remains legal. |
| Resolution → `completed` or `canceled`/`failed` | After responses revalidate the exact transfer, then atomically move all cards and establish any alliance/loan-derived privilege. Failure/cancellation moves nothing and creates no alliance; prior resolved responses/committed allowances stand. Terminal proposals never reopen automatically. |

Use distinct offer status and acceptance-legality/reason fields: invalid current
terms do not manufacture a blocking decision. Serialize revisions, acceptance,
withdrawal, decline and expiry under the same match boundary so exactly one
competing transition wins. Identical acceptance retries return their original
action/result, including while that action waits for responses. Reconnect restores
that same revision/status/action; the client must not optimistically transfer or
reserve cards the authority has not moved.

Illustrative acceptance payload inside the standard game-scoped envelope:

```json
{
  "type": "accept-offer",
  "payload": {
    "offer_id": "opaque-offer-id",
    "revision": 2
  }
}
```

## Deck purchase contract

[§2.16](../project/GAME_RULES.md#216-deck-purchases) requires explicit positive
integer `quantity` and exact unrestricted number-Spade payment handles from the
buyer's hand or exposed Spades. No Underground or generic two-series prerequisite
applies. At an own-turn idle boundary, validate current exact payment value
`>= quantity × price` and supply including returned payment cards. Code sets
price and Inflation changes effective Spade values; excess payment gives no
change. The buyer may choose less than the affordable maximum.

An accepted purchase starts the ordinary response procedure with its buyer as
actor. Payment waits for successful resolution: revalidate the same quantity,
physical cards, effective values and total supply after responses. On success,
return/shuffle the payment, draw exactly that quantity and update eligible
Inflation removal atomically. A failed/canceled action pays/draws nothing and
never silently reduces quantity. Hidden payment selections stay private until
public return; public selections remain known. Commit random draws once, with no
Coup or response interrupting half a purchase.

```json
{
  "type": "purchase",
  "payload": {
    "quantity": 1,
    "payment": ["authorized-number-spade-handle"]
  }
}
```

The UI previews quantity, price, effective total, physical payment and excess.
One Spade 10 with otherwise empty piles can buy a chosen one card at price 5,
because the returned payment supplies it; requesting two rejects without payment.

## Phase and decision contracts

| Phase | Admitted mutations and transition |
|---|---|
| `playing` | Legal gameplay, effect decisions, trades/loans, typed agreements and currently permitted declarations. A valid ending closes the board immediately. |
| `board_closed_settlement` | Only origin-game voluntary payment/offer acceptance/rejection, explicit promise refusal, permitted financial consent and finish-settlement commands. Reads, result reconciliation and reconnect remain available. No card action or further victory. |
| `financially_finalized` | That game's final results and cash pool are closed. A next-game command may start the next configured game only after this transition, with zero new-game available points and retained match scores/debts. After the final configured game, publish cumulative-score ranks with equal scores sharing rank. |

`finish-settlement` requires the actor's triggered promises and offers to have
terminal outcomes, including incoming/outgoing final offers; it cannot silently
refuse them. No new promises may be created after board closure. A finished
participant cannot initiate further discretionary payments/offers, but mandatory
creditor receipts still apply until global finalization. Finalize only after all
required participants finish and automatic debt work is complete. Silence is
pending, even for a departed player with a settlement decision; it is never a
refusal. Transfers retain their origin game and count as recipient gross receipts
before automatic debt redirection. These commands implement
[§4.4](../project/GAME_RULES.md#44-financial-lifecycle), including
Coup's cash replacement and departure/nullification rules. A closed board is
not yet a financially finalized game. Show the threshold declarer/Coup winner,
game score leader and eventual match rank separately. Unanimous nullification
restores the game-start financial snapshot and creates no competitive/reputation
result; the void game's replacement occupies the same configured match slot.
It always has a new game ID; finished/nullified games retain immutable audit
identity and original duplicate-command results.

Forgiveness commands name the debt, exact unpaid amount and required consent.
For a charge retained in a completed game, expose a separate nonspendable match
adjustment linked to its original debt/charged entry and the game creating the
forgiveness. This adjustment survives later Coup replacement without rewriting
the completed result. A current-game charge is corrected in that current game,
so Coup replaces both charge and correction together. Track carried-charge
lineage to ensure exactly one correction against the still-recorded charge.
Neither correction is cash or a repayable receipt. Nullification restores the
debt and removes corrections created in that void game from effective totals,
preserving the audit history. The client shows historical game scores and match
adjustments separately rather than editing old results.

For Negotiator, persist attacker and defender selections as stages of one
effect. After the final required choice, commit newly selected Club usage,
Spade transfer, jack return and attack cancellation atomically. Original Clubs
were already used on legal attack declaration. Justice and Dexter also expose
their actual decision owner and authorized stage. Justice commits/reserves each
sample and permitted knowledge before final atomic transfer/queen return; Coup
preserves that knowledge while abandoning uncommitted movements. Committed
random choices survive retry. Costs, quotas and cancellation follow
[§2.12](../project/GAME_RULES.md#212-ability-lifecycle-and-costs), not
client inference from whether cards disappeared. Coup can interrupt a wait
between atomic transitions, never a half-applied mutation.
When advancing the ordinary response cursor, automatically omit confined and
departed seats while preserving the remaining fixed seat order. These seats
have no required pass/input; skipping is an eligibility transition, not a
timeout. An eligible disconnected seat still blocks on its response. Confinement
expiry restores normal eligibility, and a confined player's separate legal Coup
permission remains available.

After an action and its responses, process the earned Compensation queue before
automatic point settlement or any ordinary action/victory. Recipients use fixed
ascending seat order, independent of initiative, and each receives their whole
owed quantity before the next seat. Apply returns/shuffle first and preserve
ordinary recycling/exhaustion; consume each five-loss entitlement even if its
draw is unavailable. Persist queue position, counters and committed random
outcomes when split into transitions. Show authorized progress rather than a
new ordinary-action boundary. A legal Coup between atomic queue steps retains
completed draws and abandons uncommitted remaining draws at board closure.

## Victory admission

Authenticate and authorize, lock the match, and deduplicate before checking the
named game instance, current phase, holdings and eligibility. Do not reject a
legal Coup solely for an unrelated aggregate version change within that same
game. A delayed first submission naming an earlier game must reject even when
its actor could win the current game. Apply the same admission rule to an ordinary
threshold claim only at a permitted completed-action boundary and while its actor is not
confined. A lost required card, alliance membership barring Coup or closed
board still rejects the intent. An already-confined player retains the explicit
Coup exception. Ordinary selected actions still use optimistic version checks;
responses/effect inputs additionally bind to their live window/decision IDs.
This preserves authoritative processing order; it does not promise arrival-time
fairness. Duplicate successful declarations return their original result even
after the board closes.

## Errors

Return stable `code`, localization `args` (currently empty) and `request_id`.
Do not encode private eligibility details in messages, response shapes or logs.
The Flutter ARB catalog localizes these codes; the server sends no English
rule-error strings. Internal illegal/stale-decision rejections intentionally use
coarse safe codes rather than disclosing which private eligibility test failed.

| HTTP / code | Meaning and next action |
|---|---|
| 200 | Committed command or original durable result for an identical retry. |
| 400 / `INVALID_REQUEST`, `INVALID_COMMAND`, `INVALID_CURSOR` | Invalid shape/value; correct the input. |
| 400 / `CREDENTIALS_REJECTED` | Registration/login/upgrade rejected without account details. |
| 401 / `UNAUTHENTICATED`, 403 / `FORBIDDEN` | Authenticate or stop. |
| 409 / `STATE_CONFLICT` | New command has stale game/version; refresh, never retarget automatically. |
| 409 / `COMMAND_CONFLICT` | Same command identity with different body. |
| 409 / `ROOM_REJECTED` | Room operation rejected. |
| 409 / `RESNAPSHOT_REQUIRED` | Cursor ahead, older than 1,000 revisions, missing event or oversized replay batch; fetch snapshot. |
| 429 / `OVERLOADED` | Bounded admission rejected; use a bounded retry budget. |
| 503 / `UNAVAILABLE`, `OUTCOME_UNKNOWN` | Reconcile ambiguous mutations using the same identity. |

WebSocket data frames are `{"cursor": N, "data": {"version": N,
"cursor": N, "projection": …}}`. Apply only contiguous cursors; ignore repeats,
resnapshot on a gap and never retarget old commands into a fresh game. A terminal
stream control frame has `{"type":"error","code":"RESNAPSHOT_REQUIRED"}`,
`OVERLOADED` or `STREAM_UNAVAILABLE`, followed by closure. Pre-upgrade library
failures can be plain stable text (`OVERLOADED`, `UNAUTHORIZED`, `INVALID_STREAM`)
instead of the ordinary JSON error envelope.

Retry count, five-second end-to-end deadline and reconciliation are defined
once in the state/delivery design. A disconnect after commit is not a failed
game action. An accepted command may durably start automatic settlement and
return a pending settlement identity; it must not claim final payment while
private batched work continues. Reconciliation preserves that result, and a
later committed completion event/state announces final publication. Reads can
show progress, but no gameplay or subsequent financial mutation overtakes it.
Responses/pass IDs and pending-action IDs prevent replaying an
old response into a new window; decision IDs prevent a continuation from answering
a different stage of the same effect. Duplicate lookup precedes game-instance and stale-version checks
and requires current authorization, so retry never grants former members access.

Implement accepted post-response settlement, including Q01 failure after
Advancement and [§2.10](../project/GAME_RULES.md#210-negotiator-resolution)/Q08
Negotiator's maximal exact-match exception with fixed committed modifiers. A final-response
crash/retry fixture must preserve paid costs and committed-card usage exactly once.
For a future undefined experimental transition discovered only after a legal
response resolves, preserve that response and an `adjudication_required` paused
action together; return its successful durable result with the pause reason.
Never relabel prior effects as a no-effect 422 or repeat their costs. Such profiles
cannot certify GAME_RULES or run as production accepted-rule games. Reject
unsupported engine/rules versions before match creation.

A valid threshold claim may count controlled hidden cards; reveal only enough
selected cards to prove it. Reject invalid digital claims privately without
changing the game or exposing hidden identities. Server order determines the
first valid declaration; Coup may interrupt pending responses but never reverse
a terminal result. Disconnection or silence preserves pending state indefinitely
for reconnection or unanimous resignation; transport timeouts never pass a turn.
Restricted Barricade draws may count toward ordinary thresholds, including their
necessary proof, but cannot supply Coup or voluntary play/payment/transfer until
their recorded round. Forced Diamond exposure never clears that restriction.

## Versioning policy

`/v1` identifies the transport contract, independently of rules, engine and
interpretation versions. Pin those at match creation. Additive fields require
compatible decoders; incompatible state/events need version negotiation or
old matches drained under their original engine. OpenAPI/stream schemas and golden contract fixtures are checked by the transport
and wire suites. Physical card IDs never appear on the wire: current capabilities
are keyed by actor/match/game/view, while known-history handles are a separate
namespace and cannot authorize selection. Committed retries resolve before
current capability checks, so visibility changes cannot destroy their receipts.

## Rate limits

The current limits are 16 KiB request bodies, 32 concurrent HTTP preflights/
requests, two-second request contexts, 128 dispatcher jobs overall and 32 per
match. IP/account buckets allow a burst of 20, replenish five/second and use a
bounded 1,024-key map. Streams allow 64 connections, two/account, eight concurrent
reads, 128 events/1 MiB per batch, one-second polling and five-second write
limits. The server pool has 16 connections. Slow readers have no broadcast queue
and cannot hold match transactions. One owned scheduler loop pages 32 matches
and advances only persisted automatic work; it never answers for a human.

Redis presence uses four concurrent calls, a 100 ms deadline and zero retries;
local fallback holds at most 1,024 expiring hints. It is neither game state nor
an allowance counter. Technical throttles never create a game-turn deadline.

## Verification scope

Real PostgreSQL three/four-client scenarios cover rooms, private proposals,
loan resolution, post-board voluntary promises at the exact 40% boundary,
financial finish/finalization, historical command reconciliation and authenticated
TCP WebSocket replay. Lower-level engine and matchstore suites cover the full
rule matrix, concurrent declarations/revisions, hidden-state metamorphism,
restartable decisions, exact settlement, nullification and cross-game corrections.
These are targeted scripted scenarios, not a human usability or production-load
study. P03 bot-strength/balance confirmation remains separately allocation-blocked.
