package game

import (
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"testing"
)

func combatFixture(t *testing.T) State {
	s, _ := NewState(3, "combat")
	for _, x := range []struct {
		id   string
		seat int
	}{{"deck-1-clubs-08", 1}, {"deck-1-diamonds-02", 1}, {"deck-1-hearts-08", 2}} {
		var e error
		s, e = s.Move([]string{x.id}, x.seat, Series, true)
		if e != nil {
			t.Fatal(e)
		}
	}
	s.Players[1].History = []Suit{Hearts, Clubs, Spades, Diamonds}
	return s
}
func TestCombatExactAndPhysicalUsage(t *testing.T) {
	s := combatFixture(t)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := n.Card("deck-1-clubs-08")
	if c.UsedTurn != 1 || n.Pending == nil {
		t.Fatal("acceptance must commit physical club")
	}
	for _, seat := range []int{2, 3} {
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: string(rune('a' + seat)), Kind: "pass", Actor: seat, WindowID: "attack"})
		if e != nil {
			t.Fatal(e)
		}
	}
	c, _ = n.Card("deck-1-hearts-08")
	if c.Zone != Draw || n.Pending != nil {
		t.Fatal("exact attack must return heart")
	}
	old, _ := s.Card("deck-1-clubs-08")
	if old.UsedTurn != 0 {
		t.Fatal("alias")
	}
}
func TestCombatRejectRetainsState(t *testing.T) {
	s := combatFixture(t)
	before := Digest(s)
	_, _, e := Apply(s, Command{GameID: s.GameID, ID: "bad", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-diamonds-02"}})
	if e == nil || Digest(s) != before {
		t.Fatal("illegal attack committed cost")
	}
}
func TestResponsesSkipConfinedButWaitDisconnected(t *testing.T) {
	s := combatFixture(t)
	s.Players[1].Connected = false
	s.Players[2].Confined = true
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "a", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	if len(n.Pending.Responders) != 1 || n.Pending.Responders[0] != 2 {
		t.Fatal("wrong traversal")
	}
	o, _ := Observe(n, 1)
	if o.RequiredActor != 2 {
		t.Fatal("disconnection fabricated pass")
	}
}
func TestProjectionHiddenEquivalence(t *testing.T) {
	s := combatFixture(t)
	s, _ = s.Move([]string{"deck-1-spades-02"}, 2, Hand, true)
	a, _ := Observe(s, 1)
	n := s.Clone()
	n, _ = n.Move([]string{"deck-1-spades-02"}, 0, Draw, true)
	n, _ = n.Move([]string{"deck-1-spades-03"}, 2, Hand, true)
	n, _ = ShuffleSupply(n, n.DrawOrder)
	b, _ := Observe(n, 1)
	if Digest(a) != Digest(b) {
		t.Fatal("hidden face changes observer")
	}
}
func TestDefenseModifierPhysicalUsageAndNoElimination(t *testing.T) {
	s := combatFixture(t)
	s, _ = s.Move([]string{"deck-1-diamonds-03"}, 2, Series, true)
	s, _ = s.Move([]string{"deck-1-clubs-01"}, 2, ConcealedAce, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "a", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "d", Kind: "advancement-defense", Actor: 2, WindowID: "a", AceID: "deck-1-clubs-01"})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "p", Kind: "pass", Actor: 3, WindowID: "a"})
	if e != nil {
		t.Fatal(e)
	}
	h, _ := n.Card("deck-1-hearts-08")
	club, _ := n.Card("deck-1-clubs-08")
	ace, _ := n.Card("deck-1-clubs-01")
	if h.Zone != Series || club.UsedTurn != 1 || ace.Zone != Discard {
		t.Fatal("virtual modifier lifecycle")
	}
}
func TestNegotiatorStagesAndRetry(t *testing.T) {
	s := combatFixture(t)
	s, _ = s.Move([]string{"deck-1-spades-06", "deck-1-spades-09"}, 2, Series, true)
	s, _ = s.Move([]string{"deck-1-spades-11"}, 2, Attachment, true)
	s, _ = s.Move([]string{"deck-1-clubs-07"}, 1, Series, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "a", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "n", Kind: "negotiate", Actor: 2, WindowID: "a"})
	if e != nil {
		t.Fatal(e)
	}
	n = roundtripCombat(t, n)
	if n.Pending.Decision.Stage != "clubs" {
		t.Fatal("missing attacker stage")
	}
	cmd := Command{GameID: s.GameID, ID: "d1", Kind: "decision", Actor: 1, WindowID: "a", DecisionID: "n/clubs", Cards: []string{"deck-1-clubs-08", "deck-1-clubs-07"}}
	n, _, e = Apply(n, cmd)
	if e != nil {
		t.Fatal(e)
	}
	same, _, e := Apply(n, cmd)
	if e != nil || Digest(same) != Digest(n) {
		t.Fatal("retry changed state")
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "d2", Kind: "decision", Actor: 2, WindowID: "a", DecisionID: "n/payment", Cards: []string{"deck-1-spades-06", "deck-1-spades-09"}})
	if e != nil {
		t.Fatal(e)
	}
	if n.Pending != nil {
		t.Fatal("did not cancel")
	}
	for _, id := range []string{"deck-1-clubs-08", "deck-1-clubs-07"} {
		c, _ := n.Card(id)
		if c.UsedTurn != 1 {
			t.Fatal("added/original club not used")
		}
	}
}
func TestDexterConfinementContinuation(t *testing.T) {
	s := combatFixture(t)
	s, _ = s.Move([]string{"deck-1-diamonds-11"}, 1, Attachment, true)
	s, _ = s.Move([]string{"deck-1-spades-02"}, 2, Series, true)
	s, _ = s.Move([]string{"deck-1-diamonds-01"}, 2, ConcealedAce, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "a", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "c", Kind: "confinement", Actor: 2, WindowID: "a", AceID: "deck-1-diamonds-01"})
	if e != nil {
		t.Fatal(e)
	}
	n = roundtripCombat(t, n)
	if n.Pending.Decision == nil || n.Pending.Decision.Actor != 1 {
		t.Fatal("missing Dexter decision")
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "prevent", Kind: "decision", Actor: 1, WindowID: "a", DecisionID: "c/dexter", Cards: []string{"deck-1-diamonds-02"}})
	if e != nil {
		t.Fatal(e)
	}
	if n.Players[0].Confined || n.Pending == nil || n.Pending.Responders[n.Pending.Cursor] != 3 {
		t.Fatal("nested/canceled response")
	}
	c, _ := n.Card("deck-1-diamonds-02")
	a, _ := n.Card("deck-1-diamonds-01")
	if c.Zone != Draw || a.Zone != Discard {
		t.Fatal("atomic prevention payment")
	}
}
func TestInflationSpendOrdering(t *testing.T) {
	s := combatFixture(t)
	s, _ = s.Move([]string{"deck-1-spades-10", "deck-2-spades-10"}, 1, Series, true)
	s, _ = s.Move([]string{"deck-1-spades-01"}, 2, ConcealedAce, true)
	s, e := s.ActivateEffect(AceEffect{ID: "i", CardID: "deck-1-spades-01", Kind: "inflation", Source: 2, Target: 1, Custodian: 1, Expiry: "qualifying-transaction-or-target-departure-or-closure"})
	if e != nil {
		t.Fatal(e)
	}
	for _, x := range []struct {
		voluntary, complete bool
		remaining           int
	}{{false, true, 1}, {true, false, 1}, {true, true, 0}} {
		n, v, e := CompleteSpadeSpend(s, 1, []string{"deck-1-spades-10", "deck-2-spades-10"}, x.voluntary, x.complete)
		if e != nil || v.Cmp(IntAmount(10)) != 0 || len(n.Effects) != x.remaining {
			t.Fatal("active-price then clear", e)
		}
	}
}
func TestJusticeContinuationAndCoupBoundary(t *testing.T) {
	s := combatFixture(t)
	queens := []string{"deck-1-hearts-12", "deck-1-clubs-12", "deck-1-spades-12", "deck-1-diamonds-12"}
	s, _ = s.Move(queens, 1, Formation, true)
	for i := range s.Cards {
		for _, id := range queens {
			if s.Cards[i].Card.ID == id {
				s.Cards[i].Allocation = "justice"
			}
		}
	}
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "j", Kind: "justice", Actor: 1, AceID: queens[0]})
	if e != nil {
		t.Fatal(e)
	}
	for _, seat := range []int{2, 3} {
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: string(rune('k' + seat)), Kind: "pass", Actor: seat, WindowID: "j"})
		if e != nil {
			t.Fatal(e)
		}
	}
	if n.Pending == nil || n.Pending.Decision.ID != "j/select/0" {
		t.Fatal("missing Justice stage")
	}
	bad := Command{GameID: s.GameID, ID: "bad", Kind: "decision", Actor: 2, WindowID: "j", DecisionID: "j/select/0", Cards: []string{"deck-1-hearts-08"}}
	old := Digest(n)
	if _, _, e = Apply(n, bad); e == nil || Digest(n) != old {
		t.Fatal("wrong actor accepted")
	}
	cmd := bad
	cmd.Actor = 1
	cmd.ID = "choose"
	n, _, e = Apply(n, cmd)
	if e != nil {
		t.Fatal(e)
	}
	card, _ := n.Card("deck-1-hearts-08")
	if card.Controller != 1 || card.Zone != Hand {
		t.Fatal("Justice transfer")
	}
	same, _, e := Apply(n, cmd)
	if e != nil || Digest(same) != Digest(n) {
		t.Fatal("Justice retry")
	}
}
func TestVirtualAttackPhysicalExchange(t *testing.T) {
	s, _ := NewState(3, "virtual")
	for _, x := range []struct {
		ids  []string
		seat int
		zone Zone
	}{{[]string{"deck-1-clubs-05", "deck-1-diamonds-02"}, 1, Series}, {[]string{"deck-1-clubs-01"}, 1, ConcealedAce}, {[]string{"deck-1-clubs-07", "deck-1-clubs-08"}, 2, Series}} {
		s, _ = s.Move(x.ids, x.seat, x.zone, true)
	}
	s.Players[1].History = []Suit{Hearts, Clubs, Spades, Diamonds}
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "a", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-05"}, Targets: []string{"deck-1-clubs-07", "deck-1-clubs-08"}, AceID: "deck-1-clubs-01"})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "p2", Kind: "pass", Actor: 2, WindowID: "a"})
	if e != nil {
		t.Fatal(e)
	}
	n, events, e := Apply(n, Command{GameID: s.GameID, ID: "p3", Kind: "pass", Actor: 3, WindowID: "a"})
	if e != nil {
		t.Fatal(e)
	}
	if len(events) != 2 || events[0].Amount.Cmp(IntAmount(3)) != 0 {
		t.Fatal("physical scoring", events)
	}
	for _, id := range []string{"deck-1-clubs-05", "deck-1-clubs-07", "deck-1-clubs-08"} {
		c, _ := n.Card(id)
		if c.Zone != Draw {
			t.Fatal("physical return")
		}
	}
	ace, _ := n.Card("deck-1-clubs-01")
	if ace.Zone != Discard {
		t.Fatal("ace cost")
	}
}
func TestCoupInterruptsDecisionAndKeepsCosts(t *testing.T) {
	s := combatFixture(t)
	s.Players[0].History = []Suit{Clubs}
	s, _ = s.Move([]string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-diamonds-12", "deck-2-diamonds-12"}, 1, Hand, true)
	s, _ = s.Move([]string{"deck-1-spades-06", "deck-1-spades-09"}, 2, Series, true)
	s, _ = s.Move([]string{"deck-1-spades-11"}, 2, Attachment, true)
	s, _ = s.Move([]string{"deck-1-clubs-07"}, 1, Series, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "a", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "n", Kind: "negotiate", Actor: 2, WindowID: "a"})
	if e != nil {
		t.Fatal(e)
	}
	n.Players[0].Confined = true
	n, ev, e := Apply(n, Command{GameID: s.GameID, ID: "c", Kind: "coup", Actor: 1})
	if e != nil || len(ev) != 1 || n.Pending != nil || n.Phase != "settlement" {
		t.Fatal("Coup boundary", e)
	}
	club, _ := n.Card("deck-1-clubs-08")
	added, _ := n.Card("deck-1-clubs-07")
	jack, _ := n.Card("deck-1-spades-11")
	if club.UsedTurn != 1 || added.UsedTurn != 0 || jack.Zone != Attachment || n.Players[1].Quotas["negotiator"] != 1 {
		t.Fatal("aborted decision costs")
	}
}
func TestBarricadeVirtualExchange(t *testing.T) {
	s := combatFixture(t)
	s, _ = s.Move([]string{"deck-1-hearts-12"}, 2, Attachment, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "a", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	for _, seat := range []int{2, 3} {
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: string(rune('a' + seat)), Kind: "pass", Actor: seat, WindowID: "a"})
		if e != nil {
			t.Fatal(e)
		}
	}
	c, _ := n.Card("deck-1-clubs-08")
	if c.Zone != Draw {
		t.Fatal("Barricade physical exchange omitted")
	}
}
func FuzzRejectedCommandPreservesState(f *testing.F) {
	f.Add("unsupported", 1)
	f.Add("attack", 2)
	f.Fuzz(func(t *testing.T, kind string, actor int) {
		s := combatFixture(t)
		before := Digest(s)
		n, _, e := Apply(s, Command{GameID: s.GameID, ID: "fuzz", Kind: kind, Actor: actor})
		if e != nil && Digest(n) != before {
			t.Fatal("rejection changed state")
		}
	})
}
func TestInfiltratorDeclarationCommitsReturn(t *testing.T) {
	s := combatFixture(t)
	s, _ = s.Move([]string{"deck-1-clubs-11"}, 1, Attachment, true)
	s, _ = s.Move([]string{"deck-1-spades-08"}, 2, Series, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "i", Kind: "infiltrator-attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-spades-08"}})
	if e != nil {
		t.Fatal(e)
	}
	jack, _ := n.Card("deck-1-clubs-11")
	if jack.Zone != Draw || n.Players[0].Quotas["infiltrator"] != turnCycle(s, 1) {
		t.Fatal("accepted bypass must spend jack and quota")
	}
}
func TestOutOfTurnPonziConfinement(t *testing.T) {
	s := combatFixture(t)
	s.Active = 3
	jacks := []string{"deck-1-clubs-11", "deck-2-clubs-11", "deck-1-spades-11", "deck-2-spades-11"}
	s, _ = s.Move(jacks, 1, Formation, true)
	for i := range s.Cards {
		for _, id := range jacks {
			if s.Cards[i].Card.ID == id {
				s.Cards[i].Allocation = "ponzi"
			}
		}
	}
	s, _ = s.Move([]string{"deck-1-spades-03"}, 2, Series, true)
	s, _ = s.Move([]string{"deck-1-diamonds-01"}, 2, ConcealedAce, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "ponzi", Kind: "ponzi", Actor: 1, Value: 2})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "confine", Kind: "confinement", Actor: 2, WindowID: "ponzi", AceID: "deck-1-diamonds-01"})
	if e != nil {
		t.Fatal(e)
	}
	if !n.Players[0].Confined || n.Active != 3 || n.Pending != nil || n.Players[0].Quotas["ponzi"] != turnCycle(s, 1) {
		t.Fatal("out-of-turn confinement changed active turn")
	}
	for _, id := range jacks {
		c, _ := n.Card(id)
		if c.Zone != Formation {
			t.Fatal("canceled Ponzi resolution cost")
		}
	}
}
func TestBombOrdinaryBaronCancellationCosts(t *testing.T) {
	for _, baron := range []bool{false, true} {
		for _, resolved := range []bool{false, true} {
			s := combatFixture(t)
			s, _ = s.Move([]string{"deck-1-hearts-11"}, 2, Attachment, true)
			for i := range s.Cards {
				if s.Cards[i].Card.ID == "deck-1-hearts-11" {
					s.Cards[i].Allocation = "hearts"
				}
			}
			if baron {
				s, _ = s.Move([]string{"deck-1-hearts-13", "deck-2-hearts-13", "deck-1-diamonds-13"}, 1, Formation, true)
				for i := range s.Cards {
					if s.Cards[i].Controller == 1 && s.Cards[i].Zone == Formation {
						s.Cards[i].Allocation = "baron"
					}
				}
			}
			n, e := BombBoundary(s, 1, []string{"deck-1-clubs-08"}, "deck-1-hearts-11", baron, resolved)
			if e != nil {
				t.Fatal(e)
			}
			club, _ := n.Card("deck-1-clubs-08")
			bomb, _ := n.Card("deck-1-hearts-11")
			if club.UsedTurn != s.Turn {
				t.Fatal("declaration usage")
			}
			want := Series
			if resolved && !baron {
				want = Draw
			}
			if club.Zone != want {
				t.Fatal("ordinary/Baron return cost")
			}
			if resolved && (bomb.Zone != Discard || n.Players[1].Quotas["bomb-immunity/hearts"] != s.Round) {
				t.Fatal("successful bomb shield")
			}
			if !resolved && (bomb.Zone != Attachment || n.Players[1].Quotas["bomb-immunity/hearts"] != 0) {
				t.Fatal("cancellation resolution cost")
			}
		}
	}
}
func TestBaronCombatAlgebra(t *testing.T) {
	if CombatMatches(IntAmount(10), IntAmount(10), true) || !CombatMatches(IntAmount(11), IntAmount(10), true) || !CombatMatches(IntAmount(10), IntAmount(10), false) {
		t.Fatal("Baron strict inequality")
	}
}

