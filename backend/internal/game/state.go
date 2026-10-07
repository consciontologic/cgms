package game

import (
	"errors"
	"fmt"
	"slices"
)

type Zone string

const (
	Draw         Zone = "draw"
	Discard      Zone = "discard"
	Hand         Zone = "hand"
	ConcealedAce Zone = "concealed-ace"
	Series       Zone = "series"
	Attachment   Zone = "attachment"
	Formation    Zone = "formation"
	Unassigned   Zone = "unassigned"
	ActiveAce    Zone = "active-ace"
	Initiative   Zone = "initiative"
)

type PlacedCard struct {
	Card               Card   `json:"card"`
	Zone               Zone   `json:"zone"`
	Controller         int    `json:"controller"`
	AvailableFromRound int    `json:"available_from_round"`
	Allocation         string `json:"allocation"`
	UsedTurn           int    `json:"used_turn"`
}
type Player struct {
	Seat             int            `json:"seat"`
	History          []Suit         `json:"history"`
	OpeningUsed      bool           `json:"opening_used"`
	Departed         bool           `json:"departed"`
	Confined         bool           `json:"confined"`
	Connected        bool           `json:"connected"`
	Underground      bool           `json:"underground"`
	Ally             int            `json:"ally"`
	AceRound         int            `json:"ace_round"`
	CompensationUsed bool           `json:"compensation_used"`
	Quotas           map[string]int `json:"quotas"`
}
type Binding struct {
	ID        string   `json:"id"`
	Kings     []string `json:"kings"`
	Slot      string   `json:"slot"`
	Formation string   `json:"formation"`
	Supported bool     `json:"supported"`
}
type AceEffect struct {
	ID        string `json:"id"`
	CardID    string `json:"card_id"`
	Kind      string `json:"kind"`
	Source    int    `json:"source"`
	Target    int    `json:"target"`
	Custodian int    `json:"custodian"`
	Expiry    string `json:"expiry"`
	Losses    int    `json:"losses"`
}
type FormationLoan struct {
	Cards      []string `json:"cards"`
	From       int      `json:"from"`
	To         int      `json:"to"`
	Allocation string   `json:"allocation"`
}
type State struct {
	Proposals     []Proposal        `json:"proposals,omitempty"`
	Formations    []FormationRecord `json:"formations,omitempty"`
	Code          *CodeSettings     `json:"code,omitempty"`
	Loans         []FormationLoan   `json:"loans"`
	Pending       *PendingAction    `json:"pending,omitempty"`
	Decisions     []DecisionReceipt `json:"decisions"`
	Version       int               `json:"version"`
	Commands      map[string]string `json:"commands"`
	NeedsShuffle  bool              `json:"needs_shuffle"`
	DrawQueue     []DrawEntitlement `json:"draw_queue"`
	Schema        string            `json:"schema"`
	GameID        string            `json:"game_id"`
	Round         int               `json:"round"`
	Turn          int               `json:"turn"`
	Active        int               `json:"active"`
	Phase         string            `json:"phase"`
	Cards         []PlacedCard      `json:"cards"`
	Players       []Player          `json:"players"`
	DrawOrder     []string          `json:"draw_order"`
	Bindings      []Binding         `json:"bindings"`
	Effects       []AceEffect       `json:"effects"`
	Order         []int             `json:"order"`
	PublicHistory []string          `json:"public_history"`
}

