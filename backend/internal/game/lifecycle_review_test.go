package game

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestPromiseCommitRejectsRehashedForgedSettlement(t *testing.T) {
	b := promisedBook(t, IntAmount(10), "fixed")
	l := promiseLedger(t)
	c, e := b.PreparePayment("p", "forensic-test", 0, l)
	if e != nil {
		t.Fatal(e)
	}
	c.Settlement, _, e = ResumeSettlement(c.Settlement, 8, true)
	if e != nil {
		t.Fatal(e)
	}
	c.Settlement.Ledger.Scores[1] = c.Settlement.Ledger.Scores[1].Add(IntAmount(999))
	c.Settlement.Ledger.Cash[1] = c.Settlement.Ledger.Cash[1].Add(IntAmount(999))
	c.Settlement.Committed = c.Settlement.Ledger.Clone()
	c.Settlement.Digest = c.Settlement.hash()
	if _, _, e = b.CommitPayment(c, l); e == nil {
		t.Fatal("rehashed forged cursor manufactured money")
	}
}

func TestLifecycleFinancialOperationsRequirePaymentBeforeFinish(t *testing.T) {
	l := lifecycleFixture(t)
	var e error
	operation := func(id, kind string, actor int, amount Amount, accept bool) {
		t.Helper()
		l, e = l.ApplyOperation(LifecycleOperation{ID: id, GameID: l.Board.GameID, Kind: kind, Actor: actor, PromiseID: "shared", Recipient: 2, AwardID: "test-award", Mode: "fixed", Amount: amount, Accept: accept})
		if e != nil {
			t.Fatal(e)
		}
	}
	operation("offer", "promise-offer", 1, IntAmount(10), false)
	operation("accept", "promise-accept", 2, Amount{}, false)
	l.Promises, e = l.Promises.RecordAward(PromiseAward{ID: "test-award", Payer: 0, Amount: IntAmount(10)})
	if e != nil {
		t.Fatal(e)
	}
	l.Ledger, e = SettleReceipts(l.Ledger, []FinancialReceipt{{Seat: 0, Amount: IntAmount(10)}})
	if e != nil {
		t.Fatal(e)
	}
	l, e = l.closeOrdinary("round-limit", 0)
	if e != nil {
		t.Fatal(e)
	}
	for l.Automatic != nil {
		l, e = l.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	operation("reduced", "promise-final-offer", 1, IntAmount(4), false)
	operation("answer", "promise-final-answer", 2, Amount{}, true)
	if _, e = l.FinishSettlement(1); e == nil {
		t.Fatal("acceptance alone is not payment")
	}
	operation("payment", "promise-pay", 1, Amount{}, false)
	if l.Automatic == nil {
		t.Fatal("payment must create durable settlement")
	}
	for l.Automatic != nil {
		l, e = l.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	if l.Promises.Promises[0].Paid.Cmp(IntAmount(4)) != 0 || l.Ledger.Cash[1].Cmp(IntAmount(4)) != 0 {
		t.Fatal("actual delivery missing")
	}
	l, e = l.FinishSettlement(1)
	if e != nil {
		t.Fatal(e)
	}
}

func TestMatchNullificationOperationKeepsRetryIdentity(t *testing.T) {
	s, _ := NewState(3, "void-origin")
	m, e := NewMatchLifecycle(s, 3)
	if e != nil {
		t.Fatal(e)
	}
	op := LifecycleOperation{ID: "consent", GameID: s.GameID, Kind: "nullify", Actor: 1}
	n, e := m.ApplyOperation(op)
	if e != nil {
		t.Fatal(e)
	}
	if n.Game.Ledger.Phase == "void" {
		t.Fatal("single assent nullified")
	}
	for seat := 2; seat <= 3; seat++ {
		n, e = n.ApplyOperation(LifecycleOperation{ID: fmt.Sprintf("consent/%d", seat), GameID: s.GameID, Kind: "nullify", Actor: seat})
		if e != nil {
			t.Fatal(e)
		}
	}
	if n.Game.Ledger.Phase != "void" {
		t.Fatal("all individual consents did not nullify")
	}
	twice, e := n.ApplyOperation(op)
	if e != nil || Digest(twice) != Digest(n) {
		t.Fatal("void retry changed result", e)
	}
	n, e = n.NextGame("void-replacement")
	if e != nil {
		t.Fatal(e)
	}
	if len(n.CommandHistory) != 1 || n.CommandHistory[0].LifecycleCommands[op.ID] == "" {
		t.Fatal("void command history lost")
	}
}

func TestMatchCompletedGameDuplicateResolvesBeforeCurrentGameAdmission(t *testing.T) {
	s, _ := NewState(3, "historical-command")
	s.Order = []int{1, 2, 3}
	m, e := NewMatchLifecycle(s, 2)
	if e != nil {
		t.Fatal(e)
	}
	op := LifecycleOperation{ID: "original-turn", GameID: s.GameID, Kind: "begin-turn", Actor: 1}
	m, e = m.ApplyOperation(op)
	if e != nil {
		t.Fatal(e)
	}
	m.Game, e = m.Game.closeOrdinary("round-limit", 0)
	if e != nil {
		t.Fatal(e)
	}
	for m.Game.Automatic != nil {
		m.Game, e = m.Game.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	for seat := 1; seat <= 3; seat++ {
		m.Game, e = m.Game.FinishSettlement(seat)
		if e != nil {
			t.Fatal(e)
		}
	}
	m.Game, e = m.Game.Finalize()
	if e != nil {
		t.Fatal(e)
	}
	m, e = m.NextGame("later-game")
	if e != nil {
		t.Fatal(e)
	}
	before := Digest(m)
	retry, e := m.ApplyOperation(op)
	if e != nil {
		t.Fatalf("N05 committed original-game retry must resolve before new-game rejection: %v", e)
	}
	if Digest(retry) != before {
		t.Fatal("old duplicate mutated current game")
	}
	op.Actor = 2
	rejected, e := m.ApplyOperation(op)
	if e == nil || Digest(rejected) != before {
		t.Fatal("changed old-game retry accepted or mutated state")
	}
	op.ID = "never-committed"
	rejected, e = m.ApplyOperation(op)
	if e == nil || Digest(rejected) != before {
		t.Fatal("new stale intent retargeted current game")
	}
}

func TestLifecycleTransferIntegratedReferenceAndSerializedResume(t *testing.T) {
	l := lifecycleFixture(t)
	var err error
	l.Ledger, err = SettleReceipts(l.Ledger, []FinancialReceipt{{Seat: 0, Amount: IntAmount(10)}})
	if err != nil {
		t.Fatal(err)
	}
	l.Ledger, err = l.Ledger.Charge("recipient-debt", 1, 2, IntAmount(6))
	if err != nil {
		t.Fatal(err)
	}
	before := Digest(l)
	op := LifecycleOperation{ID: "transfer", GameID: l.Board.GameID, Kind: "voluntary-transfer", Actor: 1, Recipient: 2, Amount: IntAmount(10)}
	n, err := l.ApplyOperation(op)
	if err != nil {
		t.Fatal(err)
	}
	if Digest(l) != before || n.Automatic == nil {
		t.Fatal("admission mutation or lost continuation")
	}
	oracle := finishCursor(t, n.Automatic.clone(), false)
	for n.Automatic != nil {
		data, e := json.Marshal(n)
		if e != nil {
			t.Fatal(e)
		}
		var restored Lifecycle
		if e = json.Unmarshal(data, &restored); e != nil {
			t.Fatal(e)
		}
		n, e = restored.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	if !reflect.DeepEqual(n.Ledger, oracle.Ledger) {
		t.Fatal("integrated accounting differs from unbatched reference")
	}
	// Ten leaves payer, six reaches creditor, four remains with recipient.
	for seat, want := range []int64{0, 4, 6} {
		if n.Ledger.Cash[seat].Cmp(IntAmount(want)) != 0 {
			t.Fatalf("seat %d cash", seat)
		}
	}
	retry, e := n.ApplyOperation(op)
	if e != nil || Digest(retry) != Digest(n) {
		t.Fatal("retry duplicated transfer", e)
	}
	op.Amount = IntAmount(9)
	reject, e := n.ApplyOperation(op)
	if e == nil || Digest(reject) != Digest(n) {
		t.Fatal("conflicting transfer retry mutated accounting")
	}
}

func TestLifecycleDepartureRetainsDebtAndUnanimousVoidRestoresStart(t *testing.T) {
	l := lifecycleFixture(t)
	m, e := NewMatchLifecycle(l.Board, 2)
	if e != nil {
		t.Fatal(e)
	}
	m.Game.RoundClosed = true
	m.Game.TurnIndex = 2
	m.Game.Board.Active = 3
	m.Game.Ledger, e = m.Game.Ledger.Charge("departed-debt", 0, 2, IntAmount(10))
	if e != nil {
		t.Fatal(e)
	}
	m.Game.Ledger, e = SettleReceipts(m.Game.Ledger, []FinancialReceipt{{Seat: 0, Amount: IntAmount(4)}, {Seat: 1, Amount: IntAmount(9)}})
	if e != nil {
		t.Fatal(e)
	}
	m.Game, e = m.Game.DepartSeats([]int{1, 2})
	if e != nil {
		t.Fatal(e)
	}
	for m.Game.Automatic != nil {
		m.Game, e = m.Game.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	if m.Game.Ending != "departures" || m.Game.Ledger.Scores[0].Cmp(IntAmount(-6)) != 0 || m.Game.Ledger.Scores[1].Sign() != 0 || m.Game.Ledger.Debts[0].Remaining.Cmp(IntAmount(6)) != 0 {
		t.Fatal("departure rewrote unpaid obligation or retained positive score")
	}
	before := Digest(m)
	rejected, e := m.Nullify([]int{3})
	if e == nil || Digest(rejected) != before {
		t.Fatal("departed consent silently omitted")
	}
	n, e := m.Nullify([]int{1, 2, 3})
	if e != nil {
		t.Fatal(e)
	}
	if len(n.Game.Ledger.Debts) != 0 || n.Game.Ledger.Cash[2].Sign() != 0 || n.Game.Ledger.Scores[2].Sign() != 0 || n.Game.Promises.Phase != "void" {
		t.Fatal("nullification failed to restore pregame financial state")
	}
	n, e = n.NextGame("replacement-departures")
	if e != nil {
		t.Fatal(e)
	}
	if len(n.Game.Ledger.Completed) != 0 || len(n.Instances) != 2 {
		t.Fatal("void consumed finalized game slot")
	}
}

func TestLifecycleCarriedForgivenessCoupAndVoidProvenance(t *testing.T) {
	ledger := NewFinancialLedger("prior", 3)
	var e error
	ledger, e = ledger.Charge("carried", 0, 1, IntAmount(10))
	if e != nil {
		t.Fatal(e)
	}
	ledger, e = ledger.CloseOrdinary(nil)
	if e != nil {
		t.Fatal(e)
	}
	ledger = finalizeTest(t, ledger)
	original := Digest(ledger.Completed)
	ledger, e = ledger.NextGame("correction-game")
	if e != nil {
		t.Fatal(e)
	}
	board, _ := NewState(3, ledger.GameID)
	board.Order = []int{1, 2, 3}
	life, e := NewLifecycle(board, ledger)
	if e != nil {
		t.Fatal(e)
	}
	m := MatchLifecycle{GameLimit: 3, Game: life, Instances: []string{"prior", ledger.GameID}, StartLedger: ledger.Clone()}
	op := LifecycleOperation{ID: "forgiveness", GameID: ledger.GameID, Kind: "forgive", Actor: 2, DebtID: "carried", Amount: IntAmount(10), Seats: []int{2}}
	m, e = m.ApplyOperation(op)
	if e != nil {
		t.Fatal(e)
	}
	if m.Game.Ledger.Cash[0].Sign() != 0 || len(m.Game.Ledger.Corrections) != 1 || !m.Game.Ledger.Corrections[0].MatchLevel {
		t.Fatal("carried correction became spendable or lost lineage")
	}
	void, e := m.Nullify([]int{1, 2, 3})
	if e != nil {
		t.Fatal(e)
	}
	if void.Game.Ledger.Debts[0].Remaining.Cmp(IntAmount(10)) != 0 || len(void.Game.Ledger.Corrections) != 0 || Digest(void.Game.Ledger.Completed) != original {
		t.Fatal("void did not restore carried provenance")
	}
	m.Game.Board, e = m.Game.Board.Move([]string{"deck-1-diamonds-12", "deck-2-diamonds-12", "deck-1-hearts-12", "deck-2-hearts-12"}, 1, Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	m.Game.Board.Players[0].History = []Suit{Clubs}
	m.Game, e = m.Game.DeclareCoup(Command{ID: "corrected-coup", GameID: ledger.GameID, Kind: "coup", Actor: 1})
	if e != nil {
		t.Fatal(e)
	}
	for m.Game.Automatic != nil {
		m.Game, e = m.Game.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	if m.Game.Ledger.MatchScores()[0].Cmp(IntAmount(50)) != 0 || m.Game.Ledger.Cash[0].Cmp(IntAmount(50)) != 0 || Digest(m.Game.Ledger.Completed) != original {
		t.Fatal("N06: -10+10+50 must be 50 without rewriting prior game")
	}
}

func TestLifecycleCannotAssertOtherPlayersFinancialConsent(t *testing.T) {
	s, _ := NewState(3, "individual-consent")
	m, e := NewMatchLifecycle(s, 2)
	if e != nil {
		t.Fatal(e)
	}
	forged := LifecycleOperation{ID: "forged", GameID: s.GameID, Kind: "nullify", Actor: 1, Seats: []int{1, 2, 3}}
	n, e := m.ApplyOperation(forged)
	if e == nil || Digest(n) != Digest(m) {
		t.Fatal("one actor fabricated unanimous nullification")
	}
	l := lifecycleFixture(t)
	l.Ledger, e = l.Ledger.Charge("system", 0, -1, IntAmount(10))
	if e != nil {
		t.Fatal(e)
	}
	op := LifecycleOperation{ID: "forgive-forged", GameID: l.Board.GameID, Kind: "forgive", Actor: 1, Seats: []int{1, 2, 3}, DebtID: "system", Amount: IntAmount(10)}
	rejected, e := l.ApplyOperation(op)
	if e == nil || Digest(rejected) != Digest(l) {
		t.Fatal("one actor fabricated unanimous system forgiveness")
	}
}

func TestSystemForgivenessRequiresRecordedIndividualConsents(t *testing.T) {
	l := lifecycleFixture(t)
	var e error
	l.Ledger, e = l.Ledger.Charge("unanimous", 0, -1, IntAmount(10))
	if e != nil {
		t.Fatal(e)
	}
	// A departed participant remains part of unanimity and can explicitly assent.
	l.Board.Players[2].Departed = true
	l.Ledger.Departed[2] = true
	for seat := 1; seat <= 3; seat++ {
		op := LifecycleOperation{ID: fmt.Sprintf("assent/%d", seat), GameID: l.Board.GameID, Kind: "forgive", Actor: seat, DebtID: "unanimous", Amount: IntAmount(10)}
		before := Digest(l)
		n, err := l.ApplyOperation(op)
		if err != nil {
			t.Fatal(err)
		}
		if Digest(l) != before {
			t.Fatal("consent aliased predecessor")
		}
		l = n
		again, err := l.ApplyOperation(op)
		if err != nil || Digest(again) != Digest(l) {
			t.Fatal("consent retry changed state", err)
		}
		if seat < 3 && (l.Ledger.Debts[0].Remaining.Cmp(IntAmount(10)) != 0 || l.Ledger.Scores[0].Cmp(IntAmount(-10)) != 0) {
			t.Fatal("premature forgiveness")
		}
		data, _ := json.Marshal(l)
		var restored Lifecycle
		if e = json.Unmarshal(data, &restored); e != nil {
			t.Fatal(e)
		}
		if e = restored.Validate(); e != nil {
			t.Fatal(e)
		}
		l = restored
	}
	if len(l.ForgivenessConsents) != 0 || l.Ledger.Debts[0].Remaining.Sign() != 0 || l.Ledger.Scores[0].Sign() != 0 || l.Ledger.Cash[0].Sign() != 0 {
		t.Fatal("unanimous correction not exact/nonspendable")
	}
}

func TestDepartureChoicesRemainPendingUntilEveryEligibleSeatAnswers(t *testing.T) {
	l := lifecycleFixture(t)
	l.RoundClosed = true
	l.TurnIndex = 2
	l.Board.Active = 3
	op := LifecycleOperation{ID: "leave1", GameID: l.Board.GameID, Kind: "departure-choice", Actor: 1, Accept: true}
	n, e := l.ApplyOperation(op)
	if e != nil {
		t.Fatal(e)
	}
	if n.Board.Players[0].Departed || n.Ending != "" {
		t.Fatal("first choice applied premature departure")
	}
	data, _ := json.Marshal(n)
	var restored Lifecycle
	if e = json.Unmarshal(data, &restored); e != nil {
		t.Fatal(e)
	}
	before := Digest(restored)
	rejected, e := restored.ApplyOperation(LifecycleOperation{ID: "skip", GameID: l.Board.GameID, Kind: "begin-round", Actor: 3})
	if e == nil || Digest(rejected) != before {
		t.Fatal("unanswered decisions became stays")
	}
	for _, seat := range []int{2, 3} {
		restored, e = restored.ApplyOperation(LifecycleOperation{ID: fmt.Sprintf("leave%d", seat), GameID: l.Board.GameID, Kind: "departure-choice", Actor: seat, Accept: true})
		if e != nil {
			t.Fatal(e)
		}
	}
	if restored.Ending != "departures" {
		t.Fatal("simultaneous departure boundary not closed")
	}
	for _, p := range restored.Board.Players {
		if !p.Departed {
			t.Fatal("last choice lost after premature closure")
		}
	}
}

func TestDepartureChoiceExplicitStaysRetryAndCoup(t *testing.T) {
	l := lifecycleFixture(t)
	l.RoundClosed = true
	l.TurnIndex = 2
	l.Board.Active = 3
	op := LifecycleOperation{ID: "stay1", GameID: l.Board.GameID, Kind: "departure-choice", Actor: 1, Accept: false}
	n, e := l.ApplyOperation(op)
	if e != nil {
		t.Fatal(e)
	}
	again, e := n.ApplyOperation(op)
	if e != nil || Digest(again) != Digest(n) {
		t.Fatal("choice retry", e)
	}
	conflict := op
	conflict.Accept = true
	rejected, e := n.ApplyOperation(conflict)
	if e == nil || Digest(rejected) != Digest(n) {
		t.Fatal("changed choice retry accepted")
	}
	for _, seat := range []int{2, 3} {
		n, e = n.ApplyOperation(LifecycleOperation{ID: fmt.Sprintf("stay%d", seat), GameID: l.Board.GameID, Kind: "departure-choice", Actor: seat})
		if e != nil {
			t.Fatal(e)
		}
	}
	if !n.DeparturesResolved || n.Ending != "" {
		t.Fatal("explicit stays not completed")
	}
	for _, p := range n.Board.Players {
		if p.Departed {
			t.Fatal("stay became departure")
		}
	}
	// A legal Coup can close the board while the boundary still awaits others;
	// the already queued departure is not prematurely applied.
	n, e = l.ApplyOperation(LifecycleOperation{ID: "leave1", GameID: l.Board.GameID, Kind: "departure-choice", Actor: 1, Accept: true})
	if e != nil {
		t.Fatal(e)
	}
	n.Board, e = n.Board.Move([]string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-diamonds-12", "deck-2-diamonds-12"}, 2, Hand, false)
	if e != nil {
		t.Fatal(e)
	}
	n.Board.Players[1].History = []Suit{Clubs}
	n, e = n.DeclareCoup(Command{ID: "boundary-coup", GameID: l.Board.GameID, Actor: 2, Kind: "coup"})
	if e != nil {
		t.Fatal(e)
	}
	if n.Ending != "coup" || n.Board.Players[0].Departed {
		t.Fatal("Coup fabricated queued departure")
	}
}

func TestMatchAcceptedOperationDetachesWholeMatchMetadata(t *testing.T) {
	s, _ := NewState(3, "detached-match")
	s.Order = []int{1, 2, 3}
	m, e := NewMatchLifecycle(s, 2)
	if e != nil {
		t.Fatal(e)
	}
	m.CommandHistory = []LifecycleReceiptSet{{GameID: "earlier", CardCommands: map[string]string{"c": "record"}, LifecycleCommands: map[string]string{"l": "record"}}}
	before := Digest(m)
	n, e := m.ApplyOperation(LifecycleOperation{ID: "start", GameID: s.GameID, Kind: "begin-turn", Actor: 1})
	if e != nil {
		t.Fatal(e)
	}
	n.Instances[0] = "changed"
	n.StartLedger.Scores[0] = IntAmount(99)
	n.CommandHistory[0].CardCommands["c"] = "changed"
	n.CommandHistory[0].LifecycleCommands["l"] = "changed"
	if Digest(m) != before {
		t.Fatal("accepted operation aliases predecessor match metadata")
	}
}

func TestMatchNullifyAndDuplicateDetachBaseline(t *testing.T) {
	s, _ := NewState(3, "clone-void")
	m, e := NewMatchLifecycle(s, 2)
	if e != nil {
		t.Fatal(e)
	}
	op := LifecycleOperation{ID: "assent", GameID: s.GameID, Kind: "nullify", Actor: 1}
	m, e = m.ApplyOperation(op)
	if e != nil {
		t.Fatal(e)
	}
	before := Digest(m)
	duplicate, e := m.ApplyOperation(op)
	if e != nil {
		t.Fatal(e)
	}
	duplicate.NullificationConsents[0] = 3
	duplicate.StartLedger.Cash[0] = IntAmount(8)
	if Digest(m) != before {
		t.Fatal("duplicate aliases consent or start ledger")
	}
	void, e := m.Nullify([]int{1, 2, 3})
	if e != nil {
		t.Fatal(e)
	}
	void.Instances[0] = "tampered"
	void.StartLedger.Scores[0] = IntAmount(7)
	if Digest(m) != before {
		t.Fatal("nullification aliases authoritative baseline")
	}
}
