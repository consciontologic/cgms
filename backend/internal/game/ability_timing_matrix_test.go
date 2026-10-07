package game

import "testing"

// Every fixture first establishes a legal activation with complete support.
func abilityTimingFixture(t *testing.T, kind string) (State, Command) {
	t.Helper()
	s := formationFixture(t)
	move := func(ids []string, seat int, zone Zone) {
		t.Helper()
		var e error
		s, e = s.Move(ids, seat, zone, false)
		if e != nil {
			t.Fatal(e)
		}
	}
	move([]string{"deck-2-hearts-02", "deck-2-clubs-02"}, 2, Series)
	s.Players[1].History = []Suit{Hearts, Clubs}
	move([]string{"deck-1-diamonds-01"}, 2, ConcealedAce)
	c := Command{ID: "ability", GameID: s.GameID, Actor: 1, Kind: kind}
	switch kind {
	case "ringleader", "richer-sacrifice", "barricade-sacrifice":
		id := "deck-1-diamonds-13"
		if kind == "richer-sacrifice" {
			id = "deck-1-hearts-13"
		}
		if kind == "barricade-sacrifice" {
			id = "deck-1-hearts-12"
		}
		move([]string{id}, 1, Hand)
		var e error
		s, e = AttachRoyal(s, 1, id, Hearts)
		if e != nil {
			t.Fatal(e)
		}
		c.Cards = []string{id}
		if kind == "barricade-sacrifice" {
			extra := []string{"deck-1-hearts-03", "deck-1-hearts-04", "deck-1-hearts-05", "deck-1-hearts-06", "deck-1-hearts-07"}
			move(extra, 1, Hand)
			c.Cards = append(c.Cards, extra...)
		}
	case "compensation", "main-inflation":
		c.AceID = "deck-1-hearts-01"
		if kind == "main-inflation" {
			c.AceID = "deck-1-spades-01"
			c.TargetSeat = 2
		}
		move([]string{c.AceID}, 1, ConcealedAce)
	case "fate", "code", "kidnapper":
		ids := []string{"deck-1-hearts-13", "deck-1-clubs-13", "deck-1-spades-13", "deck-1-diamonds-13"}
		if kind == "code" {
			ids = []string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-clubs-12", "deck-2-clubs-12", "deck-1-spades-12"}
			c.Value = 7
			c.Price = 8
		}
		if kind == "kidnapper" {
			ids = []string{"deck-1-clubs-11", "deck-2-clubs-11"}
		}
		move(ids, 1, Hand)
		var e error
		s, e = OpenFormation(s, 1, "combo", FormationSpec{Kind: kind, Cards: ids})
		if e != nil {
			t.Fatal(e)
		}
		c.FormationID = "combo"
		if kind != "code" {
			c.TargetSeat = 2
		}
		if kind == "fate" {
			c.Cards = ids[:1]
		}
	case "dexter-assassination":
		move([]string{"deck-1-diamonds-10", "deck-2-diamonds-10"}, 1, Series)
		move([]string{"deck-1-diamonds-11"}, 1, Hand)
		var e error
		s, e = AttachRoyal(s, 1, "deck-1-diamonds-11", Diamonds)
		if e != nil {
			t.Fatal(e)
		}
		move([]string{"deck-2-spades-13"}, 2, Unassigned)
		c.Cards = []string{"deck-1-diamonds-02"}
		c.Targets = []string{"deck-2-spades-13"}
	}
	return s, c
}

func TestAbilityTimingSupportAndQuotaMatrix(t *testing.T) {
	for _, kind := range []string{"ringleader", "richer-sacrifice", "barricade-sacrifice", "compensation", "main-inflation", "fate", "code", "kidnapper", "dexter-assassination"} {
		t.Run(kind, func(t *testing.T) {
			s, c := abilityTimingFixture(t, kind)
			original := Digest(s)
			accepted, _, e := Apply(s, c)
			if e != nil {
				t.Fatal("legal supported fixture", e)
			}
			if Digest(s) != original {
				t.Fatal("legal admission mutated predecessor")
			}
			checks := map[string]func(State) State{
				"other-turn":   func(n State) State { n.Active = 2; return n },
				"confined":     func(n State) State { n.Players[0].Confined = true; return n },
				"closed-board": func(n State) State { n.Phase = "settlement"; return n },
				"pending-action": func(n State) State {
					n.Pending = &PendingAction{ID: "other", Kind: "scheduled-turn-draw", Actor: 2}
					return n
				},
				"missing-number-support": func(n State) State {
					ids := []string{}
					for _, v := range n.Cards {
						if v.Controller == 1 && v.Zone == Series {
							ids = append(ids, v.Card.ID)
						}
					}
					var err error
					n, err = n.Move(ids, 1, Hand, false)
					if err != nil {
						t.Fatal(err)
					}
					return n
				},
				"quota-consumed": func(n State) State { n.Players[0] = accepted.Clone().Players[0]; return n },
			}
			for name, mutate := range checks {
				t.Run(name, func(t *testing.T) {
					n := mutate(s.Clone())
					before := Digest(n)
					rejected, events, err := Apply(n, c)
					if err == nil || len(events) != 0 || Digest(rejected) != before || Digest(n) != before {
						t.Fatal("illegal activation accepted or spent/revealed/mutated")
					}
				})
			}
		})
	}
}