func NewState(n int, id string) (State, error) {
	if (n != 3 && n != 4) || id == "" {
		return State{}, errors.New("invalid population/game identity")
	}
	s := State{Commands: map[string]string{}, Schema: "cgms-state-v1", GameID: id, Round: 1, Turn: 1, Active: 1, Phase: "playing"}
	for _, c := range Deck() {
		s.Cards = append(s.Cards, PlacedCard{Card: c, Zone: Draw})
		s.DrawOrder = append(s.DrawOrder, c.ID)
	}
	for i := 1; i <= n; i++ {
		s.Players = append(s.Players, Player{Seat: i, Connected: true, Quotas: map[string]int{}})
	}
	return s, nil
}
func (s State) Clone() State {
	n := s
	n.Formations = cloneFormations(s.Formations)
	n.Proposals = cloneProposals(s.Proposals)
	if s.Code != nil {
		v := *s.Code
		n.Code = &v
	}
	cloneTransition(&n, s)
	n.Loans = slices.Clone(s.Loans)
	for i := range n.Loans {
		n.Loans[i].Cards = slices.Clone(s.Loans[i].Cards)
	}
	n.DrawQueue = slices.Clone(s.DrawQueue)
	n.Cards = slices.Clone(s.Cards)
	n.DrawOrder = slices.Clone(s.DrawOrder)
	n.Order = slices.Clone(s.Order)
	n.PublicHistory = slices.Clone(s.PublicHistory)
	n.Effects = slices.Clone(s.Effects)
	n.Bindings = slices.Clone(s.Bindings)
	for i := range n.Bindings {
		n.Bindings[i].Kings = slices.Clone(s.Bindings[i].Kings)
	}
	n.Players = slices.Clone(s.Players)
	for i := range n.Players {
		n.Players[i].History = slices.Clone(s.Players[i].History)
		n.Players[i].Quotas = map[string]int{}
		for k, v := range s.Players[i].Quotas {
			n.Players[i].Quotas[k] = v
		}
	}
	return n
}
func (s State) Validate() error {
	if err := validateMetadata(s); err != nil {
		return err
	}
	if err := validateDrawQueue(s); err != nil {
		return err
	}
	if s.Schema != "cgms-state-v1" || s.GameID == "" || (len(s.Players) != 3 && len(s.Players) != 4) || len(s.Cards) != 104 {
		return errors.New("invalid state identity or partition")
	}
	catalog := map[string]Card{}
	for _, c := range Deck() {
		catalog[c.ID] = c
	}
	seen := map[string]bool{}
	draw := map[string]bool{}
	for _, c := range s.Cards {
		if seen[c.Card.ID] || catalog[c.Card.ID] != c.Card {
			return errors.New("invalid or duplicate physical card")
		}
		seen[c.Card.ID] = true
		if c.AvailableFromRound < 0 || c.Controller < 0 || c.Controller > len(s.Players) {
			return errors.New("invalid card metadata")
		}
		switch c.Zone {
		case Draw, Discard, Initiative:
			if c.Controller != 0 {
				return errors.New("supply has controller")
			}
		case Hand, ConcealedAce, Series, Attachment, Formation, Unassigned, ActiveAce:
			if c.Controller == 0 {
				return errors.New("personal zone needs controller")
			}
		default:
			return errors.New("unknown zone")
		}
		if c.Zone == ConcealedAce && c.Card.Rank != 1 || c.Zone == Hand && c.Card.Rank == 1 || c.Zone == Series && (c.Card.Rank < 2 || c.Card.Rank > 10) {
			return errors.New("card incompatible with zone")
		}
		if c.Zone == Draw {
			draw[c.Card.ID] = true
		}
	}
	for _, id := range s.DrawOrder {
		if !draw[id] {
			return errors.New("invalid draw order")
		}
		delete(draw, id)
	}
	if len(draw) != 0 {
		return errors.New("missing draw order")
	}
	for i, p := range s.Players {
		if p.Seat != i+1 {
			return errors.New("invalid seat order")
		}
	}
	bound := map[string]bool{}
	for _, b := range s.Bindings {
		if len(b.Kings) != 2 || b.ID == "" || b.Slot == "" {
			return errors.New("invalid binding")
		}
		for _, id := range b.Kings {
			if bound[id] || catalog[id].Rank != 13 {
				return errors.New("overlapping binding")
			}
			bound[id] = true
		}
	}
	active := map[string]bool{}
	inflated := map[int]bool{}
	for _, e := range s.Effects {
		c, ok := s.Card(e.CardID)
		if !ok || active[e.CardID] || c.Zone != ActiveAce || c.Card.Rank != 1 || e.Target != e.Custodian || c.Controller != e.Target || e.Source < 1 || e.Source > len(s.Players) || e.Target < 1 || e.Target > len(s.Players) {
			return errors.New("invalid active effect")
		}
		if e.Kind == "inflation" {
			if inflated[e.Target] {
				return errors.New("stacked inflation")
			}
			inflated[e.Target] = true
		} else if e.Kind != "compensation" || e.Source != e.Target {
			return errors.New("invalid effect kind")
		}
		active[e.CardID] = true
	}
	for _, c := range s.Cards {
		if c.Zone == ActiveAce && !active[c.Card.ID] {
			return errors.New("unrepresented active ace")
		}
	}
	return nil
}
func (s State) Card(id string) (PlacedCard, bool) {
	for _, c := range s.Cards {
		if c.Card.ID == id {
			return c, true
		}
	}
	return PlacedCard{}, false
}

