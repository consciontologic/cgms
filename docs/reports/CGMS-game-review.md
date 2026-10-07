# CGMS: an honest rules and balance review

> Historical review of earlier drafts, retained without rewriting its findings.
> [GAME_RULES](../project/GAME_RULES.md) and [accepted decisions](../../CHANGELOG.md)
> supersede its proposed changes and unresolved-rule descriptions; the
> [decision register](../design/RULES-IMPLEMENTATION-QUESTIONS.md) records the resolved Q01–Q12 rulings.

Based on the supplied “CGMS (Card Game Mafia State)” rules.

This is a review of the written design, not a playtest report. I distinguish definite wording conflicts, missing rules, and balance risks. The examples show situations the rules currently permit or fail to resolve; they do not establish measured win rates.

## Overall assessment

CGMS has a coherent central idea: four suits provide different resources, exposed cards make players vulnerable, and temporary cooperation can help players build power. The choice between keeping a royal for a combo, using its individual ability, and collecting royals toward victory could be interesting. Diamonds also create a useful tension between revealing strength for initiative and exposing valuable cards.

However, **the current rules are not ready for a group to learn and adjudicate independently**. Several ordinary situations do not have a single answer. Those gaps matter more than whether a particular card should cost one additional point.

The largest design risk is that the game asks players to build a position over many turns, then includes effects that can erase or appropriate nearly all of that work with limited recovery. That can suit a deliberately treacherous game, but only if the timing, counterplay, and scoring are exceptionally clear.

My priority order would be:

1. Define the game structure, scoring objective, and victory checks.
2. Write one complete attack and response procedure.
3. Clarify alliances and ownership.
4. Limit or make reversible the strongest lasting effects.
5. Playtest the resulting stable rules before tuning individual numbers.

## 1. Rules gaps that can stop a game

### 1.1 “Winning the game” and winning on points are different objectives

The introduction says the winner has the highest total score across games. Elsewhere, reaching 13 diamonds or 15 royals lets someone declare victory. There is no clear ordinary victory bonus or rule explaining how that declaration relates to the score.

This can produce a counterintuitive result: a player ends a game by collecting 15 royals, yet another player scores more through diamonds and recurring income. That is not inherently wrong, but players need to know whether the card threshold wins that individual game, merely ends it, or awards something toward winning the overall match.

The two thresholds also score very differently. Thirteen number diamonds are worth **127–159 points** under default scoring, depending on their printed values. Fifteen royals do not have a comparable guaranteed end-game value: only spade royals explicitly score 10 points each, although some other royals can generate points through abilities.

**Suggested fix:** distinguish a **turn**, **round**, **game**, and **match**. State explicitly what reaching a threshold does. For a first test, the simplest scoring model is that the threshold ends the game, everyone scores, and the highest cumulative match score wins. If the threshold achiever should instead receive a bonus or automatically win the game, write that separately and test it.

### 1.2 Round structure and changing turn order are undefined

“Each player starts the round with 20 cards” sounds like a new deal every round, while the rest of the game depends on retaining cards and drawing one per round. I suspect you intend a single initial deal per game.

You also say that turn order changes when diamonds are added. If that changes the order immediately, someone who has already acted could move below someone who has not, leaving uncertainty about repeated or skipped turns.

Recurring income and limits such as “once per round” cannot be applied reliably without a clear round boundary.

**Suggested fix:** deal 20 once per game. At the beginning of each round, determine initiative from the sum of exposed diamond values and freeze the order. Each player takes one turn and draws at its start. After everyone acts, resolve round-end effects. Specify when game 13-round scoring interrupts or replaces those effects.

Also cover zero-diamond starts and repeated ties. A player has about a **1.43%** chance of receiving no number diamonds in an initial 20-card hand, so zero is a real setup case.

### 1.3 Combat does not yet have one complete procedure

The exact-match rule is understandable in principle, but several decisive details are missing:

- May an attacker choose a subset of their clubs, or must they use the whole exposed club series?
- Does clearing hearts end that attack, or may the same clubs proceed into the next suit?
- Can different clubs attack different opponents in one turn, despite the once-per-series rule?
- Does a series count as cleared when its number cards are gone, even if royals remain attached?
- When can the defender add a non-combo card, negotiate, or play an ace?
- If an attack becomes invalid after a response, are committed cards spent?

