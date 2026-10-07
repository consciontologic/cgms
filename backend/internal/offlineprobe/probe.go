// Package offlineprobe is an experimental, local-only P01 comparison boundary.
// Its omniscient synthetic snapshots are never network or presentation payloads.
package offlineprobe

import (
	"encoding/json"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
)

const Version = "cgms-offline-probe-v1"
const MaxCommands = 16

type Request struct {
	Schema   string         `json:"schema"`
	State    game.State     `json:"state"`
	Commands []game.Command `json:"commands"`
}
type Frame struct {
	State        game.State         `json:"state"`
	Events       []game.Event       `json:"events"`
	Observations []game.Observation `json:"observations"`
	Decision     *bots.Decision     `json:"decision,omitempty"`
}
type Response struct {
	Schema     string  `json:"schema"`
	Error      string  `json:"error,omitempty"`
	FailedStep int     `json:"failed_step"`
	Frames     []Frame `json:"frames,omitempty"`
}

// Execute replays a bounded synthetic board trace with no retained handles or I/O.
// Failed requests return no partial trace. Error messages never echo input data.
// Financial scenarios use their own schema; neither schema is a save store or mobile engine.
func Execute(raw []byte) []byte {
	// Routing does not validate: each schema still uses strict canonical.Decode.
	var envelope struct {
		Schema string `json:"schema"`
	}
	if len(raw) <= canonical.MaxBytes && json.Unmarshal(raw, &envelope) == nil && envelope.Schema == CardplayChoiceVersion {
		return executeCardplayChoice(raw)
	}
	if len(raw) <= canonical.MaxBytes && json.Unmarshal(raw, &envelope) == nil && envelope.Schema == CardplayVersion {
		return executeCardplay(raw)
	}
	if len(raw) <= canonical.MaxBytes && json.Unmarshal(raw, &envelope) == nil && envelope.Schema == FinanceVersion {
		return executeFinance(raw)
	}
	fail := func(code string, step int) []byte {
		b, _ := canonical.Marshal(Response{Schema: Version, Error: code, FailedStep: step})
		return b
	}
	var r Request
	if e := canonical.Decode(raw, &r); e != nil || r.Schema != Version || len(r.Commands) > MaxCommands {
		return fail("invalid_request", -1)
	}
	if r.State.Validate() != nil {
		return fail("invalid_state", -1)
	}
	// Bound the combinatorial narrow menu used solely to time a baseline decision.
	numbers := 0
	for _, c := range r.State.Cards {
		if c.Zone == game.Series {
			numbers++
		}
	}
	if numbers > 16 {
		return fail("probe_position_too_large", -1)
	}
	result := Response{Schema: Version, FailedStep: -1}
	state := r.State
	var events []game.Event
	for i := -1; i < len(r.Commands); i++ {
		if i >= 0 {
			var e error
			state, events, e = game.Apply(state, r.Commands[i])
			if e != nil {
				return fail("command_rejected", i)
			}
		}
		frame := Frame{State: state, Events: events}
		for seat := 1; seat <= len(state.Players); seat++ {
			obs, e := game.ObserveWithoutMenu(state, seat)
			if e != nil {
				return fail("observation_failed", i)
			}
			frame.Observations = append(frame.Observations, obs)
		}
		actor := state.Active
		for _, o := range frame.Observations {
			if o.RequiredActor > 0 {
				actor = o.RequiredActor
				break
			}
		}
		obs, e := game.Observe(state, actor)
		if e != nil {
			return fail("observation_failed", i)
		}
		if len(obs.Legal) > 0 {
			d, e := bots.Choose("heuristic@v1", obs, nil, 4096)
			if e != nil {
				return fail("decision_budget", i)
			}
			frame.Decision = &d
		}
		result.Frames = append(result.Frames, frame)
	}
	b, e := canonical.Marshal(result)
	if e != nil {
		return fail("output_budget", -1)
	}
	return b
}