// Move is a state primitive, not authorization for a gameplay action. Forced
// acquisitions preserve restrictions and bindings; callers enforce timing/support.
func (s State) Move(ids []string, seat int, zone Zone, forced bool) (State, error) {
	n := s.Clone()
	seen := map[string]bool{}
	for _, id := range ids {
		c, ok := s.Card(id)
		if !ok || seen[id] {
			return s, errors.New("invalid movement selection")
		}
		seen[id] = true
		if zone == Attachment {
			for _, b := range s.Bindings {
				if slices.Contains(b.Kings, id) {
					return s, errors.New("bound king cannot attach")
				}
			}
		}
		if c.Zone == ActiveAce || (!forced && c.AvailableFromRound > s.Round) {
			return s, errors.New("card unavailable for movement")
		}
		if publicZone(c.Zone) || publicZone(zone) {
			rememberPublic(&n, id)
		}
		for i := range n.Cards {
			if n.Cards[i].Card.ID == id {
				n.Cards[i].Controller = seat
				n.Cards[i].Zone = zone
				n.Cards[i].Allocation = ""
			}
		}
		n.DrawOrder = slices.DeleteFunc(n.DrawOrder, func(x string) bool { return x == id })
		if zone == Draw {
			n.DrawOrder = append(n.DrawOrder, id)
			n.NeedsShuffle = true
		}
	}
	if e := n.Validate(); e != nil {
		return s, e
	}
	return n, nil
}
func (s State) Bind(b Binding) (State, error) {
	n := s.Clone()
	if b.ID == "" || len(b.Kings) != 2 {
		return s, errors.New("invalid binding")
	}
	for _, old := range n.Bindings {
		if old.ID == b.ID {
			return s, errors.New("binding identity exists")
		}
	}
	b.Kings = slices.Clone(b.Kings)
	n.Bindings = append(n.Bindings, b)
	for i := range n.Cards {
		if slices.Contains(b.Kings, n.Cards[i].Card.ID) {
			n.Cards[i].Allocation = b.Formation
		}
	}
	if e := n.Validate(); e != nil {
		return s, e
	}
	return n, nil
}
func (s State) BindingStatus(id string) string {
	for _, b := range s.Bindings {
		if b.ID != id {
			continue
		}
		a, _ := s.Card(b.Kings[0])
		c, _ := s.Card(b.Kings[1])
		if a.Controller != c.Controller || a.Zone != c.Zone {
			return "dormant-separated"
		}
		if a.Controller > 0 && a.Zone == Formation && a.Allocation == b.Formation && c.Allocation == b.Formation && b.Formation != "" && s.bindingSupported(b, a.Controller) {
			return "active"
		}
		return "dormant-together"
	}
	return ""
}
func (s State) ActivateEffect(e AceEffect) (State, error) {
	n := s.Clone()
	c, ok := s.Card(e.CardID)
	if !ok || c.Zone != ConcealedAce || c.Card.Rank != 1 || c.Controller != e.Source || c.AvailableFromRound > s.Round {
		return s, errors.New("invalid ace source")
	}
	if e.Kind == "inflation" && c.Card.Suit != Spades || e.Kind == "compensation" && c.Card.Suit != Hearts {
		return s, errors.New("wrong named ace")
	}
	for i := range n.Cards {
		if n.Cards[i].Card.ID == e.CardID {
			n.Cards[i].Zone = ActiveAce
			n.Cards[i].Controller = e.Target
		}
	}
	n.Effects = append(n.Effects, e)
	if err := n.Validate(); err != nil {
		return s, err
	}
	return n, nil
}
func (s State) FateReturn(seat int) (State, int, error) {
	ids := []string{}
	for _, c := range s.Cards {
		if c.Controller == seat && c.Zone != ActiveAce {
			ids = append(ids, c.Card.ID)
		}
	}
	n, e := s.Move(ids, 0, Draw, true)
	return n, len(ids), e
}
func (s State) CurrentSeries(seat int) []Suit {
	out := []Suit{}
	for _, suit := range []Suit{Hearts, Clubs, Spades, Diamonds} {
		for _, c := range s.Cards {
			if c.Controller == seat && c.Zone == Series && c.Card.Suit == suit {
				out = append(out, suit)
				break
			}
		}
	}
	return out
}
func (s State) String() string { return fmt.Sprintf("%s round %d", s.GameID, s.Round) }

