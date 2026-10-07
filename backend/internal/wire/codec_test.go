package wire

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"os"
	"reflect"
	"testing"
)

func fixture(t *testing.T) (*Codec, []byte) {
	t.Helper()
	c, _ := New(bytes.Repeat([]byte{7}, 32))
	s, _ := game.NewState(3, "game")
	s, _ = s.Move([]string{"deck-1-spades-10"}, 1, game.Hand, true)
	m, _ := game.NewMatchLifecycle(s, 3)
	e, _ := matchstore.NewEnvelope(m, "rules")
	p, _ := matchstore.SeatProjection(e, 1)
	b, _ := json.Marshal(p)
	return c, b
}
func TestWireOpaqueAuthorizedRoundtrip(t *testing.T) {
	c, raw := fixture(t)
	out, err := c.EncodeProjection("match", "actor", raw)
	if err != nil || bytes.Contains(out, []byte("deck-1")) {
		t.Fatal("raw physical identity", err)
	}
	var p matchstore.Projection
	_ = json.Unmarshal(out, &p)
	h := p.Board.Cards[0].Card.ID
	body, _ := json.Marshal(map[string]any{"command_id": "buy", "game_id": "game", "expected_version": 0, "type": "purchase", "payload": map[string]any{"quantity": 2, "payment": []string{h}}})
	in, err := c.DecodeIntent("match", "actor", body, raw)
	if err != nil || in.Request.Command.Cards[0] != "deck-1-spades-10" || in.Request.Command.Actor != 1 {
		t.Fatal("authorized roundtrip", err)
	}
	if _, err = c.DecodeIntent("match", "other", body, raw); err == nil {
		t.Fatal("crossactor handle")
	}
	stub, err := c.DecodeIntent("match", "actor", body, nil)
	if err != nil || stub.Request.Command != nil || !bytes.Equal(stub.Transport, body) {
		t.Fatal("receipt-first stub", err)
	}
}

