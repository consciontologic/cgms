package matchstore

import (
	"encoding/json"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"slices"
	"strings"
	"testing"
)

func envelopeFixture(t *testing.T) Envelope {
	t.Helper()
	s, e := game.NewState(3, "adapter-game")
	if e != nil {
		t.Fatal(e)
	}
	s.Order = []int{1, 2, 3}
	m, e := game.NewMatchLifecycle(s, 3)
	if e != nil {
		t.Fatal(e)
	}
	out, e := NewEnvelope(m, "accepted-rules-hash")
	if e != nil {
		t.Fatal(e)
	}
	return out
}
func TestEnvelopePinsAndDetachedCheckpoint(t *testing.T) {
	e := envelopeFixture(t)
	for _, field := range []string{"schema", "engine", "rules"} {
		n := e.Clone()
		switch field {
		case "schema":
			n.Schema = "future"
		case "engine":
			n.EngineVersion = "future"
		case "rules":
			n.RulesHash = ""
		}
		if n.Validate() == nil {
			t.Fatal(field)
		}
	}
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var n Envelope
	if err = json.Unmarshal(b, &n); err != nil || n.Validate() != nil {
		t.Fatal(err)
	}
	n.Match.Game.Board.Players[0].Confined = true
	if e.Match.Game.Board.Players[0].Confined {
		t.Fatal("checkpoint alias")
	}
}
func TestActorCannotInvokeServerOrImpersonate(t *testing.T) {
	e := envelopeFixture(t)
	for _, kind := range []string{"begin-turn", "begin-round", "resume-automatic", "finalize"} {
		_, err := applyRequest(e, 1, Request{Operation: &game.LifecycleOperation{ID: "x", GameID: e.Match.Game.Board.GameID, Actor: 1, Kind: kind}})
		if err == nil {
			t.Fatal(kind)
		}
	}
	_, err := applyRequest(e, 1, Request{Operation: &game.LifecycleOperation{ID: "x", GameID: e.Match.Game.Board.GameID, Actor: 2, Kind: "departure-choice"}})
	if err == nil {
		t.Fatal("impersonation")
	}
	for _, kind := range []string{"resolve-ready", "purchase-outcome", "ability-outcome"} {
		_, err := applyRequest(e, 1, Request{Command: &game.Command{ID: "x", GameID: e.Match.Game.Board.GameID, Actor: 1, Kind: kind}})
		if err == nil {
			t.Fatal(kind)
		}
	}
}
func TestServerTurnAndActorEndPreserveCheckpoint(t *testing.T) {
	e := envelopeFixture(t)
	n, err := applyServerRequest(e, ServerRequest{ID: "begin", GameID: e.Match.Game.Board.GameID, Kind: "operation", Operation: &game.LifecycleOperation{ID: "begin", GameID: e.Match.Game.Board.GameID, Actor: 1, Kind: "begin-turn"}})
	if err != nil {
		t.Fatal(err)
	}
	if !n.Match.Game.TurnStarted || e.Match.Game.TurnStarted {
		t.Fatal("turn/alias")
	}
	n, err = applyRequest(n, 1, Request{Operation: &game.LifecycleOperation{ID: "end", GameID: e.Match.Game.Board.GameID, Actor: 1, Kind: "end-turn"}})
	if err != nil {
		t.Fatal(err)
	}
	if n.Match.Game.TurnStarted {
		t.Fatal("end failed")
	}
}
func TestSecurePermutationSurvivesRestartAndRejectsRetarget(t *testing.T) {
	e := envelopeFixture(t)
	n, p, err := SecurePermutation(e, "draw/1", []string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(n)
	var restored Envelope
	_ = json.Unmarshal(b, &restored)
	_, q, err := SecurePermutation(restored, "draw/1", []string{"c", "a", "b"})
	if err != nil || game.Digest(p) != game.Digest(q) {
		t.Fatal("restart resampled", err)
	}
	if _, _, err = SecurePermutation(restored, "draw/1", []string{"a", "b", "d"}); err == nil {
		t.Fatal("retarget accepted")
	}
	p[0] = "mutation"
	if restored.Chance["draw/1"][0] == "mutation" {
		t.Fatal("chance aliases")
	}
}

func TestAdapterTransferSerializationAndPartyProjection(t *testing.T) {
	e := envelopeFixture(t)
	var err error
	e.Match.Game.Ledger, err = game.SettleReceipts(e.Match.Game.Ledger, []game.FinancialReceipt{{Seat: 0, Amount: game.IntAmount(10)}})
	if err != nil {
		t.Fatal(err)
	}
	e.Match.Game.Ledger, err = e.Match.Game.Ledger.Charge("recipient-debt", 1, 2, game.IntAmount(6))
	if err != nil {
		t.Fatal(err)
	}
	p, err := SeatProjection(e, 1)
	if err != nil || len(p.Debts) != 0 || p.Cash.Cmp(game.IntAmount(10)) != 0 {
		t.Fatal("unauthorized debt projection", err)
	}
	n, err := applyRequest(e, 1, Request{Operation: &game.LifecycleOperation{ID: "transfer", GameID: e.Match.Game.Board.GameID, Kind: "voluntary-transfer", Actor: 1, Recipient: 2, Amount: game.IntAmount(10)}})
	if err != nil || n.Match.Game.Automatic == nil {
		t.Fatal("lost accounting cursor", err)
	}
	for i := 0; n.Match.Game.Automatic != nil; i++ {
		if i > 30 {
			t.Fatal("did not settle")
		}
		data, err := json.Marshal(n)
		if err != nil {
			t.Fatal(err)
		}
		var restored Envelope
		if err = json.Unmarshal(data, &restored); err != nil {
			t.Fatal(err)
		}
		id := fmt.Sprintf("resume/%d", i)
		n, err = applyServerRequest(restored, ServerRequest{ID: id, GameID: restored.Match.Game.Board.GameID, Kind: "operation", Operation: &game.LifecycleOperation{ID: id, GameID: restored.Match.Game.Board.GameID, Kind: "resume-automatic", Budget: 1}})
		if err != nil {
			t.Fatal(err)
		}
	}
	for seat, want := range []int64{0, 4, 6} {
		if n.Match.Game.Ledger.Cash[seat].Cmp(game.IntAmount(want)) != 0 {
			t.Fatal("wrong exact settlement")
		}
	}
	if e.Match.Game.Ledger.Cash[0].Cmp(game.IntAmount(10)) != 0 {
		t.Fatal("input mutated")
	}
}
func TestServerDealRetainsPrivateChance(t *testing.T) {
	e := envelopeFixture(t)
	n, err := applyServerRequest(e, ServerRequest{ID: "deal", GameID: e.Match.Game.Board.GameID, Kind: "deal"})
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Chance) != 1 || len(e.Chance) != 0 {
		t.Fatal("missing or aliased random outcome")
	}
	p, err := SeatProjection(n, 1)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(p)
	if strings.Contains(string(data), "chance") || strings.Contains(string(data), "rules_hash") {
		t.Fatal("private envelope projected")
	}
}