// LoanFormation and ReturnLoan are custody primitives; the action layer owns
// offers, consent, response windows, eligibility and alliance admission.
func (s State) LoanFormation(ids []string, from, to int) (State, error) {
	if from == to || from < 1 || to < 1 || from > len(s.Players) || to > len(s.Players) || len(ids) == 0 {
		return s, errors.New("invalid loan parties")
	}
	first, ok := s.Card(ids[0])
	if !ok || first.Allocation == "" {
		return s, errors.New("unallocated formation")
	}
	wanted := []string{}
	for _, c := range s.Cards {
		if c.Controller == from && c.Zone == Formation && c.Allocation == first.Allocation {
			wanted = append(wanted, c.Card.ID)
		}
	}
	if !containsSelection([][]string{wanted}, ids) {
		return s, errors.New("intact formation required")
	}
	for _, id := range ids {
		c, _ := s.Card(id)
		if c.AvailableFromRound > s.Round {
			return s, errors.New("restricted loan")
		}
	}
	n := s.Clone()
	for i := range n.Cards {
		if slices.Contains(ids, n.Cards[i].Card.ID) {
			n.Cards[i].Controller = to
		}
	}
	n.Loans = append(n.Loans, FormationLoan{Cards: slices.Clone(ids), From: from, To: to, Allocation: first.Allocation})
	n.refreshBindingSupport(ids, to)
	if e := n.Validate(); e != nil {
		return s, e
	}
	return n, nil
}
func (s State) ReturnLoan(ids []string, from, to int) (State, error) {
	if len(ids) == 0 {
		return s, errors.New("empty loan return")
	}
	index := -1
	for i, l := range s.Loans {
		if l.From == to && l.To == from {
			for _, id := range ids {
				if slices.Contains(l.Cards, id) {
					index = i
					break
				}
			}
		}
	}
	if index < 0 {
		return s, errors.New("unknown loan")
	}
	loan := s.Loans[index]
	seen := map[string]bool{}
	for _, id := range ids {
		c, ok := s.Card(id)
		if !ok || seen[id] || !slices.Contains(loan.Cards, id) || c.Controller != from || c.AvailableFromRound > s.Round {
			return s, errors.New("invalid loan return selection")
		}
		seen[id] = true
	}
	intact := containsSelection([][]string{loan.Cards}, ids)
	if intact {
		for _, id := range ids {
			c, _ := s.Card(id)
			if c.Zone != Formation || c.Allocation != loan.Allocation {
				intact = false
			}
		}
	}
	n := s.Clone()
	for i := range n.Cards {
		if seen[n.Cards[i].Card.ID] {
			n.Cards[i].Controller = to
			if intact {
				n.Cards[i].Zone = Formation
				n.Cards[i].Allocation = loan.Allocation
			} else {
				n.Cards[i].Zone = Unassigned
				n.Cards[i].Allocation = ""
			}
		} else if !intact && slices.Contains(loan.Cards, n.Cards[i].Card.ID) && n.Cards[i].Controller == from && n.Cards[i].Zone == Formation {
			n.Cards[i].Zone = Unassigned
			n.Cards[i].Allocation = ""
		}
	}
	if intact {
		n.refreshBindingSupport(ids, to)
		for i := range n.Formations {
			if n.Formations[i].ID == loan.Allocation && n.Formations[i].Controller == from {
				n.Formations[i].Controller = to
			}
		}
	} else {
		for i := range n.Bindings {
			for _, id := range ids {
				if slices.Contains(n.Bindings[i].Kings, id) {
					n.Bindings[i].Supported = false
				}
			}
		}
	}
	remaining := slices.DeleteFunc(slices.Clone(loan.Cards), func(id string) bool { return seen[id] })
	if len(remaining) == 0 {
		n.Loans = append(n.Loans[:index], n.Loans[index+1:]...)
	} else {
		n.Loans[index].Cards = remaining
	}
	if e := n.Validate(); e != nil {
		return s, e
	}
	return n, nil
}