func roundtripCombat(t *testing.T, s State) State {
	t.Helper()
	b, e := canonical.Marshal(s)
	if e != nil {
		t.Fatal(e)
	}
	var n State
	if e = canonical.Decode(b, &n); e != nil {
		t.Fatal(e)
	}
	if Digest(n) != Digest(s) {
		t.Fatal("checkpoint digest changed")
	}
	return n
}
func TestJusticeRandomOutcomeRestartRetry(t *testing.T) {
	s := combatFixture(t)
	s.Players[2].History = []Suit{Hearts, Clubs, Spades, Diamonds}
	queens := []string{"deck-1-hearts-12", "deck-1-clubs-12", "deck-1-spades-12", "deck-1-diamonds-12"}
	s, _ = s.Move(queens, 1, Formation, true)
	for i := range s.Cards {
		if s.Cards[i].Controller == 1 && s.Cards[i].Zone == Formation {
			s.Cards[i].Allocation = "justice"
		}
	}
	s, _ = s.Move([]string{"deck-1-spades-02"}, 2, Hand, true)
	s, _ = s.Move([]string{"deck-1-spades-03"}, 3, Hand, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "j", Kind: "justice", Actor: 1, AceID: queens[0]})
	if e != nil {
		t.Fatal(e)
	}
	for _, seat := range []int{2, 3} {
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: string(rune('k' + seat)), Kind: "pass", Actor: seat, WindowID: "j"})
		if e != nil {
			t.Fatal(e)
		}
	}
	n = roundtripCombat(t, n)
	before := Digest(n)
	if _, _, e = Apply(n, Command{GameID: s.GameID, ID: "pick", Kind: "decision", Actor: 1, WindowID: "j", DecisionID: "j/select/0", Cards: []string{"deck-1-spades-02"}}); e == nil || Digest(n) != before {
		t.Fatal("hidden identity selection accepted")
	}
	cmd := Command{GameID: s.GameID, ID: "random", Kind: "decision", Actor: 1, WindowID: "j", DecisionID: "j/select/0", RandomWords: []string{"0"}}
	n, _, e = Apply(n, cmd)
	if e != nil {
		t.Fatal(e)
	}
	n = roundtripCombat(t, n)
	same, _, e := Apply(n, cmd)
	if e != nil || Digest(same) != Digest(n) {
		t.Fatal("retry resampled")
	}
	own, _ := Observe(n, 1)
	other, _ := Observe(n, 3)
	if len(own.KnownReserved) != 1 || own.KnownReserved[0].ID != "deck-1-spades-02" || len(other.KnownReserved) != 0 {
		t.Fatal("reserved identity visibility")
	}
	stale := cmd
	stale.ID = "stale"
	if _, _, e = Apply(n, stale); e == nil {
		t.Fatal("stale decision accepted")
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "next", Kind: "decision", Actor: 1, WindowID: "j", DecisionID: "j/select/1", RandomWords: []string{"0"}})
	if e != nil || n.Pending != nil {
		t.Fatal("second sample resolution", e)
	}
}
func TestConfinementWaitsForScheduledTurnDraw(t *testing.T) {
	s := combatFixture(t)
	s.Order = []int{1, 3, 2}
	s, _ = s.Move([]string{"deck-1-spades-03"}, 2, Series, true)
	s, _ = s.Move([]string{"deck-1-diamonds-01"}, 2, ConcealedAce, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "a", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "c", Kind: "confinement", Actor: 2, WindowID: "a", AceID: "deck-1-diamonds-01"})
	if e != nil {
		t.Fatal(e)
	}
	if n.Active != 3 || n.Pending == nil || n.Pending.Kind != "scheduled-turn-draw" {
		t.Fatal("missing scheduled draw boundary")
	}
	o, _ := Observe(n, 3)
	if o.RequiredActor != 3 || o.DecisionKind != "scheduled-turn-draw" || len(o.Legal) != 0 {
		t.Fatal("fabricated next-turn action")
	}
}