func TestAdapterNullificationNextGameDoesNotRetarget(t *testing.T) {
	e := envelopeFixture(t)
	var err error
	for seat := 1; seat <= 3; seat++ {
		e, err = applyRequest(e, seat, Request{Operation: &game.LifecycleOperation{ID: fmt.Sprintf("consent/%d", seat), GameID: "adapter-game", Kind: "nullify", Actor: seat}})
		if err != nil {
			t.Fatal(err)
		}
	}
	e, err = applyServerRequest(e, ServerRequest{ID: "next", GameID: "adapter-game", Kind: "next-game", NextGameID: "adapter-game-2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(e.Match.Instances) != 2 || e.Match.Game.Board.GameID != "adapter-game-2" {
		t.Fatal("identity lost")
	}
	if _, err = applyRequest(e, 1, Request{Operation: &game.LifecycleOperation{ID: "fresh", GameID: "adapter-game", Kind: "nullify", Actor: 1}}); err == nil {
		t.Fatal("old intent retargeted")
	}
	if _, err = applyServerRequest(e, ServerRequest{ID: "old", GameID: "adapter-game", Kind: "deal"}); err == nil {
		t.Fatal("old scheduler retargeted")
	}
	if _, err = applyServerRequest(e, ServerRequest{ID: "nextagain", GameID: "adapter-game-2", Kind: "next-game", NextGameID: "adapter-game"}); err == nil {
		t.Fatal("old game reused")
	}
}
func TestServerAutomaticRejectsUnboundedBudget(t *testing.T) {
	e := envelopeFixture(t)
	for _, budget := range []int{0, -1, 10001} {
		_, err := applyServerRequest(e, ServerRequest{ID: "resume", GameID: "adapter-game", Kind: "operation", Operation: &game.LifecycleOperation{ID: "resume", GameID: "adapter-game", Kind: "resume-automatic", Budget: budget}})
		if err == nil {
			t.Fatal("unbounded continuation")
		}
	}
}

func TestClosedJusticeUsesServerRandomnessAndKeepsPendingDecision(t *testing.T) {
	e := envelopeFixture(t)
	s := e.Match.Game.Board
	for _, x := range []struct {
		id   string
		seat int
		zone game.Zone
	}{{"deck-1-clubs-08", 1, game.Series}, {"deck-1-diamonds-02", 1, game.Series}, {"deck-1-hearts-08", 2, game.Series}, {"deck-1-spades-02", 2, game.Hand}, {"deck-1-spades-03", 3, game.Hand}} {
		var err error
		s, err = s.Move([]string{x.id}, x.seat, x.zone, true)
		if err != nil {
			t.Fatal(err)
		}
	}
	s.Players[1].History = []game.Suit{game.Hearts, game.Clubs, game.Spades, game.Diamonds}
	s.Players[2].History = slices.Clone(s.Players[1].History)
	queens := []string{"deck-1-hearts-12", "deck-1-clubs-12", "deck-1-spades-12", "deck-1-diamonds-12"}
	s, err := s.Move(queens, 1, game.Formation, true)
	if err != nil {
		t.Fatal(err)
	}
	for i := range s.Cards {
		if slices.Contains(queens, s.Cards[i].Card.ID) {
			s.Cards[i].Allocation = "justice"
		}
	}
	e.Match.Game.Board = s
	e.Match.Game.TurnStarted = true
	e, err = applyRequest(e, 1, Request{Command: &game.Command{ID: "justice", GameID: s.GameID, Actor: 1, Kind: "justice", AceID: queens[0]}})
	if err != nil {
		t.Fatal(err)
	}
	for _, seat := range []int{2, 3} {
		e, err = applyRequest(e, seat, Request{Command: &game.Command{ID: fmt.Sprintf("pass/%d", seat), GameID: s.GameID, Actor: seat, Kind: "pass", WindowID: "justice"}})
		if err != nil {
			t.Fatal(err)
		}
	}
	if e.Match.Game.Board.Pending == nil || e.Match.Game.Board.Pending.Decision == nil {
		t.Fatal("missing decision")
	}
	c := game.Command{ID: "closed", GameID: s.GameID, Actor: 1, Kind: "decision-closed", WindowID: "justice", DecisionID: e.Match.Game.Board.Pending.Decision.ID}
	bad := c
	bad.RandomWords = []string{"0"}
	if _, err = applyRequest(e, 1, Request{Command: &bad}); err == nil {
		t.Fatal("client controlled random result")
	}
	n, err := applyRequest(e, 1, Request{Command: &c})
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Chance) != 1 || n.Match.Game.Board.Pending == nil || n.Match.Game.Board.Pending.Decision == nil {
		t.Fatal("lost chance or next decision")
	}
	p, err := SeatProjection(n, 3)
	if err != nil || len(p.Board.KnownReserved) > 0 {
		t.Fatal("reserved hidden identity leaked", err)
	}
}