// bindingSupported derives current registered formation support from physical
// custody. Loan primitives may precede the action layer's record-controller update.
func (s State) bindingSupported(b Binding, seat int) bool {
	for _, r := range s.Formations {
		if r.ID != b.Formation {
			continue
		}
		if r.Spec.Substitute == nil || !containsSelection([][]string{r.Spec.Substitute.Kings}, b.Kings) || formationSlotKey(r.Spec.Substitute.Slot) != b.Slot || ValidateFormationShape(s, seat, r.Spec) != nil || !formationSupport(s, seat, r.Spec) {
			return false
		}
		for _, id := range r.Spec.Cards {
			c, ok := s.Card(id)
			if !ok || c.Controller != seat || c.Zone != Formation || c.Allocation != r.ID || c.AvailableFromRound > s.Round {
				return false
			}
		}
		return true
	}
	// Historical Phase 0 primitives have no formation record; preserve their
	// explicit support flag rather than interpreting an opaque ID as a kind.
	return b.Supported
}
func (s *State) refreshBindingSupport(ids []string, seat int) {
	for i, b := range s.Bindings {
		if !slices.Contains(ids, b.Kings[0]) || !slices.Contains(ids, b.Kings[1]) {
			continue
		}
		registered := false
		for _, r := range s.Formations {
			if r.ID == b.Formation {
				registered = true
				break
			}
		}
		if registered {
			s.Bindings[i].Supported = s.bindingSupported(b, seat)
			continue
		}
		support := Clubs
		if b.Formation == "people" || b.Formation == "great-people" || b.Formation == "doppelganger" {
			support = Hearts
		}
		s.Bindings[i].Supported = slices.Contains(s.CurrentSeries(seat), support)
	}
}
func (s State) AssassinateBinding(id string) (State, error) {
	c, ok := s.Card(id)
	if !ok || c.Card.Rank != 13 || (c.Zone != Formation && c.Zone != Attachment && c.Zone != Unassigned) {
		return s, errors.New("assassination requires exposed king")
	}
	n := s.Clone()
	for i, b := range n.Bindings {
		if !slices.Contains(b.Kings, id) {
			continue
		}
		for j := range n.Cards {
			card := &n.Cards[j]
			if slices.Contains(b.Kings, card.Card.ID) {
				card.Allocation = ""
				if card.Card.ID != id && (card.Zone == Formation || card.Zone == Attachment) {
					card.Zone = Unassigned
				}
			}
		}
		n.Bindings = append(n.Bindings[:i], n.Bindings[i+1:]...)
		break
	}
	return n.Move([]string{id}, 0, Discard, true)
}

// DepartPlayer applies only card/effect cleanup; financial departure is separate.
func (s State) DepartPlayer(seat int) (State, error) {
	if seat < 1 || seat > len(s.Players) {
		return s, errors.New("invalid departing seat")
	}
	n := s.Clone()
	n.Players[seat-1].Departed = true
	effects := []AceEffect{}
	for _, e := range n.Effects {
		if e.Target == seat {
			for i := range n.Cards {
				if n.Cards[i].Card.ID == e.CardID {
					n.Cards[i].Zone = Discard
					n.Cards[i].Controller = 0
				}
			}
		} else {
			effects = append(effects, e)
		}
	}
	n.Effects = effects
	n.DrawQueue = slices.DeleteFunc(n.DrawQueue, func(e DrawEntitlement) bool { return e.Seat == seat })
	ids := []string{}
	for _, c := range n.Cards {
		if c.Controller == seat {
			ids = append(ids, c.Card.ID)
		}
	}
	var e error
	n, e = n.Move(ids, 0, Draw, true)
	if e != nil {
		return s, e
	}
	return n, nil
}
func (s State) AcquireCaptured(ids []string, seat int) (State, error) {
	n := s.Clone()
	seen := map[string]bool{}
	for _, id := range ids {
		c, ok := n.Card(id)
		if !ok || seen[id] || c.Zone == ActiveAce {
			return s, errors.New("invalid capture")
		}
		seen[id] = true
		zone := Hand
		if c.Card.Rank == 1 {
			zone = ConcealedAce
		} else if c.Card.Suit == Diamonds && c.Card.Rank <= 10 {
			zone = Series
		}
		var e error
		n, e = n.Move([]string{id}, seat, zone, true)
		if e != nil {
			return s, e
		}
	}
	return n, nil
}
