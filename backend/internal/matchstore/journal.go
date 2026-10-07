package matchstore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"
)

var ErrJournal = errors.New("private journal verification failed")

// VerifyJournal is an administrative integrity check, never a member endpoint.
// It checks deterministic transitions using recorded random outcomes; it does
// not authenticate a random seed or prove that a recorded draw was unbiased.
func (s *Store) VerifyJournal(ctx context.Context, id string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)
	final, version, err := s.locked(ctx, tx, id)
	if err != nil {
		return err
	}
	var configBytes []byte
	if err = tx.QueryRow(ctx, `SELECT config FROM matches WHERE match_id=$1`, id).Scan(&configBytes); err != nil {
		return err
	}
	var config persistedConfig
	if len(configBytes) > 64<<20 || json.Unmarshal(configBytes, &config) != nil || config.RulesHash != s.rulesHash || config.GameLimit != final.Match.GameLimit || config.Initial.Match.GameLimit != config.GameLimit || config.Initial.RulesHash != config.RulesHash || config.Initial.Validate() != nil {
		return ErrJournal
	}
	current := config.Initial.Clone()
	var count, max int64
	if err = tx.QueryRow(ctx, `SELECT count(*),coalesce(max(seq),0) FROM match_events WHERE match_id=$1`, id).Scan(&count, &max); err != nil {
		return err
	}
	if count != version || max != version {
		return ErrJournal
	}
	for seq := int64(1); seq <= version; seq++ {
		var raw []byte
		var eventGame string
		if err = tx.QueryRow(ctx, `SELECT game_id,event FROM match_events WHERE match_id=$1 AND seq=$2`, id, seq).Scan(&eventGame, &raw); err != nil {
			return ErrJournal
		}
		var j JournalEntry
		if len(raw) > 64<<20 || json.Unmarshal(raw, &j) != nil || j.Version != seq || j.CommandID == "" || j.Snapshot.Validate() != nil || j.Snapshot.RulesHash != config.RulesHash || j.Snapshot.Match.GameLimit != config.GameLimit || eventGame != j.Snapshot.Match.Game.Board.GameID {
			return ErrJournal
		}
		// Never permit replay to replace a previously committed random outcome.
		for key, prior := range current.Chance {
			after, ok := j.Snapshot.Chance[key]
			if !ok || !slices.Equal(prior, after) {
				return ErrJournal
			}
		}
		replay := current.Clone()
		replay.Chance = j.Snapshot.Clone().Chance
		tracker := &chanceReplay{used: map[string]bool{}}
		replay.replay = tracker
		var next Envelope
		var hash []byte
		seat := 0
		if j.Server {
			if j.ActorID != "" {
				return ErrJournal
			}
			var in ServerIntent
			if json.Unmarshal(j.Input, &in) != nil || in.ID != j.CommandID || in.GameID != current.Match.Game.Board.GameID || in.ExpectedVersion != seq-1 || in.Request.GameID != in.GameID {
				return ErrJournal
			}
			hash, err = bodyHash(in)
			if err != nil {
				return ErrJournal
			}
			req := in.Request
			req.ID = scopedID("scheduler:", in.ID)
			if req.Operation != nil {
				op := *req.Operation
				op.ID = req.ID
				req.Operation = &op
			}
			next, err = applyServerRequest(replay, req)
		} else {
			if j.ActorID == "" {
				return ErrJournal
			}
			seat, err = member(ctx, tx, id, j.ActorID)
			if err != nil {
				return ErrJournal
			}
			var in Intent
			if json.Unmarshal(j.Input, &in) != nil || in.ID != j.CommandID || in.GameID != current.Match.Game.Board.GameID || in.ExpectedVersion < 0 {
				return ErrJournal
			}
			victory := (in.Request.Command != nil && in.Request.Command.Kind == "coup") || (in.Request.Operation != nil && (in.Request.Operation.Kind == "coup" || in.Request.Operation.Kind == "declare-ordinary"))
			if !victory && in.ExpectedVersion != seq-1 {
				return ErrJournal
			}
			hash, err = bodyHash(in)
			if err != nil {
				return ErrJournal
			}
			req := in.Request
			if req.Command != nil {
				c := *req.Command
				if c.GameID != in.GameID || c.ID != in.ID {
					return ErrJournal
				}
				c.ID = scopedID("actor:"+j.ActorID, in.ID)
				req.Command = &c
			}
			if req.Operation != nil {
				op := *req.Operation
				if op.GameID != in.GameID || op.ID != in.ID {
					return ErrJournal
				}
				op.ID = scopedID("actor:"+j.ActorID, in.ID)
				req.Operation = &op
			}
			next, err = applyRequest(replay, seat, req)
		}
		if err != nil {
			return ErrJournal
		}
		for key := range j.Snapshot.Chance {
			if _, old := current.Chance[key]; !old && !tracker.used[key] {
				return ErrJournal
			}
		}
		next.replay = nil
		if !sameEnvelope(next, j.Snapshot) {
			return ErrJournal
		}
		result, found, err := receipt(ctx, tx, id, j.ActorID, j.CommandID, hash, j.Server)
		if err != nil || !found || result.Version != seq || result.GameID != eventGame {
			return ErrJournal
		}
		status := "committed"
		if next.Match.Game.Automatic != nil || next.Match.Game.PromisePayment != nil {
			status = "committed-pending-settlement"
		}
		if result.Status != status {
			return ErrJournal
		}
		if j.Server {
			if len(result.Projection) != 0 {
				return ErrJournal
			}
		} else {
			projection, err := project(next, seat)
			if err != nil || !sameHistoricalProjection(projection, result.Projection) {
				return ErrJournal
			}
		}
		current = next
	}
	if !sameEnvelope(current, final) {
		return ErrJournal
	}
	return nil
}