func TestWireOwnBindingReusesOnlyOwnCurrentCapabilities(t *testing.T) {
	c, raw := fixture(t)
	var p matchstore.Projection
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	p.Board.Cards[0].Card.ID = "deck-1-spades-13"
	p.Board.Cards[0].Card.Rank = 13
	p.Board.OwnBindings = []game.FormationSubstitution{{Kings: []string{p.Board.Cards[0].Card.ID}, Slot: game.FormationSlot{Rank: 12, Suit: game.Hearts}}}
	raw, _ = json.Marshal(p)
	out, err := c.EncodeProjection("match", "actor", raw)
	if err != nil || bytes.Contains(out, []byte("deck-")) {
		t.Fatalf("binding leaked raw identity: %s, %v", out, err)
	}
	var encoded matchstore.Projection
	if err := json.Unmarshal(out, &encoded); err != nil {
		t.Fatal(err)
	}
	if len(encoded.Board.OwnBindings) != 1 || len(encoded.Board.OwnBindings[0].Kings) != 1 || encoded.Board.OwnBindings[0].Kings[0] != encoded.Board.Cards[0].Card.ID {
		t.Fatalf("binding did not reuse current capability: %#v", encoded.Board.OwnBindings)
	}
	for _, foreign := range []bool{false, true} {
		if foreign {
			p.Board.Cards[0].Controller = 2
		} else {
			p.Board.OwnBindings[0].Kings = []string{"deck-2-hearts-13"}
		}
		raw, _ = json.Marshal(p)
		if _, err := c.EncodeProjection("match", "actor", raw); err == nil {
			t.Fatal("binding issued unauthorized/foreign capability")
		}
		p.Board.OwnBindings[0].Kings = []string{p.Board.Cards[0].Card.ID}
	}
}
func TestWireRejectsForgedFieldsAndRawIDs(t *testing.T) {
	c, raw := fixture(t)
	for _, body := range []string{`{"command_id":"x","game_id":"game","expected_version":0,"type":"attack","payload":{"actor":2}}`, `{"command_id":"x","command_id":"y","game_id":"game","expected_version":0,"type":"attack","payload":{}}`, `{"command_id":"x","game_id":"game","expected_version":0,"type":"attack","payload":{"selection":["deck-1-spades-10"]}}`, `{"command_id":"x","game_id":"game","expected_version":0,"type":"decision","payload":{"random_words":["0"]}}`, `{"command_id":"x","game_id":"game","expected_version":0,"type":"promise-offer","payload":{"award_id":"a","recipient":2,"mode":"fixed"}}`} {
		if _, err := c.DecodeIntent("match", "actor", []byte(body), raw); err == nil {
			t.Fatal("forgery accepted", body)
		}
	}
}
func TestWireHistoricalAndStaleCapabilitiesDoNotAuthorize(t *testing.T) {
	c, raw := fixture(t)
	var original matchstore.Projection
	_ = json.Unmarshal(raw, &original)
	original.Board.PublicHistory = []string{"deck-1-spades-10"}
	raw, _ = json.Marshal(original)
	out, _ := c.EncodeProjection("match", "actor", raw)
	var encoded matchstore.Projection
	_ = json.Unmarshal(out, &encoded)
	makeBody := func(h string) []byte {
		b, _ := json.Marshal(map[string]any{"command_id": "x", "game_id": "game", "expected_version": 0, "type": "attach", "payload": map[string]any{"selection": []string{h}}})
		return b
	}
	if _, err := c.DecodeIntent("match", "actor", makeBody(encoded.Board.PublicHistory[0]), raw); err == nil {
		t.Fatal("history authorized live selection")
	}
	original.Board.Version++
	fresh, _ := json.Marshal(original)
	if _, err := c.DecodeIntent("match", "actor", makeBody(encoded.Board.Cards[0].Card.ID), fresh); err == nil {
		t.Fatal("staleview handle accepted")
	}
	for _, body := range []string{`{"command_id":"x","game_id":"game","expected_version":0,"type":"end-turn","payload":{"selection":[]}}`, `{"command_id":"x","game_id":"game","expected_version":0,"type":"attack","payload":{"amount":{"numerator":"1","denominator":"1"}}}`} {
		if _, err := c.DecodeIntent("match", "actor", []byte(body), raw); err == nil {
			t.Fatal("foreign action fields accepted")
		}
	}
}
func TestWireRequiresOuterKeysAndObjectPayload(t *testing.T) {
	c, raw := fixture(t)
	for _, body := range []string{`{"command_id":"x","game_id":"game","type":"end-turn","payload":{}}`, `{"command_id":"x","game_id":"game","expected_version":0,"type":"end-turn","payload":null}`, `{"command_id":"x","game_id":"game","expected_version":null,"type":"end-turn","payload":{}}`} {
		for _, view := range [][]byte{nil, raw} {
			if _, err := c.DecodeIntent("match", "actor", []byte(body), view); err == nil {
				t.Fatal("required outer field/object absent", body)
			}
		}
	}
}

type styleProjection struct {
	Board struct {
		Players []struct {
			Seat      int    `json:"seat"`
			CardStyle string `json:"card_style"`
		} `json:"players"`
		Cards []game.PlacedCard `json:"cards"`
	} `json:"board"`
}