**Suggested first-test rule:** choose one opponent, one legal target suit, and any subset of your exposed clubs. Declare the target cards. Resolve responses, then check for equal totals. Resolve one exchange or capture and mark the club series used for that turn. Removing the last number card clears that suit for attack-order purposes. This is a proposed simplification, not something your current wording already settles.

If you intend multi-stage attacks, write a worked example that clears hearts, exchanges clubs, and then captures another suit, specifying what is available at every stage.

### 1.4 Instant effects need a response order

Aces can be used outside your turn. Coup and Ponzi can also be declared outside your turn. Confinement can stop actions before an opponent finishes their moves.

Consider: Alice declares Coup; Bob immediately declares Confinement against Alice. Does Coup finish the game first? Does Confinement cancel the declaration? Could Alice have used Confinement on Bob first? The current text cannot decide this without an argument about who spoke first.

**Suggested fix:** every action has a declaration, a defined response window, and resolution. State who may respond, in what order, and whether further responses are possible. Check victory only after the action and its responses finish. Explicitly state whether Coup and Ponzi ignore ordinary requirements such as two open series and an active number club.

For Confinement, replace “including speaking” with a precise restriction on game actions. Silencing a player complicates rules questions and negotiation without clarifying how the card works. Also specify whether it can target its user or an ally for immunity, how long immunity lasts, and whether it cancels an already declared effect.

### 1.5 Alliances change victory without defining whose cards count

The text alternates between a player reaching a threshold and a player **or alliance** doing so. Giving or receiving a combo creates an alliance that cannot be broken, even after the cards are returned.

You need answers to all of these:

- Do allies add their card counts together, or must one player physically control the winning cards?
- Does pooling apply to diamonds, royals, or both?
- Must every member have previously opened all four suits?
- Do hidden royals and later-drawn hidden diamonds count?
- Does a loaned combo count only for its current controller?
- If A allies with B and B with C, are all three one alliance?
- Can everyone join the same alliance?
- What determines the highest-scoring member before the rewards are allocated?

**Suggested fix:** separate legal control of cards from promises about rewards. Count each physical card once. Either have one controller meet the threshold, or define a formal team victory with explicit membership, eligibility, and scoring. Make accepting a permanent alliance a deliberate declaration; an ordinary card transaction should not create one accidentally.

### 1.6 Combo-opening rules have multiple plausible readings

You allow either a combo or a series in a turn, then allow “as many combos as you need,” then say “only one combo at a time.” That last phrase could mean one at a time sequentially, but some players will read it as one per turn. Underground advertises unlimited combo opening even though the base rules appear to allow it already.

**Suggested fix:** put the whole rule in one place. For example: on your turn, either open one new number series, or open/reconfigure any number of combos; you may still use existing abilities as specified. State whether adding cards to an existing series counts as opening one, and list Underground’s exact exceptions.

## 2. Concrete wording corrections

| Location | Problem | Correction needed |
|---|---|---|
| Basic combat | The queen of clubs is said to make attackers spend clubs against hearts. | Barricade is the queen of hearts; Exile is the queen of clubs. Correct the reference. |
| Diamonds | Setup requires exposing initial diamonds and prevents taking series back, but the diamond description says to keep them “in your hand.” | Use “under your control” if both exposed and hidden holdings score; separately define which holdings count for victory. |
| Aces | The general rule says all used aces except hearts are discarded, but Inflation stays open until removed. | List Inflation as another persistent exception. |
| Scoring | “Red army” is not otherwise defined. | It appears to mean Coup; use one name. |
| Diamonds | “Rule” changes diamond scoring, but no such named ability exists. | It appears to mean Code. Code also allows decreases, not just increases. |
| Baron / Suicide Bomb | Baron mentions protection for the rest of the turn; Suicide Bomb protects against everyone for the rest of the round. | Say explicitly whether Baron changes that duration. The current wording may be read either as an exception or a shortened reminder of the normal rule. |

These are comparatively easy fixes, but they should be made before the next independent rules test.

## 3. Strongest balance and exploit risks

### 3.1 Ponzi: the most concerning effect

Ponzi claims all of another player’s points earned before and after activation in the current game. Its cards return to the deck, but the claim lasts for the remainder of that game, with no Ponzi-specific removal procedure or guaranteed recovery action. The general rules allow negotiating away penalties, but do not establish clearly whether that can end Ponzi.