func TestPonziQuotaUsesOwnTurnCycle(t *testing.T) {
	s := combatFixture(t)
	s.Order = []int{1, 2, 3}
	s.Active = 2
	a := turnCycle(s, 1)
	s.Active = 3
	if turnCycle(s, 1) != a {
		t.Fatal("opponent turn reset own quota")
	}
	s.Round++
	s.Active = 1
	if turnCycle(s, 1) == a {
		t.Fatal("own turn failed reset")
	}
}
func TestSecondHistoricalOpeningEndsProtectionPermanently(t *testing.T) {
	s := combatFixture(t)
	s.Players[1].History = []Suit{Diamonds, Hearts}
	if len(s.CurrentSeries(2)) != 1 {
		t.Fatal("fixture must have lost a current series")
	}
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "two-history", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil || n.Pending == nil {
		t.Fatal("second opening must permanently end protection", e)
	}
	s.Players[1].History = []Suit{Diamonds}
	if _, _, e = Apply(s, Command{GameID: s.GameID, ID: "still-protected", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}}); e == nil {
		t.Fatal("first series protection lost")
	}
}

func TestFullBaronAttackRequiresWholeSeriesAndScoresNoClubs(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-1-clubs-10"}, 1, Series, false)
	s, _ = s.Move([]string{"deck-2-clubs-03", "deck-2-clubs-04", "deck-2-diamonds-02"}, 2, Series, false)
	s.Players[1].History = []Suit{Clubs, Diamonds}
	ids := []string{"deck-1-hearts-13", "deck-2-hearts-13", "deck-1-diamonds-13"}
	s, _ = s.Move(ids, 1, Hand, false)
	s, e := OpenFormation(s, 1, "b", FormationSpec{Kind: "baron", Cards: ids})
	if e != nil {
		t.Fatal(e)
	}
	c := Command{ID: "baron", Kind: "baron-attack", Actor: 1, Cards: []string{"deck-1-clubs-10"}, Targets: []string{"deck-2-clubs-03", "deck-2-clubs-04"}}
	n, e := attack(s, c)
	if e != nil {
		t.Fatal(e)
	}
	n.Pending.Cursor = len(n.Pending.Responders)
	n, events, e := finishCombat(n)
	if e != nil {
		t.Fatal(e)
	}
	for _, ev := range events {
		if ev.Kind == "combat-score" {
			t.Fatal("Baron incorrectly scored Clubs")
		}
	}
	card, _ := n.Card("deck-1-clubs-10")
	if card.Zone != Draw {
		t.Fatal("physical Club exchange missing")
	}
	c.Targets = c.Targets[:1]
	if _, e = attack(s, c); e == nil {
		t.Fatal("Baron partial target accepted")
	}
}

