package game

import "testing"

func TestJusticeExplicitEmptyClosedChoiceSkipsWithoutCapture(t *testing.T) {
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
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: string(rune('a' + seat)), Kind: "pass", Actor: seat, WindowID: "j"})
		if e != nil {
			t.Fatal(e)
		}
	}
	if n.Pending == nil || n.Pending.Decision == nil {
		t.Fatal("missing decision")
	}
	before := n
	c := Command{GameID: s.GameID, ID: "closed", Kind: "decision-closed", Actor: 1, WindowID: "j", DecisionID: n.Pending.Decision.ID}
	n, _, e = Apply(n, c)
	if e != nil {
		t.Fatal("explicit closed choice with zero hidden eligible must resolve no capture", e)
	}
	if n.Pending != nil {
		t.Fatal("empty hidden choice left wait")
	}
	v, _ := n.Card("deck-1-hearts-08")
	if v.Controller != 2 {
		t.Fatal("captured exposed card despite closed choice")
	}
	v, _ = n.Card(queens[0])
	if v.Zone != Draw {
		t.Fatal("Justice success cost not returned")
	}
	if Digest(before) == Digest(n) {
		t.Fatal("choice did not advance")
	}
}
