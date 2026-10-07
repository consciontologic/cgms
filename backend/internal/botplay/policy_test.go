package botplay

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
)

func TestOnlineLazyValidationMatchesEagerChoicesAndHistoricalTies(t *testing.T) {
	board, err := game.NewState(3, "real-v1-ties")
	if err != nil {
		t.Fatal(err)
	}
	board, err = board.Move([]string{"deck-1-clubs-02", "deck-1-hearts-02", "deck-1-spades-02", "deck-1-diamonds-02"}, 1, game.Series, false)
	if err != nil {
		t.Fatal(err)
	}
	board.Players[0].History = []game.Suit{game.Clubs, game.Hearts, game.Spades, game.Diamonds}
	board, err = board.Move([]string{"deck-1-hearts-12", "deck-2-hearts-12"}, 1, game.Hand, false)
	if err != nil {
		t.Fatal(err)
	}
	board, err = game.OpenFormation(board, 1, "public-protector", game.FormationSpec{Kind: "great-people", Cards: []string{"deck-1-hearts-12", "deck-2-hearts-12"}, Protection: &game.FormationProtection{Series: game.Diamonds}})
	if err != nil {
		t.Fatal(err)
	}
	board.Turn++
	board.Round = 2
	board.Order = []int{1, 2, 3}
	board.Players[0].OpeningUsed = false
	menu, err := game.GameplayObservation(board, 1, menuBudget)
	if err != nil || !bots.GameplayV1OrderingHasTies(menu) {
		t.Fatal("fixture must retain actual historical reconfiguration ties", err)
	}
	match, err := game.NewMatchLifecycle(board, 1)
	if err != nil {
		t.Fatal(err)
	}
	match.Game.TurnStarted = true
	collision, err := matchstore.NewEnvelope(match, "rules-test")
	if err != nil {
		t.Fatal(err)
	}
	if _, automatic := matchstore.NextServer(collision, 0); automatic {
		t.Fatal("collision fixture must be an actionable turn")
	}
	for _, position := range []struct {
		name string
		env  matchstore.Envelope
		seat int
	}{
		{"empty-supply", purchaseCycle(t, false), 2},
		{"productive-purchase", purchaseCycle(t, true), 2},
		{"actual-v1-ties", collision, 1},
	} {
		for _, difficulty := range []string{"beginner", "standard", "advanced"} {
			t.Run(position.name+"/"+difficulty, func(t *testing.T) {
				want, beforeErr := suggest(position.env, 0, position.seat, difficulty, true)
				got, afterErr := SuggestOnline(position.env, 0, position.seat, difficulty)
				if beforeErr != nil || afterErr != nil || game.Digest(got) != game.Digest(want) {
					t.Fatal("lazy validation changed the eager policy's exact command", beforeErr, afterErr)
				}
				if position.name == "actual-v1-ties" && (got.Command == nil || got.Command.Kind != "open-formation") {
					t.Fatal("historical tie fixture must select an actual formation reconfiguration")
				}
			})
		}
	}
}

func TestOnlineLazyValidationRejectsHighestScoredInvalidCandidate(t *testing.T) {
	env := purchaseCycle(t, false)
	actions, err := Actions(env, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	invalid := game.LifecycleOperation{ID: "invalid-victory", GameID: env.Match.Game.Board.GameID, Actor: 2, Kind: "coup"}
	if matchstore.ValidateActorRequest(env, 2, matchstore.Request{Operation: &invalid}) == nil {
		t.Fatal("fixture Coup must be rejected by authority")
	}
	actions = append(actions, matchstore.Request{Operation: &invalid})
	for _, difficulty := range []string{"beginner", "standard", "advanced"} {
		raw, err := chooseProjected(env, 2, difficulty, true, actions)
		if err != nil || raw.Operation == nil || raw.Operation.Kind != "coup" {
			t.Fatal("fixture must rank rejected action first", err)
		}
		choice, err := chooseOnline(context.Background(), env, 2, difficulty, append([]matchstore.Request(nil), actions...))
		if err != nil || choice.Operation == nil || choice.Operation.Kind != "end-turn" {
			t.Fatal("invalid highest score did not fall back to validated EndTurn", err)
		}
	}
}

func TestOnlinePlanningCancellationReturnsNoRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	choice, err := SuggestOnlineContext(ctx, purchaseCycle(t, false), 0, 2, "advanced")
	if !errors.Is(err, context.Canceled) || choice.Command != nil || choice.Operation != nil {
		t.Fatal("canceled planning returned an action", err)
	}
}