func TestResponseTimingSupportAndQuotaMatrix(t *testing.T) {
	for _, kind := range []string{"confinement", "inflation", "compensation-response", "numerical-defense", "advancement-defense", "negotiate"} {
		t.Run(kind, func(t *testing.T) {
			s := combatFixture(t)
			var e error
			s, e = s.Move([]string{"deck-1-spades-08"}, 2, Series, true)
			if e != nil {
				t.Fatal(e)
			}
			c := Command{GameID: s.GameID, ID: "response", WindowID: "attack", Actor: 2, Kind: kind, AceID: "deck-1-hearts-01", Value: 2}
			if kind == "confinement" {
				c.AceID = "deck-1-diamonds-01"
			}
			if kind == "inflation" {
				c.AceID = "deck-1-spades-01"
			}
			if kind == "advancement-defense" {
				c.AceID = "deck-1-clubs-01"
			}
			if kind == "negotiate" {
				c.AceID = ""
				s, e = s.Move([]string{"deck-1-spades-11"}, 2, Attachment, true)
			} else {
				s, e = s.Move([]string{c.AceID}, 2, ConcealedAce, false)
			}
			if e != nil {
				t.Fatal(e)
			}
			s, _, e = Apply(s, Command{ID: "attack", GameID: s.GameID, Actor: 1, Kind: "attack", Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
			if e != nil {
				t.Fatal(e)
			}
			if _, _, e = Apply(s, c); e != nil {
				t.Fatal("legal response", e)
			}
			for _, name := range []string{"idle", "wrong-responder", "confined", "closed-board", "already-used", "missing-support"} {
				t.Run(name, func(t *testing.T) {
					n := s.Clone()
					switch name {
					case "idle":
						n.Pending = nil
					case "wrong-responder":
						n.Pending.Cursor = 1
					case "confined":
						n.Players[1].Confined = true
					case "closed-board":
						n.Phase = "settlement"
					case "already-used":
						if kind == "negotiate" {
							n.Players[1].Quotas["negotiator"] = n.Round
						} else {
							n.Players[1].AceRound = n.Round
						}
					case "missing-support":
						ids := []string{}
						for _, v := range n.Cards {
							if v.Controller == 2 && v.Zone == Series {
								ids = append(ids, v.Card.ID)
							}
						}
						n, e = n.Move(ids, 2, Hand, false)
						if e != nil {
							t.Fatal(e)
						}
					}
					before := Digest(n)
					out, events, err := Apply(n, c)
					if err == nil || len(events) != 0 || Digest(out) != before {
						t.Fatal("illegal response accepted or spent/moved cards")
					}
				})
			}
		})
	}
}

func TestRestrictedHoldingsCountForOrdinaryProofButBlockCoupUntilRoundBoundary(t *testing.T) {
	l := lifecycleFixture(t)
	ids := []string{}
	for _, c := range Deck() {
		if c.Suit == Diamonds && c.Rank >= 2 && c.Rank <= 10 && len(ids) < 13 {
			ids = append(ids, c.ID)
		}
	}
	var e error
	l.Board, e = l.Board.Move(ids, 1, Hand, false)
	if e != nil {
		t.Fatal(e)
	}
	l.Board.Players[0].History = []Suit{Hearts, Clubs, Spades, Diamonds}
	for i := range l.Board.Cards {
		if l.Board.Cards[i].Controller == 1 {
			l.Board.Cards[i].AvailableFromRound = 2
		}
	}
	n, e := l.DeclareOrdinary(1, "diamonds")
	if e != nil {
		t.Fatal("restricted holdings still count toward ordinary proof", e)
	}
	if n.Ending != "diamonds" || len(n.Board.PublicHistory) < 13 {
		t.Fatal("sufficient restricted proof missing")
	}
	l = lifecycleFixture(t)
	queens := []string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-diamonds-12", "deck-2-diamonds-12"}
	l.Board, e = l.Board.Move(queens, 1, Hand, false)
	if e != nil {
		t.Fatal(e)
	}
	l.Board.Players[0].History = []Suit{Clubs}
	for i := range l.Board.Cards {
		if l.Board.Cards[i].Card.ID == queens[0] {
			l.Board.Cards[i].AvailableFromRound = 2
		}
	}
	before := Digest(l)
	if rejected, err := l.DeclareCoup(Command{GameID: l.Board.GameID, ID: "restricted-coup", Kind: "coup", Actor: 1}); err == nil || Digest(rejected) != before {
		t.Fatal("restricted queen supplied Coup")
	}
	for seat, id := range []string{"deck-1-diamonds-02", "deck-1-diamonds-03", "deck-1-diamonds-04"} {
		l.Board, e = l.Board.Move([]string{id}, seat+1, Series, true)
		if e != nil {
			t.Fatal(e)
		}
	}
	l.RoundClosed = true
	l.TurnIndex = 2
	l.Board.Active = 3
	l, e = l.BeginRound(nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	card, _ := l.Board.Card(queens[0])
	if card.AvailableFromRound != 0 {
		t.Fatal("round boundary failed marker expiry before initiative")
	}
	l, e = l.DeclareCoup(Command{GameID: l.Board.GameID, ID: "released-coup", Kind: "coup", Actor: 1})
	if e != nil || l.Ending != "coup" {
		t.Fatal("expired queen unavailable for Coup", e)
	}
}

func TestOrdinaryClosureDiscardsBothPersistentAceKinds(t *testing.T) {
	l := lifecycleFixture(t)
	var e error
	for _, v := range []struct {
		id, kind       string
		source, target int
	}{{"deck-1-hearts-01", "compensation", 1, 1}, {"deck-1-spades-01", "inflation", 2, 1}} {
		l.Board, e = l.Board.Move([]string{v.id}, v.source, ConcealedAce, false)
		if e != nil {
			t.Fatal(e)
		}
		expiry := "game-end"
		if v.kind == "inflation" {
			expiry = "qualifying-spend"
		}
		l.Board, e = l.Board.ActivateEffect(AceEffect{ID: v.kind, CardID: v.id, Kind: v.kind, Source: v.source, Target: v.target, Custodian: v.target, Expiry: expiry})
		if e != nil {
			t.Fatal(e)
		}
	}
	l, e = l.closeOrdinary("round-limit", 0)
	if e != nil {
		t.Fatal(e)
	}
	if len(l.Board.Effects) != 0 {
		t.Fatal("closed board retains persistent live effects")
	}
	for _, id := range []string{"deck-1-hearts-01", "deck-1-spades-01"} {
		c, _ := l.Board.Card(id)
		if c.Zone != Discard || c.Controller != 0 {
			t.Fatal("closed persistent Ace custody", id)
		}
	}
	if e = l.Board.Validate(); e != nil {
		t.Fatal("closure conservation", e)
	}
}

func TestPurchasePostResponseSupplyLossNeverResizesChosenQuantity(t *testing.T) {
	s, _ := NewState(3, "supply-revalidation")
	var e error
	for _, v := range s.Cards {
		if v.Card.ID == "deck-1-hearts-02" {
			continue
		}
		zone := Hand
		if v.Card.Rank == 1 {
			zone = ConcealedAce
		}
		s, e = s.Move([]string{v.Card.ID}, 2, zone, true)
		if e != nil {
			t.Fatal(e)
		}
	}
	payment := "deck-1-spades-10"
	s, e = s.Move([]string{payment}, 1, Hand, false)
	if e != nil {
		t.Fatal(e)
	}
	s, _, e = Apply(s, Command{ID: "buy", GameID: s.GameID, Actor: 1, Kind: "purchase", Value: 2, Cards: []string{payment}})
	if e != nil {
		t.Fatal(e)
	}
	for _, seat := range []int{2, 3} {
		s, _, e = Apply(s, Command{ID: string(rune('a' + seat)), GameID: s.GameID, Actor: seat, Kind: "pass", WindowID: "buy"})
		if e != nil {
			t.Fatal(e)
		}
	}
	// Boundary fixture: prescribed supply differs after the response sequence.
	s, e = s.Move([]string{"deck-1-hearts-02"}, 3, Hand, false)
	if e != nil {
		t.Fatal(e)
	}
	n, events, e := Apply(s, Command{ID: "outcome", GameID: s.GameID, Actor: 1, Kind: "purchase-outcome", WindowID: "buy"})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := n.Card(payment)
	if c.Controller != 1 || c.Zone != Hand || n.Pending != nil || len(events) != 1 || events[0].Kind != "purchase-invalidated" {
		t.Fatal("failed full supply spent payment or resized purchase")
	}
	for _, v := range n.Cards {
		if v.Controller == 1 && v.Card.ID != payment {
			t.Fatal("partial purchase drew a card")
		}
	}
}
