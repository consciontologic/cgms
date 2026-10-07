package game

import "testing"

func TestAbilityRingleaderCostAfterResponses(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-1-diamonds-13"}, 1, Hand, false)
	s, e := AttachRoyal(s, 1, "deck-1-diamonds-13", Hearts)
	if e != nil {
		t.Fatal(e)
	}
	n, e := DeclareAbility(s, Command{ID: "ringleader", Actor: 1, Kind: "ringleader", Cards: []string{"deck-1-diamonds-13"}})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := n.Card("deck-1-diamonds-13")
	if c.Zone != Attachment {
		t.Fatal("paid resolution cost early")
	}
	if n.Players[0].Quotas["ringleader"] != 1 {
		t.Fatal("game quota not committed")
	}
	if s.Players[0].Quotas["ringleader"] != 0 {
		t.Fatal("input aliased")
	}
	n.Pending.Cursor = len(n.Pending.Responders)
	n, events, e := ResolveAbility(n, nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	c, _ = n.Card("deck-1-diamonds-13")
	if c.Zone != Discard || len(events) != 1 || events[0].Amount.Cmp(IntAmount(10)) != 0 {
		t.Fatal("Ringleader canonical award/cost missing")
	}
}

func readyAbility(t *testing.T, s State, c Command) State {
	t.Helper()
	n, e := DeclareAbility(s, c)
	if e != nil {
		t.Fatal(e)
	}
	n.Pending.Cursor = len(n.Pending.Responders)
	return n
}
func TestAbilitySupportLossKeepsQuotaAndCost(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-1-diamonds-13"}, 1, Hand, false)
	s, _ = AttachRoyal(s, 1, "deck-1-diamonds-13", Hearts)
	n := readyAbility(t, s, Command{ID: "r", Kind: "ringleader", Actor: 1, Cards: []string{"deck-1-diamonds-13"}})
	n, _ = n.Move([]string{"deck-1-hearts-02"}, 0, Discard, true)
	n, ev, e := ResolveAbility(n, nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	c, _ := n.Card("deck-1-diamonds-13")
	if c.Zone != Attachment || n.Players[0].Quotas["ringleader"] != 1 || len(ev) != 1 || ev[0].Kind != "ability-invalidated" {
		t.Fatal("invalidated resolution charged cost or refunded quota")
	}
}
func TestAbilityBarricadeDrawRestrictions(t *testing.T) {
	s := formationFixture(t)
	ids := []string{"deck-1-hearts-12", "deck-1-hearts-03", "deck-1-hearts-04", "deck-1-hearts-05", "deck-1-hearts-06", "deck-1-hearts-07"}
	s, _ = s.Move(ids, 1, Hand, false)
	s, _ = AttachRoyal(s, 1, ids[0], Hearts)
	n := readyAbility(t, s, Command{ID: "b", Kind: "barricade-sacrifice", Actor: 1, Cards: ids})
	supplied := append(append([]string{}, s.DrawOrder...), ids[1:]...)
	if _, _, e := ResolveAbility(n, nil, nil); e == nil {
		t.Fatal("missing mandatory shuffle accepted")
	}
	done, _, e := ResolveAbility(n, supplied, nil)
	if e != nil {
		t.Fatal(e)
	}
	for _, id := range supplied[:5] {
		c, _ := done.Card(id)
		if c.Controller != 1 || c.AvailableFromRound != 2 {
			t.Fatal("missing next-round marker")
		}
	}
	c, _ := done.Card(ids[0])
	if c.Zone != Discard {
		t.Fatal("queen not discarded")
	}
}
func TestAbilityFateReturnsCostAndPreservesPersistentAce(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-2-diamonds-02", "deck-2-hearts-02"}, 2, Series, false)
	s.Players[1].History = []Suit{Diamonds, Hearts}
	k := []string{"deck-1-hearts-13", "deck-1-clubs-13", "deck-1-spades-13", "deck-1-diamonds-13"}
	s, _ = s.Move(k, 1, Hand, false)
	s, _ = OpenFormation(s, 1, "f", FormationSpec{Kind: "fate", Cards: k})
	s, _ = s.Move([]string{"deck-1-hearts-01"}, 2, ConcealedAce, false)
	s, _ = s.ActivateEffect(AceEffect{ID: "comp", CardID: "deck-1-hearts-01", Kind: "compensation", Source: 2, Target: 2, Custodian: 2, Expiry: "game-end", Losses: 3})
	n := readyAbility(t, s, Command{ID: "fate", Kind: "fate", Actor: 1, FormationID: "f", TargetSeat: 2, Cards: k[:1]})
	order := append(append([]string{}, s.DrawOrder...), k[0], "deck-2-hearts-02", "deck-2-diamonds-02")
	done, _, e := ResolveAbility(n, order, nil)
	if e != nil {
		t.Fatal(e)
	}
	c, _ := done.Card("deck-1-hearts-01")
	if c.Zone != ActiveAce || done.Effects[0].Losses != 3 {
		t.Fatal("Fate reset persistent effect")
	}
	personal := 0
	for _, c := range done.Cards {
		if c.Controller == 2 && c.Zone != ActiveAce {
			personal++
		}
	}
	if personal != 2 {
		t.Fatalf("Fate redraw count %d", personal)
	}
	if FormationFunctioning(done, "f") {
		t.Fatal("Fate cost not paid")
	}
}
func TestAbilityCompensationActivationAndMainInflation(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-1-hearts-01", "deck-1-spades-01"}, 1, ConcealedAce, false)
	n := readyAbility(t, s, Command{ID: "comp", Kind: "compensation", Actor: 1, AceID: "deck-1-hearts-01"})
	if !n.Players[0].CompensationUsed || n.Players[0].AceRound != 1 {
		t.Fatal("acceptance quotas missing")
	}
	n, _, e := ResolveAbility(n, nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	if len(n.Effects) != 1 || n.Effects[0].Kind != "compensation" {
		t.Fatal("effect missing")
	}
	if _, e = DeclareAbility(n, Command{ID: "comp2", Kind: "compensation", Actor: 1, AceID: "deck-1-hearts-01"}); e == nil {
		t.Fatal("second activation accepted")
	}
	s, _ = s.Move([]string{"deck-2-hearts-02", "deck-2-diamonds-02"}, 2, Series, false)
	s.Players[1].History = []Suit{Hearts, Diamonds}
	n = readyAbility(t, s, Command{ID: "infl", Kind: "main-inflation", Actor: 1, TargetSeat: 2, AceID: "deck-1-spades-01"})
	n, _, e = ResolveAbility(n, nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	if len(n.Effects) != 1 || n.Effects[0].Target != 2 {
		t.Fatal("wrong inflation target")
	}
}

func TestAbilityKidnapperClosedSelectionMandatory(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-2-hearts-02", "deck-2-diamonds-02"}, 2, Series, false)
	s.Players[1].History = []Suit{Hearts, Diamonds}
	kids := []string{"deck-1-clubs-11", "deck-2-clubs-11"}
	s, _ = s.Move(kids, 1, Hand, false)
	s, _ = OpenFormation(s, 1, "kid", FormationSpec{Kind: "kidnapper", Cards: kids})
	s, _ = s.Move([]string{"deck-2-spades-01"}, 2, ConcealedAce, false)
	s, _ = s.Move([]string{"deck-2-spades-12"}, 2, Unassigned, false)
	c := Command{ID: "k", Actor: 1, Kind: "kidnapper", FormationID: "kid", TargetSeat: 2, Targets: []string{"deck-2-spades-12"}}
	if _, e := DeclareAbility(s, c); e == nil {
		t.Fatal("chose exposed card despite closed candidate")
	}
	c.Targets = nil
	n := readyAbility(t, s, c)
	if _, _, e := ResolveKidnapper(n, "deck-2-spades-12"); e == nil {
		t.Fatal("noncandidate sampled")
	}
	done, _, e := ResolveKidnapper(n, "deck-2-spades-01")
	if e != nil {
		t.Fatal(e)
	}
	ace, _ := done.Card("deck-2-spades-01")
	if ace.Controller != 1 || ace.Zone != ConcealedAce {
		t.Fatal("ace acquisition destination")
	}
}

func TestAbilityDexterRestrictedThresholdAndBindingRelease(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-1-diamonds-10", "deck-2-diamonds-10"}, 1, Series, false)
	for i := range s.Cards {
		if s.Cards[i].Card.ID == "deck-2-diamonds-10" {
			s.Cards[i].AvailableFromRound = 2
		}
	}
	s, _ = s.Move([]string{"deck-1-diamonds-11"}, 1, Hand, false)
	s, _ = AttachRoyal(s, 1, "deck-1-diamonds-11", Diamonds)
	s, _ = s.Move([]string{"deck-2-hearts-02", "deck-2-diamonds-02"}, 2, Series, false)
	s.Players[1].History = []Suit{Hearts, Diamonds}
	s, _ = s.Move([]string{"deck-2-clubs-13"}, 2, Unassigned, false)
	n := readyAbility(t, s, Command{ID: "dex", Kind: "dexter-assassination", Actor: 1, Cards: []string{"deck-1-diamonds-02"}, Targets: []string{"deck-2-clubs-13"}})
	n, _, e := ResolveAbility(n, nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	c, _ := n.Card("deck-2-clubs-13")
	if c.Zone != Discard {
		t.Fatal("assassination failed")
	}
	pay, _ := n.Card("deck-1-diamonds-02")
	if pay.Zone != Draw || !n.NeedsShuffle {
		t.Fatal("diamond payment missing")
	}
}
func TestAbilityRicherRequiresSingleAttachedKing(t *testing.T) {
	s := formationFixture(t)
	ids := []string{"deck-1-clubs-13", "deck-1-spades-13"}
	s, _ = s.Move(ids, 1, Hand, false)
	s, _ = AttachRoyal(s, 1, ids[0], Hearts)
	n := readyAbility(t, s, Command{ID: "rich", Kind: "richer-sacrifice", Actor: 1, Cards: ids[:1]})
	n, _, e := ResolveAbility(n, nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	for _, id := range s.DrawOrder[:2] {
		c, _ := n.Card(id)
		if c.Controller != 1 {
			t.Fatal("missing Richer draw")
		}
	}
	s, _ = AttachRoyal(s, 1, ids[1], Hearts)
	if _, e = DeclareAbility(s, Command{ID: "bad", Kind: "richer-sacrifice", Actor: 1, Cards: ids[:1]}); e == nil {
		t.Fatal("two-king series Richer accepted")
	}
}
func TestAbilityCodeSettingsPersistAfterCardLoss(t *testing.T) {
	s := formationFixture(t)
	ids := []string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-clubs-12", "deck-2-clubs-12", "deck-1-spades-12"}
	s, _ = s.Move(ids, 1, Hand, false)
	s, _ = OpenFormation(s, 1, "code", FormationSpec{Kind: "code", Cards: ids})
	n := readyAbility(t, s, Command{ID: "cd", Kind: "code", Actor: 1, FormationID: "code", Value: 1, Price: 10})
	n, _, e := ResolveAbility(n, nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	n, e = n.Move(ids, 0, Draw, true)
	if e != nil {
		t.Fatal(e)
	}
	if n.Code == nil || n.Code.DiamondBonus != 1 || n.Code.PurchasePrice != 10 {
		t.Fatal("Code settings did not survive formation loss")
	}
}

func TestAbilityAssassinationCountsAttachedLossForCompensation(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-1-diamonds-10", "deck-2-diamonds-10"}, 1, Series, false)
	s, _ = s.Move([]string{"deck-1-diamonds-11"}, 1, Hand, false)
	s, _ = AttachRoyal(s, 1, "deck-1-diamonds-11", Diamonds)
	s, _ = s.Move([]string{"deck-2-hearts-02", "deck-2-diamonds-02"}, 2, Series, false)
	s.Players[1].History = []Suit{Hearts, Diamonds}
	s, _ = s.Move([]string{"deck-2-clubs-13"}, 2, Attachment, false)
	for i := range s.Cards {
		if s.Cards[i].Card.ID == "deck-2-clubs-13" {
			s.Cards[i].Allocation = string(Hearts)
		}
	}
	s, _ = s.Move([]string{"deck-2-hearts-01"}, 2, ConcealedAce, false)
	s, _ = s.ActivateEffect(AceEffect{ID: "comp", CardID: "deck-2-hearts-01", Kind: "compensation", Source: 2, Target: 2, Custodian: 2, Expiry: "game-end", Losses: 4})
	n := readyAbility(t, s, Command{ID: "dex", Kind: "dexter-assassination", Actor: 1, Cards: []string{"deck-1-diamonds-02"}, Targets: []string{"deck-2-clubs-13"}})
	n, _, e := ResolveAbility(n, nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	if n.Effects[0].Losses != 5 {
		t.Fatalf("attached loss counter %d want5", n.Effects[0].Losses)
	}
}

func TestKidnapperDeclarationDoesNotRevealClosedRoyalEligibility(t *testing.T) {
	s := formationFixture(t)
	s.Players[1].History = []Suit{Hearts, Clubs}
	ids := []string{"deck-1-clubs-11", "deck-2-clubs-11"}
	s, _ = s.Move(ids, 1, Hand, true)
	s.Players[0].OpeningUsed = false
	var e error
	s, e = OpenFormation(s, 1, "kid", FormationSpec{Kind: "kidnapper", Cards: ids})
	if e != nil {
		t.Fatal(e)
	}
	c := Command{GameID: s.GameID, ID: "kid", Actor: 1, Kind: "kidnapper", FormationID: "kid", TargetSeat: 2}
	a, e := DeclareAbility(s, c)
	if e != nil {
		t.Fatal(e)
	}
	s, _ = s.Move([]string{"deck-1-spades-13"}, 2, Hand, true)
	b, e := DeclareAbility(s, c)
	if e != nil {
		t.Fatal(e)
	}
	if a.Pending == nil || b.Pending == nil {
		t.Fatal("target activation missing")
	}
}
