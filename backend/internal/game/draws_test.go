package game

import (
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"reflect"
	"testing"
)

func compensationPosition(t *testing.T) State {
	s := position(t)
	ids := append([]string(nil), s.DrawOrder...)
	s, _ = s.Move(ids, 1, Unassigned, true)
	for _, seat := range []int{2, 3} {
		id := cid(Hearts, 1)
		if seat == 3 {
			for _, c := range Deck() {
				if c.Deck == 2 && c.Suit == Hearts && c.Rank == 1 {
					id = c.ID
				}
			}
		}
		s, _ = s.Move([]string{id}, seat, ConcealedAce, true)
		var e error
		s, e = s.ActivateEffect(AceEffect{ID: id, CardID: id, Kind: "compensation", Source: seat, Target: seat, Custodian: seat, Expiry: "target-departure-or-closure", Losses: 5})
		if e != nil {
			t.Fatal(e)
		}
	}
	s, _ = s.Move([]string{cid(Clubs, 12)}, 0, Draw, true)
	s, _ = ShuffleSupply(s, s.DrawOrder)
	s.Order = []int{3, 2, 1}
	return s
}
func TestDeferredDrawsSeatOrderExhaustion(t *testing.T) {
	s := compensationPosition(t)
	q, e := QueueCompensation(s)
	if e != nil {
		t.Fatal(e)
	}
	if len(q.DrawQueue) != 2 || q.DrawQueue[0].Seat != 2 {
		t.Fatal("not fixed seat order")
	}
	if _, e = OpenSeries(q, 1, Hearts); e == nil {
		t.Fatal("ordinary action overtook draws")
	}
	n, got, e := StepCompensation(q, nil)
	if e != nil || len(got) != 1 {
		t.Fatal(got, e)
	}
	c, _ := n.Card(cid(Clubs, 12))
	if c.Controller != 2 {
		t.Fatal("wrong first recipient")
	}
	n, got, e = StepCompensation(n, nil)
	if e != nil || len(got) != 0 || len(n.DrawQueue) != 0 {
		t.Fatal("exhaustion fabricated entitlement")
	}
	for _, eff := range n.Effects {
		if eff.Losses != 0 {
			t.Fatal("loss counter not consumed")
		}
	}
	if !reflect.DeepEqual(s.Effects[0].Losses, 5) {
		t.Fatal("queue mutated input")
	}
}
func TestDeferredClosurePreservesCompletedDraw(t *testing.T) {
	s := compensationPosition(t)
	s, _ = QueueCompensation(s)
	s, _, _ = StepCompensation(s, nil)
	n, e := CloseCardBoard(s)
	if e != nil {
		t.Fatal(e)
	}
	c, _ := n.Card(cid(Clubs, 12))
	if c.Controller != 2 || len(n.DrawQueue) != 0 || len(n.Effects) != 0 {
		t.Fatal("closure reset draw or kept pending effect")
	}
	if e = n.Validate(); e != nil {
		t.Fatal(e)
	}
}
func TestDeferredRejectsInvalidProgress(t *testing.T) {
	s := compensationPosition(t)
	s.DrawQueue = []DrawEntitlement{{Seat: 2, Remaining: 0}}
	if _, _, e := StepCompensation(s, nil); e == nil {
		t.Fatal("zero entitlement accepted and became negative")
	}
	s.DrawQueue = []DrawEntitlement{{Seat: 3, Remaining: 1}, {Seat: 2, Remaining: 1}}
	if _, _, e := StepCompensation(s, nil); e == nil {
		t.Fatal("out-of-order queue accepted")
	}
}
func TestDeferredMultipleRecycleCanonicalResume(t *testing.T) {
	s := compensationPosition(t)
	s.Effects[0].Losses = 11
	spade := cid(Spades, 2)
	s, e := s.Move([]string{spade}, 0, Discard, true)
	if e != nil {
		t.Fatal(e)
	}
	s, e = QueueCompensation(s)
	if e != nil {
		t.Fatal(e)
	}
	if s.DrawQueue[0].Remaining != 2 || s.Effects[0].Losses != 1 {
		t.Fatal("multiple quantity/remainder")
	}
	for step := 0; step < 3; step++ {
		raw, e := canonical.Marshal(s)
		if e != nil {
			t.Fatal(e)
		}
		var restored State
		if e = canonical.Decode(raw, &restored); e != nil {
			t.Fatal(e)
		}
		if Digest(restored) != Digest(s) {
			t.Fatal("checkpoint hash changed")
		}
		var recycle []string
		if step == 1 {
			recycle = []string{spade}
		}
		next, got, e := StepCompensation(restored, recycle)
		if e != nil {
			t.Fatal(e)
		}
		direct, directGot, e := StepCompensation(s, recycle)
		if e != nil || Digest(next) != Digest(direct) || !reflect.DeepEqual(got, directGot) {
			t.Fatal("resume repeated/diverged draw")
		}
		if step < 2 && (len(got) != 1 || next.DrawQueue[0].Seat != map[int]int{0: 2, 1: 3}[step]) {
			t.Fatal("recipient quantity ordering")
		}
		if step == 2 && len(got) != 0 {
			t.Fatal("exhaustion fabricated card")
		}
		s = next
	}
	queen, _ := s.Card(cid(Clubs, 12))
	card, _ := s.Card(spade)
	if queen.Controller != 2 || card.Controller != 2 || len(s.DrawQueue) != 0 || s.Effects[0].Losses != 1 || s.Effects[1].Losses != 0 {
		t.Fatal("resumed final state")
	}
}
func TestDeferredLegalCoupAtEachBoundary(t *testing.T) {
	base := compensationPosition(t)
	base.Players[0].History = []Suit{Clubs}
	base.Players[0].Confined = true
	for boundary := 0; boundary <= 2; boundary++ {
		s, e := QueueCompensation(base)
		if e != nil {
			t.Fatal(e)
		}
		for j := 0; j < boundary; j++ {
			s, _, e = StepCompensation(s, nil)
			if e != nil {
				t.Fatal(e)
			}
		}
		n, events, e := Apply(s, Command{GameID: s.GameID, ID: "coup", Kind: "coup", Actor: 1})
		if e != nil || n.Phase != "settlement" || len(n.DrawQueue) != 0 || len(events) != 1 || events[0].Kind != "coup" {
			t.Fatal("legal confined Coup at draw boundary", boundary, e)
		}
		queen, _ := n.Card(cid(Clubs, 12))
		if boundary == 0 && queen.Controller != 0 || boundary > 0 && queen.Controller != 2 {
			t.Fatal("Coup reversed committed draw")
		}
		if len(n.Effects) != 0 {
			t.Fatal("closure retains live effects")
		}
		if e = n.Validate(); e != nil {
			t.Fatal(e)
		}
	}
}

