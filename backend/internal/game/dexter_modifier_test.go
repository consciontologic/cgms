package game

import "testing"

func dexterModifierFixture(t *testing.T, attackRank string, targetIDs []string) State {
	t.Helper()
	s, _ := NewState(3, "dexter-modifier")
	moves := []struct {
		ids  []string
		seat int
		zone Zone
	}{{[]string{attackRank, "deck-1-diamonds-02", "deck-1-diamonds-03"}, 1, Series}, {targetIDs, 2, Series}, {[]string{"deck-2-diamonds-02", "deck-2-diamonds-03"}, 2, Series}}
	for _, m := range moves {
		var e error
		s, e = s.Move(m.ids, m.seat, m.zone, true)
		if e != nil {
			t.Fatal(e)
		}
	}
	s.Players[1].History = []Suit{Hearts, Clubs, Spades, Diamonds}
	return s
}
func TestDexterDefensiveModifierPrevention(t *testing.T) {
	for _, kind := range []string{"numerical-defense", "advancement-defense"} {
		t.Run(kind, func(t *testing.T) {
			s := dexterModifierFixture(t, "deck-1-clubs-08", []string{"deck-1-hearts-08"})
			s, _ = s.Move([]string{"deck-1-diamonds-11"}, 1, Attachment, true)
			s, _ = s.Move([]string{"deck-1-clubs-01"}, 2, ConcealedAce, true)
			n, _, e := Apply(s, Command{GameID: s.GameID, ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
			if e != nil {
				t.Fatal(e)
			}
			n, _, e = Apply(n, Command{GameID: s.GameID, ID: "defend", Kind: kind, Actor: 2, WindowID: "attack", AceID: "deck-1-clubs-01", Value: 3})
			if e != nil {
				t.Fatal(e)
			}
			if n.Pending == nil || n.Pending.Decision == nil || n.Pending.Decision.Actor != 1 || n.Pending.Cursor != 0 {
				t.Fatal("hostile defensive modifier skipped Dexter decision")
			}
			if n.Players[1].AceRound != s.Round {
				t.Fatal("ace allowance uncommitted")
			}
			answer := Command{GameID: s.GameID, ID: "prevent", Kind: "decision", Actor: 1, WindowID: "attack", DecisionID: "defend/dexter", Cards: []string{"deck-1-diamonds-02"}}
			n, _, e = Apply(n, answer)
			if e != nil {
				t.Fatal(e)
			}
			if n.Pending.DefenseModifier != 0 || n.Pending.Cursor != 1 || n.Players[0].Quotas["dexter"] != s.Round {
				t.Fatal("prevention failed or consumed ordinary window")
			}
			a, _ := n.Card("deck-1-clubs-01")
			payment, _ := n.Card("deck-1-diamonds-02")
			if a.Zone != Discard || payment.Zone != Draw {
				t.Fatal("physical prevention costs")
			}
			same, _, e := Apply(n, answer)
			if e != nil || Digest(same) != Digest(n) {
				t.Fatal("retry duplicated payment")
			}
			n, _, e = Apply(n, Command{GameID: s.GameID, ID: "pass3", Kind: "pass", Actor: 3, WindowID: "attack"})
			if e != nil {
				t.Fatal(e)
			}
			target, _ := n.Card("deck-1-hearts-08")
			if target.Zone != Draw {
				t.Fatal("prevented modifier still affects final match")
			}
			if e = n.Validate(); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestDexterOffensiveModifierDecision(t *testing.T) {
	for _, prevent := range []bool{false, true} {
		s := dexterModifierFixture(t, "deck-1-clubs-05", []string{"deck-1-hearts-08"})
		s, _ = s.Move([]string{"deck-1-diamonds-11"}, 2, Attachment, true)
		s, _ = s.Move([]string{"deck-1-clubs-01"}, 1, ConcealedAce, true)
		n, _, e := Apply(s, Command{GameID: s.GameID, ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-05"}, Targets: []string{"deck-1-hearts-08"}, AceID: "deck-1-clubs-01", Value: 3})
		if e != nil {
			t.Fatal(e)
		}
		if n.Pending.Decision == nil || n.Pending.Decision.Actor != 2 {
			t.Fatal("offensive modifier skipped Dexter")
		}
		cards := []string{}
		if prevent {
			cards = []string{"deck-2-diamonds-02"}
		}
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: "answer", Kind: "decision", Actor: 2, WindowID: "attack", DecisionID: "attack/dexter", Cards: cards})
		if e != nil {
			t.Fatal(e)
		}
		if n.Pending.Cursor != 0 {
			t.Fatal("Dexter consumed ordinary defense opportunity")
		}
		for _, seat := range []int{2, 3} {
			n, _, e = Apply(n, Command{GameID: s.GameID, ID: string(rune('a' + seat)), Kind: "pass", Actor: seat, WindowID: "attack"})
			if e != nil {
				t.Fatal(e)
			}
		}
		target, _ := n.Card("deck-1-hearts-08")
		if (target.Zone == Draw) == prevent {
			t.Fatal("offensive modifier outcome")
		}
		club, _ := n.Card("deck-1-clubs-05")
		ace, _ := n.Card("deck-1-clubs-01")
		if club.UsedTurn != s.Turn || ace.Zone != Discard || n.Players[0].AceRound != s.Round {
			t.Fatal("consumed allowances lost")
		}
	}
}

func TestDexterAdvancementAttackAndInvalidPayment(t *testing.T) {
	s := dexterModifierFixture(t, "deck-1-clubs-05", []string{"deck-1-hearts-07", "deck-1-hearts-08"})
	s, _ = s.Move([]string{"deck-1-diamonds-11"}, 2, Attachment, true)
	s, _ = s.Move([]string{"deck-1-clubs-01"}, 1, ConcealedAce, true)
	for i := range s.Cards {
		if s.Cards[i].Card.ID == "deck-2-diamonds-02" {
			s.Cards[i].AvailableFromRound = s.Round + 1
		}
	}
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-05"}, Targets: []string{"deck-1-hearts-07", "deck-1-hearts-08"}, AceID: "deck-1-clubs-01"})
	if e != nil || n.Pending.Decision == nil {
		t.Fatal("Advancement must offer Dexter", e)
	}
	before := Digest(n)
	bad := Command{GameID: s.GameID, ID: "bad", Kind: "decision", Actor: 2, WindowID: "attack", DecisionID: "attack/dexter", Cards: []string{"deck-2-diamonds-02"}}
	unchanged, _, e := Apply(n, bad)
	if e == nil || Digest(unchanged) != before {
		t.Fatal("restricted payment accepted or state changed")
	}
	bad.ID = "good"
	bad.Cards = []string{"deck-2-diamonds-03"}
	n, _, e = Apply(n, bad)
	if e != nil || n.Pending.AttackModifier != 0 {
		t.Fatal("Advancement not prevented", e)
	}
}
func TestDexterPaymentCannotBeRecapturedAsFrozenTarget(t *testing.T) {
	s := dexterModifierFixture(t, "deck-1-clubs-02", []string{})
	s, _ = s.Move([]string{"deck-1-diamonds-11"}, 2, Attachment, true)
	s, _ = s.Move([]string{"deck-1-clubs-01"}, 1, ConcealedAce, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-02"}, Targets: []string{"deck-2-diamonds-02", "deck-2-diamonds-03"}, AceID: "deck-1-clubs-01", Value: 3})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "prevent", Kind: "decision", Actor: 2, WindowID: "attack", DecisionID: "attack/dexter", Cards: []string{"deck-2-diamonds-03"}})
	if e != nil {
		t.Fatal(e)
	}
	for _, seat := range []int{2, 3} {
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: string(rune('a' + seat)), Kind: "pass", Actor: seat, WindowID: "attack"})
		if e != nil {
			t.Fatal(e)
		}
	}
	paid, _ := n.Card("deck-2-diamonds-03")
	if paid.Zone != Draw || paid.Controller != 0 {
		t.Fatal("Dexter payment was recaptured")
	}
	remaining, _ := n.Card("deck-2-diamonds-02")
	if remaining.Controller != 2 {
		t.Fatal("invalidated frozen subset still captured")
	}
}

func TestDexterDeclinePreservesDefensiveModifier(t *testing.T) {
	s := dexterModifierFixture(t, "deck-1-clubs-08", []string{"deck-1-hearts-08"})
	s, _ = s.Move([]string{"deck-1-diamonds-11"}, 1, Attachment, true)
	s, _ = s.Move([]string{"deck-1-clubs-01"}, 2, ConcealedAce, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "defend", Kind: "advancement-defense", Actor: 2, WindowID: "attack", AceID: "deck-1-clubs-01"})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "decline", Kind: "decision", Actor: 1, WindowID: "attack", DecisionID: "defend/dexter", Cards: []string{}})
	if e != nil {
		t.Fatal(e)
	}
	if n.Pending.DefenseModifier != 10 || n.Players[0].Quotas["dexter"] == n.Round {
		t.Fatal("decline altered effect/quota")
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "pass", Kind: "pass", Actor: 3, WindowID: "attack"})
	if e != nil {
		t.Fatal(e)
	}
	target, _ := n.Card("deck-1-hearts-08")
	ace, _ := n.Card("deck-1-clubs-01")
	if target.Controller != 2 || ace.Zone != Discard {
		t.Fatal("surviving modifier/cost lifecycle")
	}
}
