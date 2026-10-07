package matchstore

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/metaphy6/cgms/backend/internal/game"
	"strings"
	"testing"
)

func TestOwnCorrectionProjectionAndNullification(t *testing.T) {
	l := game.NewFinancialLedger("prior", 3)
	var err error
	check := func(e error) {
		t.Helper()
		if e != nil {
			t.Fatal(e)
		}
	}
	l, err = l.Charge("original-charge", 0, 1, game.IntAmount(10))
	check(err)
	l, err = l.CloseOrdinary(nil)
	check(err)
	for i := 0; i < 3; i++ {
		l, err = l.FinishSettlement(i)
		check(err)
	}
	l, err = l.Finalize()
	check(err)
	l, err = l.NextGame("adapter-game")
	check(err)
	start := l.Clone()
	l, err = l.Forgive("consent-operation", l.Debts[0].ID, game.IntAmount(4), []int{1})
	check(err)
	e := envelopeFixture(t)
	e.Match.Game.Ledger = l
	read := func(seat int) []game.FinancialCorrection {
		t.Helper()
		b, x := json.Marshal(onlineProjection(e, seat))
		check(x)
		var out struct {
			Corrections []game.FinancialCorrection `json:"corrections"`
		}
		check(json.Unmarshal(b, &out))
		return out.Corrections
	}
	own := read(1)
	if len(own) != 1 || own[0] != l.Corrections[0] || !own[0].MatchLevel {
		t.Fatal("own adjustment lineage missing")
	}
	if len(read(2)) != 0 || len(read(3)) != 0 {
		t.Fatal("another player's correction disclosed")
	}
	if onlineProjection(e, 1).OwnMatchScore.Cmp(game.IntAmount(-6)) != 0 {
		t.Fatal("match correction total")
	}
	l, err = l.Nullify(start, []int{0, 1, 2})
	check(err)
	e.Match.Game.Ledger = l
	if len(read(1)) != 0 {
		t.Fatal("void correction remains effective")
	}
	if onlineProjection(e, 1).OwnMatchScore.Cmp(game.IntAmount(-10)) != 0 {
		t.Fatal("nullification total")
	}
}

func TestUncheckedTransportRejectedBeforeStorage(t *testing.T) {
	defer func() {
		if recover() != nil {
			t.Fatal("unchecked transport reached storage")
		}
	}()
	s := New(nil, "rules")
	_, err := s.Submit(context.Background(), "m", "a", Intent{ID: "x", GameID: "g", Transport: json.RawMessage(`{}`)})
	if !errors.Is(err, ErrInvalid) {
		t.Fatal("transport requires admission callback")
	}
}
func TestCreateTxRejectsOversizedIdentitiesBeforeStorage(t *testing.T) {
	e := envelopeFixture(t)
	s := New(nil, e.RulesHash)
	for _, actor := range []bool{false, true} {
		t.Run(map[bool]string{false: "match", true: "actor"}[actor], func(t *testing.T) {
			defer func() {
				if recover() != nil {
					t.Fatal("oversized identity reached storage")
				}
			}()
			id := "m"
			actors := []string{"a", "b", "c"}
			if actor {
				actors[0] = strings.Repeat("x", 257)
			} else {
				id = strings.Repeat("x", 257)
			}
			if !errors.Is(s.CreateTx(context.Background(), nil, id, e, actors), ErrInvalid) {
				t.Fatal("oversized identity accepted")
			}
		})
	}
}

func TestOnlinePromisePaymentPendingIndicator(t *testing.T) {
	e := envelopeFixture(t)
	e.Match.Game.PromisePayment = &game.PromisePaymentCursor{}
	if !onlineProjection(e, 1).AutomaticPending {
		t.Fatal("pending promise payment omitted")
	}
}