func TestFullBombAttackPhysicalCostAndShield(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-1-clubs-05"}, 1, Series, false)
	s, _ = s.Move([]string{"deck-2-hearts-05", "deck-2-diamonds-02"}, 2, Series, false)
	s.Players[1].History = []Suit{Hearts, Diamonds}
	s, _ = s.Move([]string{"deck-2-hearts-11"}, 2, Attachment, false)
	for i := range s.Cards {
		if s.Cards[i].Card.ID == "deck-2-hearts-11" {
			s.Cards[i].Allocation = string(Hearts)
		}
	}
	if _, e := attack(s, Command{ID: "illegal", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-05"}, Targets: []string{"deck-2-hearts-05"}}); e == nil {
		t.Fatal("attacked series before Bomb")
	}
	n, e := attack(s, Command{ID: "bomb", Kind: "bomb-attack", Actor: 1, Cards: []string{"deck-1-clubs-05"}, Targets: []string{"deck-2-hearts-11"}})
	if e != nil {
		t.Fatal(e)
	}
	n.Pending.Cursor = len(n.Pending.Responders)
	n, _, e = finishCombat(n)
	if e != nil {
		t.Fatal(e)
	}
	c, _ := n.Card("deck-1-clubs-05")
	b, _ := n.Card("deck-2-hearts-11")
	if c.Zone != Draw || b.Zone != Discard || n.Players[1].Quotas["bomb-immunity/hearts"] != 1 {
		t.Fatal("Bomb physical cost/shield missing")
	}
}
func TestFullExileCommitsQuotaAndAvoidsBarricadeExchange(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-1-clubs-05"}, 1, Series, false)
	s, _ = s.Move([]string{"deck-2-hearts-05", "deck-2-diamonds-02"}, 2, Series, false)
	s.Players[1].History = []Suit{Hearts, Diamonds}
	s, _ = s.Move([]string{"deck-2-hearts-12"}, 2, Attachment, false)
	for i := range s.Cards {
		if s.Cards[i].Card.ID == "deck-2-hearts-12" {
			s.Cards[i].Allocation = string(Hearts)
		}
	}
	s, _ = s.Move([]string{"deck-1-clubs-12"}, 1, Hand, false)
	s, _ = AttachRoyal(s, 1, "deck-1-clubs-12", Clubs)
	n, e := attack(s, Command{ID: "exile", Kind: "attack", Actor: 1, Exile: true, Cards: []string{"deck-1-clubs-05"}, Targets: []string{"deck-2-hearts-05"}})
	if e != nil {
		t.Fatal(e)
	}
	if n.Players[0].Quotas["exile"] != 1 {
		t.Fatal("Exile acceptance quota missing")
	}
	n.Pending.Cursor = len(n.Pending.Responders)
	n, _, e = finishCombat(n)
	if e != nil {
		t.Fatal(e)
	}
	c, _ := n.Card("deck-1-clubs-05")
	if c.Zone != Series || c.UsedTurn != 1 {
		t.Fatal("Exile club returned or reusable")
	}
}
func TestFullPeopleProtectsOnlyNamedSeriesAndBarsOwnClubs(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-1-clubs-05"}, 1, Series, false)
	s, _ = s.Move([]string{"deck-2-hearts-05", "deck-2-diamonds-02", "deck-2-clubs-05"}, 2, Series, false)
	s.Players[1].History = []Suit{Hearts, Diamonds, Clubs}
	s.Active = 2
	ids := []string{"deck-1-hearts-12", "deck-2-hearts-12"}
	s, _ = s.Move(ids, 2, Hand, false)
	s, e := OpenFormation(s, 2, "gp", FormationSpec{Kind: "great-people", Cards: ids, Protection: &FormationProtection{Series: Clubs}})
	if e != nil {
		t.Fatal(e)
	}
	s.Active = 1
	if _, e = attack(s, Command{ID: "heart", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-05"}, Targets: []string{"deck-2-hearts-05"}}); e != nil {
		t.Fatal("People incorrectly protected Hearts", e)
	}
	s.Active = 2
	s.Players[0].History = []Suit{Hearts, Diamonds, Clubs}
	if _, e = attack(s, Command{ID: "blocked", Kind: "attack", Actor: 2, Cards: []string{"deck-2-clubs-05"}, Targets: []string{"deck-1-hearts-02"}}); e == nil {
		t.Fatal("protected Clubs attacked")
	}
}