Under that reading, the victim may spend the rest of the game generating points for the attacker. Their remaining incentives could become helping someone else win, forcing the game to end, or quitting. That is a serious risk to engagement in a scoring game.

Returning the cards also makes another activation possible. If A uses Ponzi on B, then B uses it on A, you need rules for overlapping claims. Does the second effect capture points transferred by the first? What happens to negative scores?

The prohibition on alliances and Doppelganger makes Ponzi harder to assemble. That does not solve what happens after it succeeds.

**Recommended change:** start by testing a one-time, capped theft, or a claim on income for one round. If permanent control of income is essential to the design, attach it to an exposed, removable object and give the victim a concrete way to recover. Explicitly write “both jacks of clubs and both jacks of spades” if that is the intended recipe.

### 3.2 Coup: potentially exciting, but able to override most of the game

Coup instantly ends the game, awards +50 to its owner and −50 to opponents other than those who have quit or been defeated, and ignores previous points. That creates a **100-point difference** between its owner and each affected opponent, regardless of their positions before it resolves.

The exact four-card requirement is restrictive. A particular initial 20-card hand contains all four required queens only about **0.105%** of the time, assuming a fair shuffle. That is not its full-game occurrence rate: trades, extra draws, and theft can assemble it later.

Its concern is the combination of a score reset, immediate ending, and out-of-turn declaration. Players may reasonably feel their earlier decisions stopped mattering.

**Recommended change:** first test Coup as an immediate game end with a stated bonus while preserving earned points. If you keep the score reset, make it an intentional signature feature and provide a clear response opportunity. Explain whether all end-game card points are also replaced by the +50/−50 result.

### 3.3 Code: two different effects, only one of which can be stopped

Code changes other players’ diamond scoring and purchase costs permanently, even if its queen count drops below five. Keeping it intact also drains 5 points per enemy per round. Its owner continues using default rules.

For an opponent holding 13 diamonds, reducing the per-card scoring bonus from 5 to 1 removes **52 points**. Raising the purchase cost from 5 to 10 halves how many cards an exact 10-spade payment buys. These are substantial effects before the recurring penalty.

Destroying Code stops the recurring penalty under the “preserve it” condition, but does not reverse the economic rules. That weakens the reward for successfully dismantling it. Multiple Codes also create unresolved questions about whose prices apply and who is exempt.

**Recommended change:** make its economic changes last only while the combo is intact and active. Define whether multiple Codes can coexist and how they interact. Start with one benefit at a time—price manipulation or recurring point loss—until tests justify both.

### 3.4 Underground: an income engine plus extensive flexibility

Five kings provide 10 points per round, permit opening without the usual two-series requirement, and allow open-card trading. More kings add more income.

At its base rate, remaining active across all 13 round ends would produce **130 points**, before other scoring. That is comparable to the value of an entire 13-diamond winning holding. This is an upper-bound illustration, not an expected outcome; assembling and preserving five kings has a real cost.

For a fair comparison, the literal Richer wording permits a non-diamond king in each of four series, producing 8 points per round. Underground's income alone therefore does not prove it is overpowered; the combination of income and exceptional flexibility is what needs testing.

Open-card trading is also a major benefit. It can let a player convert otherwise committed cards into resources or move exposed holdings out of danger. The normal irreversibility of opening a series is much less binding on that player.

**Recommended change:** explicitly end all ongoing benefits when the combo falls below five kings or loses any required enabling condition. Specify whether its defensive classification means it requires exposed hearts even during its early-opening exception. Test lower income before adding more constraints; its trading flexibility may already be a sufficient reward.

### 3.5 Kidnapper may be more efficient than the larger combos

Two identical jacks permit a theft every round, including exposed royal cards, with no stated per-use resource cost. A theft can increase the attacker’s royal count, reduce the defender’s count, and break a costly combo simultaneously.

That is potentially excellent value compared with collecting four or five royals for other effects. Great People is a counter, but naturally forming it requires the only two queens of hearts.

**Recommended change:** define what “steal one closed card” means—random selection, a named request, or inspection. Then test Kidnapper closely. If it dominates, consider consuming a component, charging a resource, or restricting repeat thefts. Its Doppelganger exception also needs to specify whether it takes both physical kings and exactly how much of an affected combo collapses.

## 4. Structural gameplay difficulties

### 4.1 Exact matching can make a smaller defense better than a larger one