func TestOnlineServerPendingCoversDealInitiativeAndBeginTurn(t *testing.T) {
	e := envelopeFixture(t)
	e.Match.Game.Board.Order = nil
	seen := map[string]bool{}
	for version := int64(0); version < 32; version++ {
		intent, pending := NextServer(e, version)
		for seat := 1; seat <= 3; seat++ {
			encoded, err := json.Marshal(onlineProjection(e, seat))
			if err != nil {
				t.Fatal(err)
			}
			var projected map[string]any
			if err = json.Unmarshal(encoded, &projected); err != nil {
				t.Fatal(err)
			}
			if projected["server_pending"] != pending {
				t.Fatalf("version %d seat %d exposed provisional turn readiness: server_pending=%v want %v", version, seat, projected["server_pending"], pending)
			}
			if projected["turn_started"] != e.Match.Game.TurnStarted {
				t.Fatalf("version %d seat %d omitted authoritative scheduled-draw boundary", version, seat)
			}
		}
		if !pending {
			if !e.Match.Game.TurnStarted || !seen["deal"] || !seen["initiative"] || !seen["begin-turn"] {
				t.Fatal("ready before authoritative initialization completed", seen)
			}
			// A real human response stops automatic scheduling, even when it is
			// another seat's turn; this indicator must never auto-answer it.
			e.Match.Game.Board.Pending = &game.PendingAction{ID: "human-response", Actor: 2, Responders: []int{1}}
			encoded, err := json.Marshal(onlineProjection(e, 1))
			if err != nil {
				t.Fatal(err)
			}
			var projected map[string]any
			if err = json.Unmarshal(encoded, &projected); err != nil || projected["server_pending"] != false {
				t.Fatal("human wait incorrectly marked as automatic", err)
			}
			return
		}
		request := intent.Request
		request.ID = intent.ID
		if request.Operation != nil {
			request.Operation.ID = intent.ID
			seen[request.Operation.Kind] = true
		} else {
			seen[request.Kind] = true
		}
		var err error
		e, err = applyServerRequest(e, request)
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("initialization exceeded bounded transition budget")
}

func TestHistoricalReadinessAllowsOnlyAbsentAdditiveMetadata(t *testing.T) {
	current := []byte(`{"online":{"round_closed":false,"departure_pending":false,"automatic_pending":false,"server_pending":true,"games_per_match":3},"board":{"active":1}}`)
	for _, tc := range []struct {
		name string
		old  string
		want bool
	}{
		{"current", string(current), true},
		{"pre-readiness", `{"online":{"round_closed":false,"departure_pending":false,"automatic_pending":false,"games_per_match":3},"board":{"active":1}}`, true},
		{"pre-round-metadata", `{"online":{"automatic_pending":false,"games_per_match":3},"board":{"active":1}}`, true},
		{"wrong-readiness", `{"online":{"round_closed":false,"departure_pending":false,"automatic_pending":false,"server_pending":false,"games_per_match":3},"board":{"active":1}}`, false},
		{"unrelated-tamper", `{"online":{"round_closed":false,"departure_pending":false,"automatic_pending":false,"games_per_match":99},"board":{"active":1}}`, false},
		{"partial-round-fields", `{"online":{"round_closed":false,"automatic_pending":false,"games_per_match":3},"board":{"active":1}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sameHistoricalProjection(current, []byte(tc.old)); got != tc.want {
				t.Fatalf("compatibility=%v want %v", got, tc.want)
			}
		})
	}
}

func TestHistoricalTurnStartedAllowsOnlyAbsentAdditiveField(t *testing.T) {
	current := []byte(`{"online":{"round_closed":false,"departure_pending":false,"turn_started":true},"board":{"active":1}}`)
	for _, tc := range []struct {
		old  string
		want bool
	}{
		{string(current), true},
		{`{"online":{"round_closed":false,"departure_pending":false},"board":{"active":1}}`, true},
		{`{"online":{"round_closed":false,"departure_pending":false,"turn_started":false},"board":{"active":1}}`, false},
		{`{"online":{"round_closed":false,"departure_pending":false,"turn_started":null},"board":{"active":1}}`, false},
		{`{"online":{"round_closed":false,"departure_pending":false,"turn_started":false,"turn_started":true},"board":{"active":1}}`, false},
		{`{"online":{"round_closed":false,"departure_pending":false},"board":{"active":2}}`, false},
	} {
		if got := sameHistoricalProjection(current, []byte(tc.old)); got != tc.want {
			t.Fatalf("compatibility=%v want %v for %s", got, tc.want, tc.old)
		}
	}
}

func TestHistoricalOwnBindingsAllowsOnlyAbsentAdditiveField(t *testing.T) {
	current := `{"online":{},"board":{"active":1,"own_bindings":[{"kings":["king"],"slot":{"rank":12,"suit":"hearts"}}]}}`
	for _, tc := range []struct {
		old  string
		want bool
	}{
		{current, true},
		{`{"online":{},"board":{"active":1}}`, true},
		{strings.Replace(current, `"rank":12`, `"rank":11`, 1), false},
		{strings.Replace(current, `"king"`, `"different-king"`, 1), false},
		{`{"online":{},"board":{"active":1,"own_bindings":null}}`, false},
		{`{"online":{},"board":{"active":1,"own_bindings":[]}}`, false},
		{strings.Replace(current, `"own_bindings":`, `"own_bindings":[],"own_bindings":`, 1), false},
		{`{"online":{},"board":{"active":2}}`, false},
	} {
		if got := sameHistoricalProjection([]byte(current), []byte(tc.old)); got != tc.want {
			t.Fatalf("compatibility=%v want %v for %s", got, tc.want, tc.old)
		}
	}
}

func TestHistoricalCardContextAllowsOnlyAbsentAdditiveFields(t *testing.T) {
	current := `{"online":{"rules_context":{"opening_used":false,"pending":{"type":"attack","actor":1,"target_seat":2,"response_types":[]},"combat":{"actor":1,"attack_cards":["club"],"target_cards":["heart"],"attack":"8","defense":"8"}}},"board":{"version":7,"required_actor":2,"window_id":"attack"}}`
	withoutPending := strings.Replace(current, `"pending":{"type":"attack","actor":1,"target_seat":2,"response_types":[]},`, "", 1)
	withoutCards := strings.Replace(current, `"attack_cards":["club"],"target_cards":["heart"],`, "", 1)
	legacy := strings.Replace(withoutPending, `"attack_cards":["club"],"target_cards":["heart"],`, "", 1)
	for _, tc := range []struct {
		name, historical string
		want             bool
	}{
		{"current", current, true},
		{"missing-pending", withoutPending, true},
		{"missing-combat-operands", withoutCards, true},
		{"legacy-pending-combat", legacy, true},
		{"changed-pending", strings.Replace(current, `"target_seat":2`, `"target_seat":3`, 1), false},
		{"null-pending", strings.Replace(current, `{"type":"attack","actor":1,"target_seat":2,"response_types":[]}`, `null`, 1), false},
		{"changed-operand", strings.Replace(current, `["club"]`, `["other"]`, 1), false},
		{"partial-operand-pair", strings.Replace(current, `"attack_cards":["club"],`, "", 1), false},
		{"changed-existing-total", strings.Replace(legacy, `"attack":"8"`, `"attack":"9"`, 1), false},
		{"changed-existing-marker", strings.Replace(legacy, `"opening_used":false`, `"opening_used":true`, 1), false},
		{"duplicate-existing-marker", strings.Replace(legacy, `"opening_used":false`, `"opening_used":true,"opening_used":false`, 1), false},
		{"changed-window", strings.Replace(legacy, `"window_id":"attack"`, `"window_id":"other"`, 1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			now, old := []byte(current), []byte(tc.historical)
			if got := sameHistoricalProjection(now, old); got != tc.want {
				t.Fatalf("compatibility=%v want %v", got, tc.want)
			}
			if string(now) != current || string(old) != tc.historical {
				t.Fatal("historical comparison rewrote an input")
			}
		})
	}
}

func TestHistoricalKidnapperAllowsOnlyPreviousUnpromptedBoundary(t *testing.T) {
	current := `{"online":{"rules_context":{"pending":{"type":"kidnapper","actor":1,"target_seat":2,"response_types":[]}}},"board":{"version":7,"required_actor":1,"window_id":"kidnap","decision_kind":"kidnapper-outcome","choices":[["royal"]]}}`
	legacy := `{"online":{"rules_context":{}},"board":{"version":7,"required_actor":0,"window_id":"kidnap"}}`
	for _, tc := range []struct {
		name, historical string
		want             bool
	}{
		{"current", current, true},
		{"legacy-outcome", legacy, true},
		{"wrong-old-actor", strings.Replace(legacy, `"required_actor":0`, `"required_actor":2`, 1), false},
		{"missing-old-actor", strings.Replace(legacy, `"required_actor":0,`, "", 1), false},
		{"changed-present-kind", strings.Replace(current, `"kidnapper-outcome"`, `"justice"`, 1), false},
		{"partial-kind-without-choices", strings.Replace(current, `,"choices":[["royal"]]`, "", 1), false},
		{"old-choices-without-kind", strings.Replace(legacy, `"window_id":"kidnap"`, `"window_id":"kidnap","choices":[["royal"]]`, 1), false},
		{"old-decision-id", strings.Replace(legacy, `"window_id":"kidnap"`, `"window_id":"kidnap","decision_id":"decision"`, 1), false},
		{"changed-present-choice", strings.Replace(current, `"royal"`, `"other"`, 1), false},
		{"changed-present-actor", strings.Replace(current, `"required_actor":1`, `"required_actor":0`, 1), false},
		{"duplicate-old-actor", strings.Replace(legacy, `"required_actor":0`, `"required_actor":2,"required_actor":0`, 1), false},
		{"changed-board-version", strings.Replace(legacy, `"version":7`, `"version":6`, 1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			now, old := []byte(current), []byte(tc.historical)
			if got := sameHistoricalProjection(now, old); got != tc.want {
				t.Fatalf("compatibility=%v want %v", got, tc.want)
			}
			if string(now) != current || string(old) != tc.historical {
				t.Fatal("historical comparison rewrote an input")
			}
		})
	}
	// Nonacting observers gain only the boundary metadata, never the choices.
	nonactor := strings.Replace(current, `,"choices":[["royal"]]`, "", 1)
	if !sameHistoricalProjection([]byte(nonactor), []byte(legacy)) {
		t.Fatal("older nonacting observer receipt rejected")
	}
	otherDecision := strings.Replace(current, `"kidnapper-outcome"`, `"justice"`, 1)
	if sameHistoricalProjection([]byte(otherDecision), []byte(legacy)) {
		t.Fatal("compatibility stripped an unrelated decision")
	}
}

func TestOnlineDeparturePromptIsOwnAnswerOnly(t *testing.T) {
	e := envelopeFixture(t)
	e.Match.Game.Board.Active = 3
	e.Match.Game.TurnIndex = 2
	e.Match.Game.RoundClosed = true
	e.Match.Game.DepartureChoices = []game.DepartureChoice{{Seat: 2, Leave: true}}
	read := func(seat int) map[string]any {
		t.Helper()
		p, err := SeatProjection(e, seat)
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(p.Online)
		if err != nil {
			t.Fatal(err)
		}
		var out map[string]any
		if err = json.Unmarshal(data, &out); err != nil {
			t.Fatal(err)
		}
		if _, exists := out["departure_choices"]; exists {
			t.Fatal("other players' choices disclosed")
		}
		return out
	}
	for _, seat := range []int{1, 2, 3} {
		v := read(seat)
		if v["round_closed"] != true || v["departure_pending"] != (seat != 2) {
			t.Fatal("missing own round choice prompt", seat, v["round_closed"], v["departure_pending"])
		}
	}
	before, err := json.Marshal(read(1))
	if err != nil {
		t.Fatal(err)
	}
	e.Match.Game.DepartureChoices[0].Leave = false
	after, err := json.Marshal(read(1))
	if err != nil || string(before) != string(after) {
		t.Fatal("another player's secret stay/leave answer changed own projection")
	}
	e.Match.Game.DeparturesResolved = true
	if read(1)["departure_pending"] != false {
		t.Fatal("resolved choice remains actionable")
	}
	e.Match.Game.DeparturesResolved = false
	e.Match.Game.Board.Round = 13
	if read(1)["departure_pending"] != false {
		t.Fatal("last-round closure asks for unavailable departure")
	}
}