func styleEnvelope(t *testing.T, players int) matchstore.Envelope {
	t.Helper()
	b, err := game.NewState(players, "style-game")
	if err != nil {
		t.Fatal(err)
	}
	for seat := 1; seat <= players; seat++ {
		b.Order = append(b.Order, seat)
	}
	for seat := 1; seat <= players; seat++ {
		b.Players[seat-1].History = []game.Suit{game.Clubs, game.Diamonds}
		b, err = b.Move([]string{fmt.Sprintf("deck-1-hearts-%02d", seat+2)}, seat, game.Hand, true)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, move := range []struct {
		ids  []string
		seat int
	}{{[]string{"deck-1-clubs-08", "deck-1-diamonds-02"}, 1}, {[]string{"deck-1-diamonds-08"}, 2}} {
		b, err = b.Move(move.ids, move.seat, game.Series, true)
		if err != nil {
			t.Fatal(err)
		}
	}
	m, err := game.NewMatchLifecycle(b, 3)
	if err != nil {
		t.Fatal(err)
	}
	e, err := matchstore.NewEnvelope(m, "style-rules")
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func readStyles(t *testing.T, codec *Codec, match string, envelope matchstore.Envelope, viewer int) (map[int]string, styleProjection) {
	t.Helper()
	p, err := matchstore.SeatProjection(envelope, viewer)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("card_style")) {
		t.Fatal("cosmetic styles changed the persisted engine projection")
	}
	encoded, err := codec.EncodeProjection(match, fmt.Sprintf("viewer-%d", viewer), raw)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte("deck-")) || bytes.Contains(encoded, []byte("chance")) {
		t.Fatal("cosmetic assignment exposed private physical identity or chance")
	}
	var out styleProjection
	if err = json.Unmarshal(encoded, &out); err != nil {
		t.Fatal(err)
	}
	styles := make(map[int]string)
	for _, player := range out.Board.Players {
		styles[player.Seat] = player.CardStyle
	}
	if len(out.Board.Cards) != len(p.Board.Cards) {
		t.Fatal("cosmetic assignment changed authorized card visibility")
	}
	for _, card := range out.Board.Cards {
		if (card.Zone == game.Hand || card.Zone == game.ConcealedAce) && card.Controller != viewer {
			t.Fatal("another seat's concealed card was exposed")
		}
	}
	return styles, out
}

func TestWireCardStylesDistinctAndSharedForThreeAndFourSeats(t *testing.T) {
	codec, _ := New(bytes.Repeat([]byte{7}, 32))
	allowed := map[string]bool{"first-light": true, "rain-glaze": true, "ember-glaze": true, "drypoint": true}
	for _, players := range []int{3, 4} {
		t.Run(fmt.Sprint(players), func(t *testing.T) {
			e := styleEnvelope(t, players)
			var shared map[int]string
			for viewer := 1; viewer <= players; viewer++ {
				styles, _ := readStyles(t, codec, "public-match", e, viewer)
				seen := map[string]bool{}
				for seat := 1; seat <= players; seat++ {
					if !allowed[styles[seat]] || seen[styles[seat]] {
						t.Fatalf("seat %d must have a distinct accepted edition: %v", seat, styles)
					}
					seen[styles[seat]] = true
				}
				if shared != nil && !reflect.DeepEqual(shared, styles) {
					t.Fatal("viewers disagree on public seat styles")
				}
				shared = styles
			}
			// Pin the public v1 assignment so later code cannot silently restyle
			// an existing match after deployment or replay.
			want := []string{"ember-glaze", "rain-glaze", "drypoint", "first-light"}
			for seat := 1; seat <= players; seat++ {
				if shared[seat] != want[seat-1] {
					t.Fatal("public match style v1 assignment changed", shared)
				}
			}
		})
	}
}