// Older immutable receipts predate additive public context and prompt fields.
// Remove only those absent fields from a comparison copy of the replayed view;
// present values and all unrelated fields must still match. Never rewrite a
// historical receipt/event or relax the authoritative envelope comparison.
func sameHistoricalProjection(current, historical []byte) bool {
	if bytes.Equal(current, historical) {
		return true
	}
	object := func(raw []byte) map[string]json.RawMessage {
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil {
			return nil
		}
		// Map decoding keeps the last duplicate key. Reject duplicate members
		// before normalization so a changed earlier value cannot disappear.
		decoder := json.NewDecoder(bytes.NewReader(raw))
		if _, err := decoder.Token(); err != nil {
			return nil
		}
		members := 0
		for decoder.More() {
			if _, err := decoder.Token(); err != nil {
				return nil
			}
			var value json.RawMessage
			if decoder.Decode(&value) != nil {
				return nil
			}
			members++
		}
		if members != len(fields) {
			return nil
		}
		return fields
	}
	encoded := func(fields map[string]json.RawMessage) json.RawMessage {
		// Every value came from a successfully decoded JSON object above.
		raw, _ := json.Marshal(fields)
		return raw
	}
	now, old := object(current), object(historical)
	nowOnline, oldOnline := object(now["online"]), object(old["online"])
	if now == nil || old == nil || nowOnline == nil || oldOnline == nil {
		return false
	}
	_, hasRound := oldOnline["round_closed"]
	_, hasDeparture := oldOnline["departure_pending"]
	if hasRound != hasDeparture {
		return false
	}
	if !hasRound {
		delete(nowOnline, "round_closed")
		delete(nowOnline, "departure_pending")
	}
	if _, present := oldOnline["server_pending"]; !present {
		delete(nowOnline, "server_pending")
	}
	if _, present := oldOnline["turn_started"]; !present {
		delete(nowOnline, "turn_started")
	}
	if nowRules, oldRules := object(nowOnline["rules_context"]), object(oldOnline["rules_context"]); nowRules != nil && oldRules != nil {
		if _, present := oldRules["pending"]; !present {
			delete(nowRules, "pending")
		}
		if nowCombat, oldCombat := object(nowRules["combat"]), object(oldRules["combat"]); nowCombat != nil && oldCombat != nil {
			_, hasAttack := oldCombat["attack_cards"]
			_, hasTarget := oldCombat["target_cards"]
			if hasAttack != hasTarget {
				return false
			}
			if !hasAttack {
				delete(nowCombat, "attack_cards")
				delete(nowCombat, "target_cards")
			}
			nowRules["combat"], oldRules["combat"] = encoded(nowCombat), encoded(oldCombat)
		}
		nowOnline["rules_context"], oldOnline["rules_context"] = encoded(nowRules), encoded(oldRules)
	}
	now["online"], old["online"] = encoded(nowOnline), encoded(oldOnline)
	if nowBoard, oldBoard := object(now["board"]), object(old["board"]); nowBoard != nil && oldBoard != nil {
		if _, present := oldBoard["own_bindings"]; !present {
			delete(nowBoard, "own_bindings")
		}
		if bytes.Equal(nowBoard["decision_kind"], []byte(`"kidnapper-outcome"`)) {
			if _, hasKind := oldBoard["decision_kind"]; !hasKind {
				// Before this prompt existed, the response-complete Kidnapper
				// boundary had required_actor=0 and no decision metadata. No
				// other actor, choice, or decision value is a compatible omission.
				_, hasChoices := oldBoard["choices"]
				_, hasDecision := oldBoard["decision_id"]
				_, currentDecision := nowBoard["decision_id"]
				if !bytes.Equal(oldBoard["required_actor"], []byte("0")) || hasChoices || hasDecision || currentDecision {
					return false
				}
				nowBoard["required_actor"] = oldBoard["required_actor"]
				delete(nowBoard, "decision_kind")
				delete(nowBoard, "choices")
			}
		}
		now["board"], old["board"] = encoded(nowBoard), encoded(oldBoard)
	}
	return bytes.Equal(encoded(now), encoded(old))
}

func sameEnvelope(a, b Envelope) bool {
	x, e := encode(a)
	if e != nil {
		return false
	}
	y, e := encode(b)
	return e == nil && bytes.Equal(x, y)
}
