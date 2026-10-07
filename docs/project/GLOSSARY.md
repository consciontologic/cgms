# 📖 Glossary

Domain and project terms used across CGMS. Each entry summarizes a term;
[GAME_RULES.md](GAME_RULES.md) governs gameplay and the
[rule register](../design/RULES-IMPLEMENTATION-QUESTIONS.md) maps accepted decisions to future evidence.

| Term | Meaning |
|---|---|
| Active-effect zone | Public physical location of resolved Inflation/Compensation Aces, separate from playable inventory and excluded from Fate's reset count. |
| Actor | The authenticated player responsible for an accepted action; trade/loan acceptance makes the accepting player its actor. |
| Agent | An AI assistant working under the repository's operating rules; it is not the human decision owner or an authoritative game participant. |
| Alliance | One exclusive permanent pair for a game, established by a successful combo loan; it does not pool victory holdings. |
| ARB | Application Resource Bundle, the required canonical Flutter UI message-catalog format. |
| Available points | Spendable points in the current game's financial pool after required debt settlement, distinct from recorded score and dirt. |
| Available round | A physical card's `available_from_round`; Barricade-drawn cards retain active-use restrictions until that boundary even through forced movement. |
| Board closure | The gameplay ending that stops further card actions/victory claims and begins permitted financial settlement. |
| CGMS | Card Game Mafia State, the three- or four-player negotiation card game and this project. |
| Code | A queen combo whose latest declaration governs opponents' economics; its hostile recurring penalty has separate support and eligibility conditions. |
| Compensation | A persistent Ace effect that counts qualifying involuntary exposed-series losses and earns deferred draws, ordered across recipients by fixed ascending seat. |
| Coup | An immediate, separately qualified victory declaration with replacement +50/−50 scoring; it is distinct from an ordinary possession-threshold ending. |
| Current series | A suit presently supported by exposed number cards, distinct from the history of suits a player has opened. |
| Debt | An exact unpaid match-scoped obligation to a player or the system, repaid oldest-first from later receipts without charging the original loss twice. |
| Decision | A persisted actor-specific input stage within an accepted effect; it is not another response window. |
| Dirt | Online-earned product currency redeemable only for Daily and Weekly access; these earned packages retain ads unless a separate ad-free entitlement is active. Dirt cannot buy cosmetics, Ad-free, Monthly or Yearly products and is separate from game score, available points and debt. |
| Doppelganger | Any two Kings assigned to one replacement royal slot, with persistent binding rules and two distinct physical identities. |
| Effect | A rule-governed consequence with its own identity, commitment, required decisions and resolution; it may span multiple commands. |
| Effective value | A physical card value after applicable modifiers, used for a specified comparison/payment; virtual modifiers never create physical cards. |
| Fate | A combo reset of a target's personal cards that preserves specified history/effects and excludes persistent active-effect Aces. |
| Financial finalization | Closure of a game's remaining cash pool and final result/reputation after required automatic settlement and explicit participant finishes. |
| Fixed seat | A player's stable seat identity/order, independent of the initiative order chosen for a round. |
| Game | One uniquely identified board and financial lifecycle within a match; a nullified replacement gets a fresh identity even at the same match slot. |
| Game ID | Immutable authority-issued identity required by every game-scoped intent so delayed commands cannot affect a later game. |
| Gross receipt | A positive amount entering a recipient's ledger before automatic debt redirection; it can count as delivered sharing without becoming spendable cash. |
| Historical opening | The recorded fact that a player opened a number suit earlier in this game, retained after losing its cards or undergoing Fate. |
| Idempotency | Returning the original authorized durable result for an identical command identity/body instead of applying its effect again. |
| Inflation | A persistent targeted Ace effect halving number-Spade values until its qualifying removal or lifecycle expiry. |
| Match | The configured sequence of games with cumulative results and surviving debts; default length is three and equal totals share rank. |
| Match correction | A nonspendable adjustment for forgiveness against a retained finalized-game charge, with explicit debt, charge and creating-game provenance. |
| Negotiator | The Spades Jack ability that cancels a club attack through the specified Spade payment or greatest exact adjustment using eligible committed/unused Clubs. |
| Observation | The public and seat-private information an authorized player/bot may receive; it excludes another seat's concealed identities and future draws. |
| Ordinary threshold | A declared win based on controlled number Diamonds or physical royals plus required suit-opening history, validated at the permitted completed-action boundary. |
| Physical card | One distinct card in the two-deck, 104-card inventory; equal rank/suit copies remain separate and occupy exactly one zone each. |
| Proposal | A nonblocking trade/loan offer with exact terms and revision, no pre-acceptance card reservation, and explicit withdrawal/decline/turn-expiry rules. |
| Purchase quantity | The buyer-chosen positive integer number of deck draws, funded by sufficient eligible effective Spade value and available supply; excess payment gives no change. |
| Replay | Re-execution of recorded versions, commands, decisions and random outcomes to compare state/event results; a seed alone does not prove equivalence. |
| Response window | The single seat-ordered opportunity for eligible other players to respond or pass before a pending action resolves; nested effect decisions do not create new windows. |
| Round | One initiative-ordered cycle of scheduled turns, also defining named per-round quotas and card-availability boundaries. |
| Ruleset | The pinned versioned gameplay definition/configuration for a match; experimental interpretations must be separately named and cannot certify accepted-rule conformance. |
| Run | An agent invocation or identified simulator execution; agent work uses `run_id`, while simulator manifests separately identify reproducible experiments. |
| Scope | The named slice of repository work, such as a roadmap phase or module; it does not authorize unrelated changes. |
| Score | Recorded competitive points, distinct from spendable available points, outstanding debt and product currency. |
| Settlement | Automatic exact debt processing plus, after board closure, the permitted voluntary financial decisions needed before finalization. |
| Tracking row | An append-only record in `docs/tracking/tracking.csv` describing an agent action; runtime checkboxes require their own evidence. |
| Turn-cycle | The default player/ability allowance interval that resets at that player's next scheduled turn; named round/game limits take precedence. |
| Underground | A King combo with lasting player-earned privileges distinct from its conditional income, which counts only allocated functioning formation Kings. |