func TestWireCardStylesSurviveCheckpointAndLaterGamesWithoutPrivateEntropy(t *testing.T) {
	codec, _ := New(bytes.Repeat([]byte{7}, 32))
	e := styleEnvelope(t, 4)
	before, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := readStyles(t, codec, "persistent-match", e, 1)
	var restored matchstore.Envelope
	if err = json.Unmarshal(before, &restored); err != nil {
		t.Fatal(err)
	}
	restored.Chance["private-outcome"] = []string{"secret-draw-order"}
	// Restore changes neither the durable envelope nor its gameplay digest.
	rotatedCodec, _ := New(bytes.Repeat([]byte{9}, 32))
	for viewer := 1; viewer <= 4; viewer++ {
		got, _ := readStyles(t, rotatedCodec, "persistent-match", restored, viewer)
		if !reflect.DeepEqual(want, got) {
			t.Fatal("reload/reconnect or handle-key rotation reassigned styles")
		}
	}
	after, _ := json.Marshal(e)
	if !bytes.Equal(before, after) {
		t.Fatal("view generation mutated gameplay checkpoint")
	}
	p, _ := matchstore.SeatProjection(restored, 1)
	p.Board.GameID = "later-game-in-the-same-match"
	p.Board.Version += 99
	p.Board.Round = 8
	raw, _ := json.Marshal(p)
	out, err := rotatedCodec.EncodeProjection("persistent-match", "viewer-1", raw)
	if err != nil {
		t.Fatal(err)
	}
	var next styleProjection
	if err = json.Unmarshal(out, &next); err != nil {
		t.Fatal(err)
	}
	for _, player := range next.Board.Players {
		if player.CardStyle != want[player.Seat] {
			t.Fatal("a later game changed match-long styles")
		}
	}
	assignments := map[string]bool{}
	for i := 0; i < 32; i++ {
		styles, _ := readStyles(t, codec, fmt.Sprintf("new-match-%d", i), e, 1)
		encoded, _ := json.Marshal(styles)
		assignments[string(encoded)] = true
	}
	if len(assignments) < 2 {
		t.Fatal("new matches never receive fresh assignments")
	}
}

