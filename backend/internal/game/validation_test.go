package game

import "testing"

func TestImportedStateRejectsMalformedMetadata(t *testing.T) {
	cases := map[string]func(*State){
		"negative round": func(s *State) { s.Round = -1 }, "zero turn": func(s *State) { s.Turn = 0 }, "invalid active": func(s *State) { s.Active = 9 }, "negative version": func(s *State) { s.Version = -1 }, "unknown phase": func(s *State) { s.Phase = "invented" },
		"duplicate history": func(s *State) { s.Players[0].History = []Suit{Hearts, Hearts} }, "unknown suit history": func(s *State) { s.Players[0].History = []Suit{"unknown"} }, "invalid ally": func(s *State) { s.Players[0].Ally = 8 }, "unknown public card": func(s *State) { s.PublicHistory = []string{"unknown"} }, "duplicate public history": func(s *State) { s.PublicHistory = []string{s.Cards[0].Card.ID, s.Cards[0].Card.ID} }, "invalid command digest": func(s *State) { s.Commands["x"] = "bogus" }, "invalid order": func(s *State) { s.Order = []int{1, 1, 3} },
		"invalid pending actor": func(s *State) { s.Pending = &PendingAction{ID: "pending", Kind: "justice", Actor: 9} },
		"invalid pending cursor": func(s *State) {
			s.Pending = &PendingAction{ID: "pending", Kind: "justice", Actor: 1, Responders: []int{2}, Cursor: -1}
		},
		"empty justice choice": func(s *State) {
			s.Pending = &PendingAction{ID: "pending", Kind: "justice", Actor: 1, JusticeQueen: cid(Hearts, 12), Decision: &EffectDecision{ID: "d", EffectID: "e", Kind: "justice", Stage: "selection", Actor: 1, Opponents: []int{2}, QueenID: cid(Hearts, 12), Choices: [][]string{{}}}}
		},
		"invalid decision actor": func(s *State) {
			s.Pending = &PendingAction{ID: "pending", Kind: "justice", Actor: 1, JusticeQueen: cid(Hearts, 12), Decision: &EffectDecision{ID: "d", EffectID: "e", Kind: "justice", Stage: "selection", Actor: 99, Opponents: []int{2}, QueenID: cid(Hearts, 12), Choices: [][]string{{cid(Clubs, 2)}}}}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			s := position(t)
			mutate(&s)
			if s.Validate() == nil {
				t.Fatal("malformed imported state accepted")
			}
		})
	}
}