func earnedCompensation(t *testing.T, s State) State {
	t.Helper()
	id := cid(Hearts, 1)
	var e error
	s, e = s.Move([]string{id}, 2, ConcealedAce, true)
	if e != nil {
		t.Fatal(e)
	}
	s, e = s.ActivateEffect(AceEffect{ID: "earned", CardID: id, Kind: "compensation", Source: 2, Target: 2, Custodian: 2, Expiry: "closure", Losses: 5})
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestDeferredIdleAdmissionBeforeQueue(t *testing.T) {
	for _, kind := range []string{"shuffle", "earned"} {
		t.Run(kind, func(t *testing.T) {
			s := combatFixture(t)
			s, e := s.Move([]string{cid(Spades, 2)}, 1, Hand, true)
			if e != nil {
				t.Fatal(e)
			}
			if kind == "shuffle" {
				s.NeedsShuffle = true
			} else {
				s = earnedCompensation(t, s)
			}
			before := Digest(s)
			cmd := Command{GameID: s.GameID, ID: "premature", Kind: "attack", Actor: 1, Cards: []string{cid(Clubs, 8)}, Targets: []string{cid(Hearts, 8)}}
			n, _, e := Apply(s, cmd)
			if e == nil || Digest(n) != before || Digest(s) != before {
				t.Fatal("ordinary attack overtook post-action work")
			}
			n, e = OpenSeries(s, 1, Spades)
			if e == nil || Digest(n) != before {
				t.Fatal("opening overtook post-action work")
			}
			if kind == "shuffle" {
				s, e = ShuffleSupply(s, s.DrawOrder)
			} else {
				s, e = QueueCompensation(s)
				if e == nil {
					s, _, e = StepCompensation(s, nil)
				}
			}
			if e != nil {
				t.Fatal(e)
			}
			if _, _, e = Apply(s, cmd); e != nil {
				t.Fatal("explicit post-action work did not release admission", e)
			}
		})
	}
}
func TestDeferredExistingResponsesStillFinish(t *testing.T) {
	s := combatFixture(t)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "a", Kind: "attack", Actor: 1, Cards: []string{cid(Clubs, 8)}, Targets: []string{cid(Hearts, 8)}})
	if e != nil {
		t.Fatal(e)
	}
	n = earnedCompensation(t, n)
	n.NeedsShuffle = true
	for _, seat := range []int{2, 3} {
		n, _, e = Apply(n, Command{GameID: n.GameID, ID: fmt.Sprint("pass", seat), Kind: "pass", Actor: seat, WindowID: "a"})
		if e != nil {
			t.Fatal("pending response improperly blocked", e)
		}
	}
	if n.Pending != nil || n.Effects[0].Losses != 6 {
		t.Fatal("action response/counter unfinished")
	}
	if _, e = QueueCompensation(n); e == nil {
		t.Fatal("queue bypassed required shuffle")
	}
	n, e = ShuffleSupply(n, n.DrawOrder)
	if e != nil {
		t.Fatal(e)
	}
	n, e = QueueCompensation(n)
	if e != nil || len(n.DrawQueue) != 1 || n.Effects[0].Losses != 1 {
		t.Fatal("required post-action ordering", e)
	}
}
func TestDeferredCoupBeforeQueueAndShuffle(t *testing.T) {
	s := compensationPosition(t)
	s.Players[0].History = []Suit{Clubs}
	s.NeedsShuffle = true
	n, ev, e := Apply(s, Command{GameID: s.GameID, ID: "c", Kind: "coup", Actor: 1})
	if e != nil || len(ev) != 1 || n.Phase != "settlement" {
		t.Fatal("Coup must remain legal before queue/shuffle", e)
	}
}