Suppose the attacker has clubs 3, 4, and 8, and the defender has only a 2 of hearts. Even if the attacker can choose any subset, none totals 2. The lone heart prevents progress through the normal attack order and can keep defensive abilities enabled.

This is not automatically bad. It could be an intentional numerical puzzle. But it creates a risk of stalled combat, especially when the relevant ace or bypass card is unavailable. Larger armies do not reliably overcome small defenses.

It also asks players to search combinations of attacking and defending cards repeatedly, with other players waiting.

Small clubs gain another advantage from combat scoring: spending 2, 3, and 5 against enemy clubs earns 9 points, while spending one 10 earns 3 points. Both commit the same printed strength. Decide whether rewarding the number of cards spent, rather than their strength or the number of enemy cards removed, is intentional.

**Recommended test:** compare strict equality with a version that permits overpayment, with the excess attack value wasted. Do not assume the easier rule is better; measure blocked turns and decision time. If exact equality is a defining feature, give players a dependable, limited adjustment mechanism.

### 4.2 Initiative, victory, and scoring all reward diamonds

Diamonds help determine move order, contribute heavily to scoring, and trigger one ending condition. Captured diamonds must also be exposed immediately. A player who gets ahead can therefore gain both a stronger finish and earlier opportunities to act.


The defensive suit order and exposure risk may balance this, but that requires evidence. Record how often the initiative leader stays ahead and wins. If it is excessive, change one of these rewards rather than weakening all of them at once.

### 4.3 Removing a player’s resources also removes their ways to respond

Losing hearts can disable defensive combos; losing clubs disables offensive abilities; losing spades reduces buying power. Those dependencies may create a downward spiral in which a player remains in the game but has little meaningful agency.

Compensation could help, but requiring the right ace and losing five cards per replacement may make recovery inconsistent. Fate can also reset a player’s developed position, depending on what “all their cards” includes.

**Recommended test:** track consecutive turns with no useful attack, purchase, or ability. If this commonly lasts multiple rounds, add a small universal recovery option, such as a limited exchange, rather than relying entirely on drawing a rescue card.

### 4.4 Returned cards can be deliberately positioned in the deck

The general return rule lets players place cards somewhere in the deck or shuffle it as they like. That permits intentional top-deck placement and feeding known draws to players according to initiative. Some specific effects require shuffling, while others do not.

This is allowed manipulation under the written rules, not cheating. But it can have consequences much larger than the ability that returned the card, and the rules do not assign control of the insertion point clearly.

**Recommended change:** standardize the return procedure. For example, shuffle cards into the draw pile whenever an effect returns them, and use a separate discard pile for everything else. Also specify what happens if both the draw pile and discard pile are empty.

### 4.5 Numerical aces cannot be adjudicated clearly

An ace can represent “any number,” but must stay closed and its role must not be exposed. Exact-match combat requires opponents to know the value used. For example, clubs 8 plus a hidden ace attacking hearts 10 immediately implies a value of 2.

It is unclear whether “exposed” means physically showing the ace’s suit, announcing the represented number, or simply letting someone deduce its role. The permitted number range is also undefined.

**Recommended change:** announce a value in a defined range, then reveal and discard the ace when resolving the action. If bluffing is intended, it needs a separate challenge rule. Clarify whether trading an unused ace merely transfers it or consumes that player’s once-per-round ace allowance.

### 4.6 Alliances may produce kingmaking rather than competition

Nonbinding promises and betrayal fit the theme. However, permanent alliance membership disables Coup and Ponzi, while the highest-scoring partner can keep the rewards. A weaker partner may therefore give up independent options without gaining any enforceable benefit.

That makes negotiation experience and personal relationships especially influential. It also increases the chance that a player who can no longer win mainly determines which other player wins.

Keep betrayal if it is central to CGMS. But distinguish binding game rules from unenforceable promises, and decide whether an alliance is a lasting strategic commitment or a temporary deal. The current design uses language suggesting both.

### 4.7 Administrative load is high for a standard-deck game

Players must track historical suit openings, current suit eligibility, used abilities, ace limits, delayed cards, persistent inflation, permanent Code rules, Ponzi claims, alliances, royal loans, assigned Doppelganger roles, and several scoring times.

Three or four players also create different conditions. A four-player deal leaves **24 cards** in the initial draw pile; a three-player deal leaves **44**. There are more opponents to negotiate with and more opportunities for out-of-turn effects at four players. The versions should be tested separately.

