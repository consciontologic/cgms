package sim

import (
	"context"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"strings"
	"testing"
)

func TestGameplayCoupOfferedBeforeDeferredDrawToActiveSeat(t *testing.T) {
	s, _ := game.NewState(3, "atomic-coup")
	s.Round = 1
	s.Turn = 1
	s.Active = 1
	s.Phase = "playing"
	s.Order = []int{1, 2, 3}
	s.Players[0].History = []game.Suit{game.Clubs}
	var e error
	s, e = s.Move([]string{"deck-1-diamonds-12", "deck-2-diamonds-12", "deck-1-hearts-12", "deck-2-hearts-12"}, 1, game.Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	s.DrawQueue = []game.DrawEntitlement{{Seat: 1, Remaining: 2}}
	m, e := game.NewMatchLifecycle(s, 1)
	if e != nil {
		t.Fatal(e)
	}
	m.Game.TurnStarted = true
	j := Job{Population: 3, Games: 1, Policies: []Policy{{Policy: "economic", Version: "v1"}, {Policy: "economic", Version: "v1"}, {Policy: "economic", Version: "v1"}}}
	o := GameplayOptions{Job: j, Root: "4", Pairing: "fixture", BotBudget: 1000}
	streams, e := gameplayStreams(o, 0)
	if e != nil {
		t.Fatal(e)
	}
	choose := func(p string, obs game.Observation, r *randomstream.Stream, b int) (bots.Decision, error) {
		for _, c := range obs.Legal {
			if c.Kind == "coup" {
				return bots.Decision{Policy: p, Command: c}, nil
			}
		}
		t.Fatal("Coup absent")
		return bots.Decision{}, nil
	}
	st := GameplayStep{}
	if e = planGameplayStep(m, streams, choose, o, &st, nil); e != nil {
		t.Fatal(e)
	}
	if st.Command == nil || st.Command.Kind != "coup" || st.Command.Actor != 1 {
		t.Fatalf("Coup bypassed: %s", st.Kind)
	}
}

func TestGameplayFinancialDeclineCachesOnlyEquivalentOpportunity(t *testing.T) {
	s, _ := game.NewState(3, "finance-opportunity")
	s.Round = 1
	s.Turn = 1
	s.Active = 1
	s.Phase = "playing"
	s.Order = []int{1, 2, 3}
	m, e := game.NewMatchLifecycle(s, 1)
	if e != nil {
		t.Fatal(e)
	}
	m.Game.TurnStarted = true
	j := Job{Population: 3, Games: 1, Policies: []Policy{{Policy: "economic", Version: "v1"}, {Policy: "economic", Version: "v1"}, {Policy: "economic", Version: "v1"}}}
	o := GameplayOptions{Job: j, Root: "4", Pairing: "fixture", BotBudget: 1000, financialDeclines: map[int]string{}}
	streams, _ := gameplayStreams(o, 0)
	st := GameplayStep{}
	handled, e := planPlayingFinance(m, streams, o, &st, nil)
	if e != nil || !handled || st.Kind != "policy-wait" || st.Operation.Kind != "decline-financial-opportunity" {
		t.Fatalf("explicit decline missing: %v %v", handled, e)
	}
	before := hash(m)
	next, _, e := applyGameplayStep(m, st)
	if e != nil || hash(next) != before {
		t.Fatal("decline changed gameplay")
	}
	o.financialDeclines[1] = st.OpportunityKey
	st2 := GameplayStep{}
	handled, e = planPlayingFinance(m, streams, o, &st2, nil)
	if e != nil || !handled || st2.Operation.Actor == 1 {
		t.Fatal("equivalent opportunity repeated")
	}
	m.Game.Ledger.Cash[0] = game.IntAmount(1)
	st3 := GameplayStep{}
	handled, e = planPlayingFinance(m, streams, o, &st3, nil)
	if e != nil || !handled || st3.Operation.Actor != 1 {
		t.Fatal("changed cash not reoffered")
	}
}

func TestGameplayVoidReplacementRetainsConfiguredSlotDenominator(t *testing.T) {
	s, _ := game.NewState(3, "match-000000/game-0/instance-0")
	m, e := game.NewMatchLifecycle(s, 2)
	if e != nil {
		t.Fatal(e)
	}
	for seat := 1; seat <= 3; seat++ {
		m, e = m.ApplyOperation(game.LifecycleOperation{GameID: s.GameID, ID: fmt.Sprint("consent", seat), Actor: seat, Kind: "nullify"})
		if e != nil {
			t.Fatal(e)
		}
	}
	j := Job{Population: 3, Games: 2}
	st := GameplayStep{}
	e = planGameplayStep(m, nil, nil, GameplayOptions{Job: j}, &st, nil)
	if e != nil || st.Kind != "next-game" || st.NextID == s.GameID {
		t.Fatal("void replacement identity", e)
	}
	m, _, e = applyGameplayStep(m, st)
	if e != nil {
		t.Fatal(e)
	}
	o := gameplayOutcome(GameplayRun{Job: j, Match: m, ExitCode: 5, Trace: []GameplayStep{{Kind: "deal", GameID: s.GameID}, {Kind: "deal", GameID: m.Game.Board.GameID}}})
	if o.StartedInstances != 2 || o.VoidedGames != 1 || o.StartedGames != 1 || o.NotStartedGameSlots != 1 || o.BoardEndedGames != 0 {
		t.Fatalf("void polluted slots: %+v", o)
	}
}

func TestGameplayCanonicalArtifactBudgetRetainsPrefix(t *testing.T) {
	j := Job{Population: 3, Games: 1, Policies: []Policy{{Policy: "heuristic", Version: "v1"}, {Policy: "heuristic", Version: "v1"}, {Policy: "heuristic", Version: "v1"}}}
	choose := func(_ string, o game.Observation, _ *randomstream.Stream, _ int) (bots.Decision, error) {
		return bots.Decision{Policy: "oversized-policy-fixture", Command: game.Command{GameID: o.GameID, Actor: o.Seat, Kind: "end-turn"}, Reasons: []string{strings.Repeat("x", canonical.MaxBytes)}}, nil
	}
	r := RunGameplay(context.Background(), GameplayOptions{Job: j, Root: "3", Pairing: "fixture", MaxSteps: 30, BotBudget: 1000, MaxOutputBytes: 1 << 30, Policy: choose, PolicyLabel: "oversized-policy-fixture"})
	if r.ExitCode != 5 || !strings.Contains(r.Reason, "output") {
		t.Fatalf("did not stop artifact budget: %d %s", r.ExitCode, r.Reason)
	}
	if _, e := canonical.Marshal(r); e != nil {
		t.Fatal("committed artifact unrepresentable", e)
	}
	if e := ReplayGameplay(r); e != nil {
		t.Fatal(e)
	}
}

func TestOpportunityEconomyProjectionPartyOnlyAndMatchHorizon(t *testing.T) {
	s, _ := game.NewState(3, "economy-context")
	m, e := game.NewMatchLifecycle(s, 3)
	if e != nil {
		t.Fatal(e)
	}
	m.Game.Ledger.Cash[0] = game.IntAmount(2)
	m.Game.Ledger, e = m.Game.Ledger.Charge("own", 0, -1, game.IntAmount(7))
	if e != nil {
		t.Fatal(e)
	}
	a := projectPolicyEconomy(m, 1)
	if a.FutureGames != 2 || a.Owed.Cmp(game.IntAmount(5)) != 0 {
		t.Fatal("wrong own debt/horizon projection", a)
	}
	n := m.Clone()
	n.Game.Ledger, e = n.Game.Ledger.Charge("not-own", 2, 1, game.IntAmount(97))
	if e != nil {
		t.Fatal(e)
	}
	b := projectPolicyEconomy(n, 1)
	if game.Digest(a) != game.Digest(b) {
		t.Fatal("other debt crossed board policy boundary")
	}
	n.Game.Ledger.Cash[0] = game.IntAmount(100)
	if a.Cash.Sign() != 0 {
		t.Fatal("scalar projection aliased mutable match")
	}
}

func TestOpportunityExplicitPolicyHookPreserved(t *testing.T) {
	j := Job{Population: 3, Games: 1, Policies: []Policy{{Policy: "opportunity", Version: "v1"}, {Policy: "opportunity", Version: "v1"}, {Policy: "opportunity", Version: "v1"}}}
	calls := 0
	hook := func(p string, o game.Observation, r *randomstream.Stream, b int) (bots.Decision, error) {
		calls++
		return restrictedEndTurn(p, o, r, b)
	}
	result := RunGameplay(context.Background(), GameplayOptions{Job: j, Root: "1", MaxSteps: 10, BotBudget: 100000, Policy: hook, PolicyLabel: "scripted-end-turn-test-only"})
	if calls == 0 {
		t.Fatal("explicit hook bypassed", result.Reason)
	}
}

func TestOpportunityAdapterDecisionIgnoresUnrelatedLedger(t *testing.T) {
	s, _ := game.NewState(3, "adapter-private")
	m, e := game.NewMatchLifecycle(s, 3)
	if e != nil {
		t.Fatal(e)
	}
	m.Game.Ledger, e = m.Game.Ledger.Charge("own", 0, -1, game.IntAmount(9))
	if e != nil {
		t.Fatal(e)
	}
	n := m.Clone()
	n.Game.Ledger, e = n.Game.Ledger.Charge("private-other", 2, 1, game.IntAmount(99))
	if e != nil {
		t.Fatal(e)
	}
	o := game.Observation{GameID: s.GameID, Seat: 1, Legal: []game.Command{{GameID: s.GameID, Actor: 1, Kind: "ringleader"}, {GameID: s.GameID, Actor: 1, Kind: "open-series", Suit: game.Clubs}}}
	a, e := bots.ChooseGameplayWithEconomy("opportunity@v1", o, projectPolicyEconomy(m, 1), nil, 100)
	if e != nil {
		t.Fatal(e)
	}
	b, e := bots.ChooseGameplayWithEconomy("opportunity@v1", o, projectPolicyEconomy(n, 1), nil, 100)
	if e != nil || game.Digest(a) != game.Digest(b) {
		t.Fatal("unrelated ledger changed policy decision", e)
	}
	if a.Command.Kind != "ringleader" {
		t.Fatal("fixture lacks active own-debt context")
	}
}

func TestOpportunityMenuBudgetForwardingAndDefaultEquivalence(t *testing.T) {
	j := Job{Population: 3, Games: 1, Policies: []Policy{{Policy: "economic", Version: "v1"}, {Policy: "pressure", Version: "v1"}, {Policy: "economic", Version: "v1"}}}
	b := Budgets{MaxTransitions: 10, MaxBotOperations: 100000, MaxOutputBytes: 1 << 24}
	a := runFullGameplay(context.Background(), j, "1", "budget-fixture", b)
	b.MaxMenuOperations = 100000
	equal := runFullGameplay(context.Background(), j, "1", "budget-fixture", b)
	if a.Gameplay == nil || equal.Gameplay == nil || game.Digest(a.Gameplay) != game.Digest(equal.Gameplay) {
		t.Fatal("explicit historical cap changed transcript")
	}
	b.MaxMenuOperations = 1
	limited := runFullGameplay(context.Background(), j, "1", "budget-fixture", b)
	if limited.Gameplay == nil || limited.Gameplay.Reason != game.ErrMenuBudget.Error() || len(limited.Gameplay.Trace) >= len(a.Gameplay.Trace) {
		t.Fatal("experiment menu budget not enforced")
	}
	b.MaxMenuOperations = 1000000
	continued := continueFullGameplay(context.Background(), j, "1", "budget-fixture", b, *limited.Gameplay)
	if continued.Gameplay == nil || continued.Gameplay.Reason == game.ErrMenuBudget.Error() || len(continued.Gameplay.Trace) <= len(limited.Gameplay.Trace) {
		t.Fatal("continuation ignored explicit larger menu cap")
	}
}