func purchaseCycle(t *testing.T, supply bool) matchstore.Envelope {
	t.Helper()
	b, err := game.NewState(3, "purchase-cycle")
	if err != nil {
		t.Fatal(err)
	}
	var payment, hand, aces []string
	for _, placed := range b.Cards {
		c := placed.Card
		switch {
		case supply && c.ID == "deck-2-clubs-02":
		case c.Suit == game.Spades && c.Rank >= 2 && c.Rank <= 10 && len(payment) < 13:
			payment = append(payment, c.ID)
		case c.Rank == 1:
			aces = append(aces, c.ID)
		default:
			hand = append(hand, c.ID)
		}
	}
	for _, movement := range []struct {
		cards []string
		seat  int
		zone  game.Zone
	}{{payment, 2, game.Series}, {hand, 1, game.Hand}, {aces, 1, game.ConcealedAce}} {
		b, err = b.Move(movement.cards, movement.seat, movement.zone, true)
		if err != nil {
			t.Fatal(err)
		}
	}
	b.Round, b.Turn, b.Active, b.Order = 2, 6, 2, []int{2, 1, 3}
	for i := range b.Players {
		b.Players[i].History = []game.Suit{game.Spades}
	}
	m, err := game.NewMatchLifecycle(b, 1)
	if err != nil {
		t.Fatal(err)
	}
	m.Game.TurnStarted = true
	e, err := matchstore.NewEnvelope(m, "rules-test")
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func resolveCycleResponses(t *testing.T, env matchstore.Envelope, version *int64) matchstore.Envelope {
	t.Helper()
	for step := 0; step < 8; step++ {
		var err error
		if automatic, ok := matchstore.NextServer(env, *version); ok {
			request := automatic.Request
			request.ID = automatic.ID
			if request.Operation != nil {
				request.Operation.ID = automatic.ID
			}
			env, err = matchstore.ApplyLocalAutomatic(env, request)
		} else if pending := env.Match.Game.Board.Pending; pending != nil {
			if pending.Decision != nil || pending.Cursor >= len(pending.Responders) {
				t.Fatal("unexpected fixture response boundary")
			}
			seat := pending.Responders[pending.Cursor]
			c := game.Command{ID: fmt.Sprintf("pass-%d", *version), GameID: env.Match.Game.Board.GameID, Kind: "pass", Actor: seat, WindowID: pending.ID}
			env, err = matchstore.ApplyLocalRequest(env, seat, matchstore.Request{Command: &c})
		} else {
			return env
		}
		if err != nil {
			t.Fatal(err)
		}
		*version++
	}
	t.Fatal("fixture responses did not finish within their bounded steps")
	return env
}

func TestOnlineSuggestionEndsExactEmptySupplyPurchaseReopenCycle(t *testing.T) {
	for _, difficulty := range []string{"beginner", "standard", "advanced"} {
		t.Run(difficulty, func(t *testing.T) {
			env := purchaseCycle(t, false)
			material := game.Digest(env.Match.Game.Board.Cards)
			version := int64(0)
			// The browser's Advanced policy maximizes quantity and exactly repeats
			// the 13-for-13 cycle. Other frozen tiers can choose losing exchanges;
			// the online progress requirement applies to all three tiers.
			cycles := 0
			if difficulty == "advanced" {
				cycles = 2
			}
			for cycle := 0; cycle < cycles; cycle++ {
				legacy, err := Suggest(env, version, 2, difficulty)
				if err != nil || legacy.Command == nil || legacy.Command.Kind != "purchase" || legacy.Command.Value != 13 || len(legacy.Command.Cards) != 13 {
					t.Fatal("fixture did not reproduce the frozen policy's non-increasing purchase", err)
				}
				env, err = matchstore.ApplyLocalRequest(env, 2, legacy)
				if err != nil {
					t.Fatal(err)
				}
				version++
				env = resolveCycleResponses(t, env, &version)
				reopen, err := Suggest(env, version, 2, difficulty)
				if err != nil || reopen.Command == nil || reopen.Command.Kind != "open-series" || reopen.Command.Suit != game.Spades {
					t.Fatal("fixture did not reopen the purchased Spades", err)
				}
				env, err = matchstore.ApplyLocalRequest(env, 2, reopen)
				if err != nil {
					t.Fatal(err)
				}
				version++
				env = resolveCycleResponses(t, env, &version)
				if game.Digest(env.Match.Game.Board.Cards) != material || env.Match.Game.Board.Turn != 6 {
					t.Fatal("fixture did not return to exactly the same material state and turn")
				}
			}
			before := game.Digest(env)
			choice, err := SuggestOnline(env, version, 2, difficulty)
			if err != nil || choice.Operation == nil || choice.Operation.Kind != "end-turn" {
				t.Fatalf("online bot repeated a material cycle instead of ending its turn: request=%+v error=%v", choice, err)
			}
			if game.Digest(env) != before {
				t.Fatal("policy mutated authority state")
			}
			legal, err := Actions(env, version, 2)
			if err != nil {
				t.Fatal(err)
			}
			humanPurchase := false
			for _, candidate := range legal {
				humanPurchase = humanPurchase || candidate.Command != nil && candidate.Command.Kind == "purchase" && candidate.Command.Value <= len(candidate.Command.Cards)
			}
			if !humanPurchase {
				t.Fatal("online policy safeguard changed human legal options")
			}
		})
	}
}

func TestOnlineSuggestionKeepsAcquisitionThatIncreasesOwnCardCount(t *testing.T) {
	for _, difficulty := range []string{"beginner", "standard", "advanced"} {
		env := purchaseCycle(t, true)
		choice, err := SuggestOnline(env, 0, 2, difficulty)
		if err != nil || choice.Command == nil || choice.Command.Kind != "purchase" || choice.Command.Value <= len(choice.Command.Cards) {
			t.Fatalf("%s missed a strictly increasing legal purchase: request=%+v error=%v", difficulty, choice, err)
		}
	}
}

func TestOnlineProgressGuardPreservesRequiredResponses(t *testing.T) {
	env := purchaseCycle(t, false)
	purchase, err := Suggest(env, 0, 2, "advanced")
	if err != nil {
		t.Fatal(err)
	}
	env, err = matchstore.ApplyLocalRequest(env, 2, purchase)
	if err != nil {
		t.Fatal(err)
	}
	pending := env.Match.Game.Board.Pending
	if pending == nil || len(pending.Responders) != 2 || pending.Responders[0] != 3 || pending.Responders[1] != 1 {
		t.Fatal("fixture must preserve clockwise response order: bot3 then human1")
	}
	baseline, err := Suggest(env, 1, 3, "advanced")
	if err != nil {
		t.Fatal(err)
	}
	online, err := SuggestOnline(env, 1, 3, "advanced")
	if err != nil || online.Command == nil || online.Command.Kind != "pass" || game.Digest(baseline) != game.Digest(online) {
		t.Fatal("required bot response changed", err)
	}
	env, err = matchstore.ApplyLocalRequest(env, 3, online)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = SuggestOnline(env, 2, 3, "advanced"); err != ErrNoDecision {
		t.Fatal("online safeguard answered the following human response", err)
	}
	pass := game.Command{ID: "explicit-human-pass", GameID: env.Match.Game.Board.GameID, Kind: "pass", Actor: 1, WindowID: pending.ID}
	env, err = matchstore.ApplyLocalRequest(env, 1, matchstore.Request{Command: &pass})
	if err != nil {
		t.Fatal(err)
	}
	if _, pending := matchstore.NextServer(env, 3); !pending {
		t.Fatal("purchase advanced without the explicit human answer")
	}
}