**Recommended aids:** a suit/ability reference, a visible round counter, “used this round” markers, and separate current-game and cumulative-match scores. Record each game’s negative result in the match total once; do not also reapply it as a fresh debt unless you intentionally design a separate debt mechanic.

## 5. Card-count reality check

Assuming “number cards” means 2 through 10 and “royals” means J, Q, K:

| Quantity | Two-deck total | Design implication |
|---|---:|---|
| All cards | 104 | Opening hands use 60 cards at three players or 80 at four. |
| Number diamonds | 18 | A 13-diamond finish requires 72.2% of this resource. |
| Royal cards | 24 | A 15-royal finish requires 62.5% of all royals. |
| Kings / queens / jacks | 8 each | A five-card rank combo commits most of that rank. |
| Aces | 8 | Special recovery and interruption options are finite but may recycle after reshuffles. |

The diamond threshold is highly concentrated, while royal victory may become much easier if alliance members pool their holdings. As an illustration, three predetermined players’ initial hands together contain at least 15 royals about **38.16%** of the time. They contain at least 13 number diamonds about **13.30%** of the time.

These are exact deal probabilities for 60 cards sampled from a well-shuffled 104-card deck. They do **not** include alliance formation, the all-four-series requirement, card losses, additional draws, or choices about which players ally. They are not probabilities of an immediate win. They show why the pooling rule needs to be explicit before you judge the ending thresholds.

## 6. Remaining card-specific questions

Resolve these in a reference sheet before independent testing:

| Mechanic | Missing decision |
|---|---|
| People / Great People | Does it protect one combo, one series, or several? Against ordinary attacks, kidnapping, assassination, or some combination? What protection remains when made with Doppelganger? |
| Doppelganger | Do two kings represent one substitute card? Can physical cards belong to more than one combo? What happens when just one of the pair is removed? |
| Forger | What price must the attacker pay, using which resource? Can they refuse or be unable to pay? What if the defender has no spades left? |
| Dexter | Is the round limit one assassination or one defense in total? Does its protective mode work outside its owner’s turn? Does “exchange” mean paying a diamond or paying and drawing a replacement? |
| Fate | Which cards return: hand, series, combos, borrowed cards, unused aces, persistent effects? Does the all-suits-opened history reset? |
| Justice / Kidnapper | How are hidden cards selected without exposing the entire hand? Define “open special cards.” |
| Compensation | Are losses cumulative across attacks and suits? Do purchases, sacrifices, and voluntary transfers count? Do partial groups of five carry forward? |
| Exile / Barricade | Where must each queen be attached, and does a queen of hearts elsewhere counter Barricade? |
| Quit / defeated players | What makes someone defeated? Which scores are lost on quitting? Can quitting erase negative points? What happens if fewer than three players remain? |
| Negotiated penalties | Which effects may their creator remove by agreement, when, and with what card movement? Can this undo transferred points or a resolved victory? |

## 7. A practical next version and playtest plan

Do not rebalance every card simultaneously. First produce a version that can be played without asking its author for rulings.

**Before the first test:** define the turn cycle, freeze initiative each round, settle the winning/scoring model, specify card ownership and alliance pooling, publish the attack/response procedure, and fix the direct wording conflicts. Standardize all deck returns.

**For the first balance comparison:** test a bounded Ponzi, reversible Code, and explicit Coup response timing. Preserve the rest of the numerical values initially so you can see which problems survive the clarification. If you test alternatives such as overpayment in combat or a different Coup reward, change one major rule at a time.

Run separate three-player and four-player sessions. Rotate seats and players where practical. Record:

- Game duration and number of rounds.
- Which ending condition actually occurred.
- Scores by source: diamonds, spade royals, recurring income, combat, and transfers.
- Every disputed rule and the ruling used.
- Turns with no meaningful option and attacks blocked only by exact matching.
- When major combos appeared, how long they survived, and whether victims recovered.
- Whether the initiative leader, first major-combo owner, or an alliance controlled the finish.

Finally, give the rules to a group without teaching them verbally. Their questions will reveal where the document still depends on knowledge you have as its designer.

**My honest verdict:** the suit economy and exposed-card decisions are worth developing. The present version puts too much weight on ambiguous timing, permanent effects, and unclear ownership. Fix those foundations first; otherwise, a win may reflect the group’s interpretation of the rules more than the players’ decisions.