func TestWireCardStylesFollowCurrentControllerAfterLegalDiamondCapture(t *testing.T) {
	codec, _ := New(bytes.Repeat([]byte{7}, 32))
	for _, players := range []int{3, 4} {
		e := styleEnvelope(t, players)
		styles, _ := readStyles(t, codec, "capture-match", e, 1)
		if styles[1] == styles[2] {
			t.Fatal("captor and former controller require distinct styles")
		}
		b := e.Match.Game.Board
		var err error
		b, _, err = game.Apply(b, game.Command{GameID: b.GameID, ID: "capture", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-diamonds-08"}})
		if err != nil {
			t.Fatal(err)
		}
		for seat := 2; seat <= players; seat++ {
			b, _, err = game.Apply(b, game.Command{GameID: b.GameID, ID: fmt.Sprintf("pass-%d", seat), Kind: "pass", Actor: seat, WindowID: b.Pending.ID})
			if err != nil {
				t.Fatal(err)
			}
		}
		captured, _ := b.Card("deck-1-diamonds-08")
		if captured.Controller != 1 || captured.Zone != game.Series {
			t.Fatal("fixture did not perform a public Diamond capture")
		}
		e.Match.Game.Board = b
		for viewer := 1; viewer <= players; viewer++ {
			current, view := readStyles(t, codec, "capture-match", e, viewer)
			if !reflect.DeepEqual(styles, current) {
				t.Fatal("capture reassigned a seat's style")
			}
			found := false
			for _, card := range view.Board.Cards {
				if card.Card.Suit == game.Diamonds && card.Card.Rank == 8 {
					found = true
					if card.Controller != 1 || current[card.Controller] != styles[1] {
						t.Fatal("captured card did not resolve to new controller's style")
					}
				}
			}
			if !found {
				t.Fatal("captured public Diamond missing")
			}
		}
	}
}

func TestWirePendingCombatUsesOnlyCurrentPublicCardHandles(t *testing.T) {
	codec, _ := New(bytes.Repeat([]byte{7}, 32))
	e := styleEnvelope(t, 4)
	b := e.Match.Game.Board
	b, _, err := game.Apply(b, game.Command{GameID: b.GameID, ID: "public-attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-diamonds-08"}})
	if err != nil {
		t.Fatal(err)
	}
	e.Match.Game.Board = b
	for viewer := 1; viewer <= 4; viewer++ {
		t.Run(fmt.Sprintf("viewer-%d", viewer), func(t *testing.T) {
			p, err := matchstore.SeatProjection(e, viewer)
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(p)
			actor := fmt.Sprintf("viewer-%d", viewer)
			encoded, err := codec.EncodeProjection("combat-match", actor, raw)
			if err != nil || bytes.Contains(encoded, []byte("deck-")) {
				t.Fatal("combat leaked raw physical identities", err)
			}
			var view struct {
				Board  game.Observation `json:"board"`
				Online struct {
					RulesContext struct {
						Combat struct {
							AttackCards []string `json:"attack_cards"`
							TargetCards []string `json:"target_cards"`
						} `json:"combat"`
					} `json:"rules_context"`
				} `json:"online"`
			}
			if err = json.Unmarshal(encoded, &view); err != nil {
				t.Fatal(err)
			}
			public := map[string]bool{}
			var attack, target string
			for _, card := range view.Board.Cards {
				if card.Zone == game.Series {
					public[card.Card.ID] = true
					if card.Controller == 1 && card.Card.Suit == game.Clubs && card.Card.Rank == 8 {
						attack = card.Card.ID
					}
					if card.Controller == 2 && card.Card.Suit == game.Diamonds && card.Card.Rank == 8 {
						target = card.Card.ID
					}
				}
			}
			combat := view.Online.RulesContext.Combat
			if attack == "" || target == "" || !reflect.DeepEqual(combat.AttackCards, []string{attack}) || !reflect.DeepEqual(combat.TargetCards, []string{target}) {
				t.Fatalf("combat operands do not match current authorized board handles: %+v", combat)
			}
			for _, handle := range append(append([]string{}, combat.AttackCards...), combat.TargetCards...) {
				if !public[handle] {
					t.Fatal("hidden or unrelated card entered public combat")
				}
			}
			body, _ := json.Marshal(map[string]any{"command_id": "context-selection", "game_id": b.GameID, "expected_version": b.Version, "type": "attack", "payload": map[string]any{"selection": combat.AttackCards, "targets": combat.TargetCards}})
			intent, err := codec.DecodeIntent("combat-match", actor, body, raw)
			if err != nil || !reflect.DeepEqual(intent.Request.Command.Cards, b.Pending.Clubs) || !reflect.DeepEqual(intent.Request.Command.Targets, b.Pending.Targets) {
				t.Fatal("combat handles do not resolve through normal authority admission", err)
			}
			p.Board.Version++
			fresh, _ := json.Marshal(p)
			if _, err = codec.DecodeIntent("combat-match", actor, body, fresh); err == nil {
				t.Fatal("stale combat handles authorized current selection")
			}
			if _, err = codec.DecodeIntent("combat-match", "other-viewer", body, raw); err == nil {
				t.Fatal("another viewer reused combat handles")
			}
			// A malformed internal projection cannot mint a capability for an
			// unobserved card merely by mentioning it in public combat context.
			p.Online.RulesContext.Combat.AttackCards = []string{"deck-4-spades-13"}
			invalid, _ := json.Marshal(p)
			if _, err = codec.EncodeProjection("combat-match", actor, invalid); err == nil {
				t.Fatal("combat minted a capability absent from the authorized board")
			}
		})
	}
}

func TestWirePendingContextPreservesPublicDescriptorAndResponderScope(t *testing.T) {
	codec, _ := New(bytes.Repeat([]byte{7}, 32))
	e := styleEnvelope(t, 4)
	b := e.Match.Game.Board
	var err error
	b, err = b.Move([]string{"deck-1-spades-08"}, 2, game.Series, true)
	if err != nil {
		t.Fatal(err)
	}
	b, _, err = game.Apply(b, game.Command{GameID: b.GameID, ID: "pending-public", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-spades-08"}})
	if err != nil {
		t.Fatal(err)
	}
	e.Match.Game.Board = b
	for viewer := 1; viewer <= 4; viewer++ {
		p, err := matchstore.SeatProjection(e, viewer)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(p)
		encoded, err := codec.EncodeProjection("pending-match", fmt.Sprintf("viewer-%d", viewer), raw)
		if err != nil {
			t.Fatal(err)
		}
		var decoded matchstore.Projection
		if err = json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		pending := decoded.Online.RulesContext.Pending
		if pending == nil || pending.Type != "attack" || pending.Actor != 1 || pending.TargetSeat != 2 || !reflect.DeepEqual(pending, p.Online.RulesContext.Pending) {
			t.Fatalf("viewer %d changed public pending context: %+v", viewer, pending)
		}
		if viewer != 2 && len(pending.ResponseTypes) != 0 {
			t.Fatalf("viewer %d can respond out of order", viewer)
		}
		if viewer == 2 && len(pending.ResponseTypes) == 0 {
			t.Fatal("current responder lost public response context")
		}
		description, _ := json.Marshal(pending)
		if bytes.Contains(description, []byte("deck-")) || bytes.Contains(description, []byte("card")) || bytes.Contains(description, []byte("proposal")) {
			t.Fatalf("private details entered descriptor: %s", description)
		}
	}
}

func TestPendingContextSchemaIsAdditiveAndClosed(t *testing.T) {
	raw, err := os.ReadFile("../../../docs/api/online.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Components struct {
			Schemas map[string]struct {
				AdditionalProperties bool                       `json:"additionalProperties"`
				Properties           map[string]json.RawMessage `json:"properties"`
				Required             []string                   `json:"required"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err = json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	rules := doc.Components.Schemas["RulesContext"]
	if !bytes.Contains(rules.Properties["pending"], []byte("#/components/schemas/PendingContext")) {
		t.Fatal("missing pending schema reference")
	}
	for _, key := range rules.Required {
		if key == "pending" {
			t.Fatal("older projections must remain valid")
		}
	}
	pending, ok := doc.Components.Schemas["PendingContext"]
	if !ok || pending.AdditionalProperties || len(pending.Properties) != 4 || !reflect.DeepEqual(pending.Required, []string{"type", "actor", "response_types"}) {
		t.Fatal("pending schema must allow only public descriptor fields")
	}
	for _, key := range []string{"type", "actor", "target_seat", "response_types"} {
		if _, ok := pending.Properties[key]; !ok {
			t.Fatalf("missing %s", key)
		}
	}
}

func TestOwnBindingContextSchemaIsPrivateAdditiveContext(t *testing.T) {
	raw, err := os.ReadFile("../../../docs/api/online.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Components struct {
			Schemas map[string]struct {
				AdditionalProperties bool                       `json:"additionalProperties"`
				Properties           map[string]json.RawMessage `json:"properties"`
				Required             []string                   `json:"required"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	binding, ok := doc.Components.Schemas["OwnBindingContext"]
	if !ok || binding.AdditionalProperties || len(binding.Properties) != 2 || !reflect.DeepEqual(binding.Required, []string{"kings", "slot"}) {
		t.Fatal("binding context must contain only own kings and fixed slot")
	}
	var kings struct {
		MinItems int  `json:"minItems"`
		MaxItems int  `json:"maxItems"`
		Unique   bool `json:"uniqueItems"`
		Items    struct {
			Ref string `json:"$ref"`
		} `json:"items"`
	}
	if json.Unmarshal(binding.Properties["kings"], &kings) != nil || kings.MinItems != 1 || kings.MaxItems != 2 || !kings.Unique || kings.Items.Ref != "#/components/schemas/CardHandle" {
		t.Fatal("partial own binding must use one or two current card handles")
	}
	var board struct {
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if json.Unmarshal(doc.Components.Schemas["Projection"].Properties["board"], &board) != nil || !bytes.Contains(board.Properties["own_bindings"], []byte("#/components/schemas/OwnBindingContext")) {
		t.Fatal("board missing private binding context")
	}
	for _, field := range board.Required {
		if field == "own_bindings" {
			t.Fatal("historical projections must remain valid")
		}
	}
}

func TestWireKidnapperOutcomeUsesOnlyOfferedPublicHandles(t *testing.T) {
	codec, _ := New(bytes.Repeat([]byte{7}, 32))
	e := styleEnvelope(t, 4)
	b := e.Match.Game.Board
	var err error
	for _, move := range []struct {
		ids  []string
		seat int
		zone game.Zone
	}{
		{[]string{"deck-1-clubs-11", "deck-2-clubs-11"}, 1, game.Hand},
		{[]string{"deck-1-hearts-12", "deck-1-spades-13"}, 2, game.Unassigned},
	} {
		b, err = b.Move(move.ids, move.seat, move.zone, true)
		if err != nil {
			t.Fatal(err)
		}
	}
	b, err = game.OpenFormation(b, 1, "kid", game.FormationSpec{Kind: "kidnapper", Cards: []string{"deck-1-clubs-11", "deck-2-clubs-11"}})
	if err != nil {
		t.Fatal(err)
	}
	b, err = game.DeclareAbility(b, game.Command{ID: "kidnap", Actor: 1, Kind: "kidnapper", FormationID: "kid", TargetSeat: 2})
	if err != nil {
		t.Fatal(err)
	}
	b.Pending.Cursor = len(b.Pending.Responders)
	e.Match.Game.Board = b
	for viewer := 1; viewer <= 4; viewer++ {
		p, err := matchstore.SeatProjection(e, viewer)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(p)
		actor := fmt.Sprintf("viewer-%d", viewer)
		encoded, err := codec.EncodeProjection("kidnapper-match", actor, raw)
		if err != nil || bytes.Contains(encoded, []byte("deck-")) {
			t.Fatal("raw identity in outcome projection", err)
		}
		var decoded matchstore.Projection
		if err = json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.Board.RequiredActor != 1 || decoded.Board.DecisionKind != "kidnapper-outcome" || decoded.Board.DecisionID != "" {
			t.Fatal("wire changed outcome boundary")
		}
		if viewer != 1 {
			if len(decoded.Board.Choices) != 0 {
				t.Fatal("another viewer received the outcome choices")
			}
			continue
		}
		if len(decoded.Board.Choices) != 2 {
			t.Fatalf("missing exposed choices: %+v", decoded.Board.Choices)
		}
		for _, choice := range decoded.Board.Choices {
			visible := false
			for _, card := range decoded.Board.Cards {
				visible = visible || card.Controller == 2 && card.Zone == game.Unassigned && card.Card.ID == choice[0]
			}
			if !visible {
				t.Fatal("choice is not a current public board handle")
			}
			body, _ := json.Marshal(map[string]any{"command_id": "outcome", "game_id": b.GameID, "expected_version": b.Version, "type": "kidnapper-outcome", "payload": map[string]any{"pending_action_id": "kidnap", "selection": choice}})
			intent, err := codec.DecodeIntent("kidnapper-match", actor, body, raw)
			if err != nil {
				t.Fatal(err)
			}
			done, _, err := game.Apply(b, *intent.Request.Command)
			if err != nil || done.Pending != nil {
				t.Fatal("offered choice did not resolve through authority", err)
			}
			card, ok := done.Card(intent.Request.Command.Cards[0])
			if !ok || card.Controller != 1 || card.Zone != game.Hand {
				t.Fatal("chosen royal did not arrive in actor hand")
			}
			newer := p
			newer.Board.Version++
			fresh, _ := json.Marshal(newer)
			if _, err = codec.DecodeIntent("kidnapper-match", actor, body, fresh); err == nil {
				t.Fatal("stale outcome handle accepted")
			}
		}
	}
}