func TestImportedStateRejectsEffectAndBindingCorruption(t *testing.T) {
	base := compensationPosition(t)
	for name, mutate := range map[string]func(*State){"effect identity": func(s *State) { s.Effects[1].ID = s.Effects[0].ID }, "negative losses": func(s *State) { s.Effects[0].Losses = -1 }, "missing expiry": func(s *State) { s.Effects[0].Expiry = "" }, "wrong Ace suit": func(s *State) { s.Effects[0].Kind = "inflation" }, "duplicate compensation target": func(s *State) {
		s.Effects[1].Target = 2
		s.Effects[1].Source = 2
		s.Effects[1].Custodian = 2
		for i := range s.Cards {
			if s.Cards[i].Card.ID == s.Effects[1].CardID {
				s.Cards[i].Controller = 2
			}
		}
	}, "duplicate binding ID": func(s *State) {
		s.Bindings = []Binding{{ID: "same", Kings: []string{cid(Hearts, 13), cid(Clubs, 13)}, Slot: "a"}, {ID: "same", Kings: []string{cid(Spades, 13), cid(Diamonds, 13)}, Slot: "b"}}
	}} {
		t.Run(name, func(t *testing.T) {
			s := base.Clone()
			mutate(&s)
			if s.Validate() == nil {
				t.Fatal("invalid effect/binding accepted")
			}
		})
	}
}
func TestImportedStateValidContinuationNeverPanics(t *testing.T) {
	s := goldenPosition(t)
	s, _, e := Apply(s, Command{GameID: s.GameID, ID: "validation", Kind: "attack", Actor: 1, Cards: []string{cid(Clubs, 8)}, Targets: []string{cid(Hearts, 8)}})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Validate(); e != nil {
		t.Fatal(e)
	}
	for seat := 1; seat <= 3; seat++ {
		if _, e = Observe(s, seat); e != nil {
			t.Fatal(e)
		}
	}
}
func TestImportedStateRejectsPartialOrderAsymmetricAllianceAndAttachedBinding(t *testing.T) {
	for name, mutate := range map[string]func(*State){"partial-order": func(s *State) { s.Order = []int{1} }, "asymmetric-ally": func(s *State) { s.Players[0].Ally = 2 }, "bound-attachment": func(s *State) {
		ids := []string{cid(Hearts, 13), cid(Clubs, 13)}
		*s, _ = s.Move(ids, 1, Formation, true)
		*s, _ = s.Bind(Binding{ID: "b", Kings: ids, Slot: "queen-hearts", Formation: "people", Supported: false})
		for i := range s.Cards {
			if s.Cards[i].Card.ID == ids[0] {
				s.Cards[i].Zone = Attachment
			}
		}
	}} {
		t.Run(name, func(t *testing.T) {
			s := position(t)
			mutate(&s)
			if s.Validate() == nil {
				t.Fatal("inconsistent imported metadata accepted")
			}
		})
	}
}
func TestImportedNegotiatorRequiresStageActor(t *testing.T) {
	s := combatFixture(t)
	s.Pending = &PendingAction{ID: "p", Actor: 1, Defender: 2, Clubs: []string{cid(Clubs, 8)}, Targets: []string{cid(Hearts, 8)}, Responders: []int{2, 3}, Decision: &EffectDecision{ID: "d", EffectID: "effect", Actor: 3, Kind: "negotiator", Stage: "clubs", Choices: [][]string{{cid(Clubs, 8)}}}}
	if s.Validate() == nil {
		t.Fatal("foreign actor may choose attacker clubs")
	}
	s.Pending.Decision.Stage = "payment"
	s.Pending.Decision.Actor = 1
	if s.Validate() == nil {
		t.Fatal("attacker may choose defender payment")
	}
}
func TestImportedDexterRequiresHostileTargetActor(t *testing.T) {
	s := combatFixture(t)
	s.Pending = &PendingAction{ID: "p", Actor: 1, Defender: 2, Clubs: []string{cid(Clubs, 8)}, Targets: []string{cid(Hearts, 8)}, Responders: []int{2, 3}, Decision: &EffectDecision{ID: "d", EffectID: "effect", Actor: 3, AceActor: 2, AceID: cid(Diamonds, 1), Kind: "dexter", Stage: "confinement", Choices: [][]string{{}}}}
	if s.Validate() == nil {
		t.Fatal("third-party Dexter can prevent hostile ace")
	}
}
func TestDecisionRejectsForgedPaymentPool(t *testing.T) {
	s := combatFixture(t)
	s, _ = s.Move([]string{cid(Spades, 11)}, 2, Attachment, true)
	s, _ = s.Move([]string{cid(Spades, 8)}, 2, Series, true)
	s, _, e := Apply(s, Command{GameID: s.GameID, ID: "a", Kind: "attack", Actor: 1, Cards: []string{cid(Clubs, 8)}, Targets: []string{cid(Hearts, 8)}})
	if e != nil {
		t.Fatal(e)
	}
	s, _, e = Apply(s, Command{GameID: s.GameID, ID: "n", Kind: "negotiate", Actor: 2, WindowID: "a"})
	if e != nil {
		t.Fatal(e)
	}
	s.Pending.Decision.Choices = [][]string{{cid(Diamonds, 10)}}
	before := Digest(s)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "forged", Kind: "decision", Actor: 2, WindowID: "a", DecisionID: "n/payment", Cards: []string{cid(Diamonds, 10)}})
	if e == nil || Digest(n) != before {
		t.Fatal("forged saved choice moved unrelated physical card")
	}
}
