//go:build integration

package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"github.com/metaphy6/cgms/backend/internal/telemetry"
)

func TestBotRoomCreatesOnlyServerOwnedSeatsAndReplaysStart(t *testing.T) {
	for _, capacity := range []int{3, 4} {
		t.Run(fmt.Sprint(capacity), func(t *testing.T) {
			s := journeyServer(t)
			owner := journeyGuests(t, s, 1)[0]
			body := map[string]any{"command_id": "bot-room", "capacity": capacity, "games_per_match": 1, "bot_difficulty": "beginner"}
			data := journeyRequest(t, s, 0, owner.Token, "POST", "/v1/rooms", body, 200)
			var room struct {
				ID         string `json:"id"`
				Difficulty string `json:"bot_difficulty"`
				Members    []struct {
					Actor string `json:"actor_id"`
					Seat  int    `json:"seat"`
					Bot   bool   `json:"bot"`
				} `json:"members"`
			}
			if err := json.Unmarshal(data, &room); err != nil || room.Difficulty != "beginner" || len(room.Members) != capacity {
				t.Fatalf("bot roster: %s %v", data, err)
			}
			for i, m := range room.Members {
				if m.Seat != i+1 || m.Bot != (i > 0) || (i == 0 && m.Actor != owner.Account.ID) {
					t.Fatal("server-owned seat assignment", room.Members)
				}
				if i == 0 {
					continue
				}
				var kind string
				var credentials bool
				if err := s.cfg.Pool.QueryRow(context.Background(), `SELECT kind, username IS NOT NULL OR password_hash IS NOT NULL OR EXISTS(SELECT 1 FROM identity_sessions WHERE account_id=$1) FROM identity_accounts WHERE account_id=$1`, m.Actor).Scan(&kind, &credentials); err != nil || kind != "bot" || credentials {
					t.Fatal("bot must have no login capability", kind, credentials, err)
				}
			}
			if replay := journeyRequest(t, s, 0, owner.Token, "POST", "/v1/rooms", body, 200); !bytes.Equal(data, replay) {
				t.Fatal("room retry changed bot identities")
			}
			body["bot_difficulty"] = "advanced"
			journeyRequest(t, s, 0, owner.Token, "POST", "/v1/rooms", body, 409)
			journeyRequest(t, s, 0, owner.Token, "GET", "/v1/rooms/"+room.ID, nil, 200)
			start := map[string]string{"command_id": "start-bots"}
			first := journeyRequest(t, s, 0, owner.Token, "POST", "/v1/rooms/"+room.ID+"/matches", start, 200)
			if again := journeyRequest(t, s, 0, owner.Token, "POST", "/v1/rooms/"+room.ID+"/matches", start, 200); !bytes.Equal(first, again) {
				t.Fatal("start replay changed match")
			}
		})
	}
}

func botHumanWait(t *testing.T, s *Server, id string) (matchstore.Envelope, int64) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for {
		env, version, err := s.matches.Restore(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		observation, err := game.ObserveWithoutMenu(env.Match.Game.Board, 1)
		if err != nil {
			t.Fatal(err)
		}
		_, automatic := matchstore.NextServer(env, version)
		if !automatic && botHumanDecisionPending(env, observation) {
			return env, version
		}
		if time.Now().After(deadline) {
			t.Fatalf("bots did not reach human decision: version=%d active=%d pending=%v required=%d round_closed=%v departures_resolved=%v turn_started=%v server_pending=%v", version, observation.Active, env.Match.Game.Board.Pending != nil, observation.RequiredActor, env.Match.Game.RoundClosed, env.Match.Game.DeparturesResolved, env.Match.Game.TurnStarted, automatic)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func botHumanDecisionPending(env matchstore.Envelope, observation game.Observation) bool {
	l := env.Match.Game
	if l.RoundClosed && l.Ending == "" && !l.DeparturesResolved && !l.Board.Players[0].Departed {
		for _, choice := range l.DepartureChoices {
			if choice.Seat == 1 {
				return false
			}
		}
		return true
	}
	if l.RoundClosed {
		return false
	}
	return observation.RequiredActor == 1 || (env.Match.Game.TurnStarted && env.Match.Game.Board.Active == 1 && env.Match.Game.Board.Pending == nil)
}

func TestBotHumanDecisionIncludesRoundChoice(t *testing.T) {
	board, err := game.NewState(3, "round-choice-wait")
	if err != nil {
		t.Fatal(err)
	}
	board.Active = 2
	match, err := game.NewMatchLifecycle(board, 1)
	if err != nil {
		t.Fatal(err)
	}
	match.Game.RoundClosed = true
	env, err := matchstore.NewEnvelope(match, "rules-test")
	if err != nil {
		t.Fatal(err)
	}
	observation, err := game.ObserveWithoutMenu(board, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !botHumanDecisionPending(env, observation) {
		t.Fatal("explicit human round choice was mistaken for a bot stall")
	}
	env.Match.Game.DepartureChoices = []game.DepartureChoice{{Seat: 1}}
	if botHumanDecisionPending(env, observation) {
		t.Fatal("already answered round choice remained pending")
	}
	env.Match.Game.DepartureChoices = nil
	env.Match.Game.DeparturesResolved = true
	if botHumanDecisionPending(env, observation) {
		t.Fatal("completed departures fabricated a human turn")
	}
}

func TestBotSupervisorResumesAfterRestartAndNeverAnswersForHuman(t *testing.T) {
	for index, tier := range []string{"beginner", "standard", "advanced"} {
		t.Run(tier, func(t *testing.T) {
			s := journeyServer(t)
			owner := journeyGuests(t, s, 1)[0]
			capacity := 3 + index%2
			room, err := s.rooms.CreateWithBots(context.Background(), owner.Account.ID, "bots", capacity, 1, tier)
			if err != nil {
				t.Fatal(err)
			}
			id, err := s.rooms.Start(context.Background(), owner.Account.ID, room.ID, "start")
			if err != nil {
				t.Fatal(err)
			}
			// Start a fresh server before any browser requests: the persisted roster
			// and game, not an in-memory room callback, must drive the opponents.
			cfg := s.cfg
			cfg.Telemetry = telemetry.New(io.Discard, false)
			if err = s.Close(context.Background()); err != nil {
				t.Fatal(err)
			}
			s, err = New(cfg)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(context.Background()); err != nil {
					t.Error(err)
				}
			})
			s.Start()
			for turn := 0; turn < 3; turn++ {
				env, version := botHumanWait(t, s, id)
				// Repeated real worker scans leave a human decision untouched.
				before := automaticDeliveryCount(t, cfg.Telemetry, "ok")
				time.Sleep(160 * time.Millisecond)
				if delta := automaticDeliveryCount(t, cfg.Telemetry, "ok") - before; delta > 2 {
					t.Fatalf("idle bot match repeated automatic work %d times", delta)
				}
				_, later, err := s.matches.Restore(context.Background(), id)
				if err != nil || later != version {
					t.Fatal("worker answered for human", version, later, err)
				}
				data := journeyRequest(t, s, 0, owner.Token, "GET", "/v1/matches/"+id, nil, 200)
				var view struct {
					Projection matchstore.Projection `json:"projection"`
				}
				if err = json.Unmarshal(data, &view); err != nil {
					t.Fatal(err)
				}
				if len(view.Projection.Online.BotSeats) != capacity-1 || view.Projection.Online.BotDifficulty != tier {
					t.Fatal("bot labels lost after restart")
				}
				for _, card := range view.Projection.Board.Cards {
					if card.Controller != 1 && (card.Zone == game.Hand || card.Zone == game.ConcealedAce || card.Zone == game.Draw) {
						t.Fatal("bot hidden information leaked")
					}
				}
				kind, payload := "end-turn", map[string]any{}
				if env.Match.Game.RoundClosed {
					kind = "departure-choice"
					payload["accept"] = false
				} else if env.Match.Game.Board.Pending != nil {
					kind = "pass"
					payload["pending_action_id"] = env.Match.Game.Board.Pending.ID
				}
				body := map[string]any{"command_id": fmt.Sprintf("human-%d", turn), "game_id": env.Match.Game.Board.GameID, "expected_version": version, "type": kind, "payload": payload}
				first := journeyRequest(t, s, 0, owner.Token, "POST", "/v1/matches/"+id+"/commands", body, 200)
				if again := journeyRequest(t, s, 0, owner.Token, "POST", "/v1/matches/"+id+"/commands", body, 200); !bytes.Equal(first, again) {
					t.Fatal("human replay changed receipt")
				}
			}
			botHumanWait(t, s, id)
			var decisions int
			if err = s.cfg.Pool.QueryRow(context.Background(), `SELECT count(*) FROM command_results JOIN identity_accounts ON account_id=actor_id WHERE match_id=$1 AND kind='bot'`, id).Scan(&decisions); err != nil || decisions == 0 {
				t.Fatal("no bot decisions persisted", decisions, err)
			}
			if err = s.matches.VerifyJournal(context.Background(), id); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBotCoupInterruptsHumanTurnOrDecisionWithoutAnsweringForHuman(t *testing.T) {
	for _, pending := range []bool{false, true} {
		for _, winning := range []bool{false, true} {
			t.Run(fmt.Sprintf("pending=%v/winning=%v", pending, winning), func(t *testing.T) {
				s := journeyServer(t)
				ctx := context.Background()
				owner := journeyGuests(t, s, 1)[0]
				room, err := s.rooms.CreateWithBots(ctx, owner.Account.ID, "room", 3, 1, "beginner")
				if err != nil {
					t.Fatal(err)
				}
				b, err := game.NewState(3, "bot-coup")
				if err != nil {
					t.Fatal(err)
				}
				move := func(seat int, zone game.Zone, cards ...string) {
					b, err = b.Move(cards, seat, zone, true)
					if err != nil {
						t.Fatal(err)
					}
				}
				b.Order = []int{1, 2, 3}
				for i := range b.Players {
					b.Players[i].History = []game.Suit{game.Clubs}
				}
				if pending {
					move(1, game.Series, "deck-1-clubs-08", "deck-1-clubs-07", "deck-1-diamonds-02")
					move(2, game.Series, "deck-1-hearts-08", "deck-1-spades-06", "deck-1-spades-09")
					move(2, game.Attachment, "deck-1-spades-11")
					b.Players[1].History = []game.Suit{game.Hearts, game.Clubs, game.Spades, game.Diamonds}
					for _, c := range []game.Command{
						{ID: "attack", GameID: b.GameID, Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}},
						{ID: "negotiate", GameID: b.GameID, Kind: "negotiate", Actor: 2, WindowID: "attack"},
					} {
						b, _, err = game.Apply(b, c)
						if err != nil {
							t.Fatal(err)
						}
					}
					if b.Pending.Decision.Actor != 1 {
						t.Fatal("fixture must require the human")
					}
				}
				if winning {
					move(3, game.Hand, "deck-1-hearts-12", "deck-2-hearts-12", "deck-1-diamonds-12", "deck-2-diamonds-12")
				}
				m, err := game.NewMatchLifecycle(b, 1)
				if err != nil {
					t.Fatal(err)
				}
				m.Game.TurnStarted = true
				env, err := matchstore.NewEnvelope(m, "rules-test")
				if err != nil {
					t.Fatal(err)
				}
				env.BotDifficulty = room.BotDifficulty
				actors := []string{room.Members[0].ActorID, room.Members[1].ActorID, room.Members[2].ActorID}
				if err = s.matches.Create(ctx, "coup", env, actors); err != nil {
					t.Fatal(err)
				}
				if _, err = s.automaticStep(ctx, "coup"); err != nil {
					t.Fatal(err)
				}
				after, version, err := s.matches.Restore(ctx, "coup")
				if err != nil {
					t.Fatal(err)
				}
				if winning {
					if version != 1 || after.Match.Game.Declarer != 3 || after.Match.Game.Ending != "coup" || after.Match.Game.Board.Pending != nil {
						t.Fatalf("bot missed a legal out-of-turn Coup: version=%d ending=%s declarer=%d", version, after.Match.Game.Ending, after.Match.Game.Declarer)
					}
				} else if version != 0 || game.Digest(env) != game.Digest(after) {
					t.Fatal("bot answered or mutated a human boundary")
				}
			})
		}
	}
}

func TestBotSettlementFinishesAndStartsNextGameWithHumanEconomyOnly(t *testing.T) {
	for _, capacity := range []int{3, 4} {
		t.Run(fmt.Sprint(capacity), func(t *testing.T) {
			s := economyJourneyServer(t)
			ctx := context.Background()
			owner := journeyGuests(t, s, 1)[0]
			room, err := s.rooms.CreateWithBots(ctx, owner.Account.ID, "settlement-room", capacity, 2, "standard")
			if err != nil {
				t.Fatal(err)
			}
			b, err := game.NewState(capacity, "settled-game")
			if err != nil {
				t.Fatal(err)
			}
			m, err := game.NewMatchLifecycle(b, 2)
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Board, err = game.CloseCardBoard(m.Game.Board)
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Ledger, err = m.Game.Ledger.CloseOrdinary(nil)
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Promises, err = m.Game.Promises.CloseBoard()
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Ending = "ordinary"
			env, err := matchstore.NewEnvelope(m, "rules-test")
			if err != nil {
				t.Fatal(err)
			}
			env.BotDifficulty = room.BotDifficulty
			actors := make([]string, capacity)
			for i, member := range room.Members {
				actors[i] = member.ActorID
			}
			if err = s.matches.Create(ctx, "settlement", env, actors); err != nil {
				t.Fatal(err)
			}
			s.Start()
			deadline := time.Now().Add(5 * time.Second)
			var version int64
			for {
				env, version, err = s.matches.Restore(ctx, "settlement")
				if err != nil {
					t.Fatal(err)
				}
				finished := 0
				for _, done := range env.Match.Game.Ledger.Finished {
					if done {
						finished++
					}
				}
				if finished == capacity-1 {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("bots did not finish settlement: version=%d finished=%d", version, finished)
				}
				time.Sleep(10 * time.Millisecond)
			}
			if env.Match.Game.Ledger.Finished[0] {
				t.Fatal("bot accepted human settlement")
			}
			time.Sleep(160 * time.Millisecond)
			_, later, err := s.matches.Restore(ctx, "settlement")
			if err != nil || later != version {
				t.Fatal("human settlement wait changed", version, later, err)
			}
			body := map[string]any{"command_id": "human-finish", "game_id": "settled-game", "expected_version": version, "type": "finish-settlement", "payload": map[string]any{}}
			first := journeyRequest(t, s, 0, owner.Token, "POST", "/v1/matches/settlement/commands", body, 200)
			deadline = time.Now().Add(10 * time.Second)
			for {
				env, _, err = s.matches.Restore(ctx, "settlement")
				if err != nil {
					t.Fatal(err)
				}
				if env.Match.Game.Board.GameID != "settled-game" {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("bots blocked next-game admission")
				}
				time.Sleep(10 * time.Millisecond)
			}
			if env.BotDifficulty != "standard" || len(env.Match.Instances) != 2 || len(env.Match.Game.Ledger.Completed) != 1 {
				t.Fatal("bot configuration or financial history lost across next game")
			}
			if replay := journeyRequest(t, s, 0, owner.Token, "POST", "/v1/matches/settlement/commands", body, 200); !bytes.Equal(first, replay) {
				t.Fatal("old-game human finish receipt changed")
			}
			var charges, rewards, botAccounts int
			if err = s.cfg.Pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM economy_charges), (SELECT count(*) FROM economy_rewards), (SELECT count(*) FROM economy_accounts JOIN identity_accounts USING(account_id) WHERE kind='bot')`).Scan(&charges, &rewards, &botAccounts); err != nil || charges != 2 || rewards != 1 || botAccounts != 0 {
				t.Fatal("bot economy leakage or blocked second game", charges, rewards, botAccounts, err)
			}
			botHumanWait(t, s, "settlement")
			if err = s.matches.VerifyJournal(ctx, "settlement"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBotRoundWaitAcceptsFirstHumanChoiceThenResumes(t *testing.T) {
	for _, leave := range []bool{false, true} {
		t.Run(fmt.Sprintf("leave=%v", leave), func(t *testing.T) {
			s := journeyServer(t)
			ctx := context.Background()
			owner := journeyGuests(t, s, 1)[0]
			room, err := s.rooms.CreateWithBots(ctx, owner.Account.ID, "round-room", 3, 1, "beginner")
			if err != nil {
				t.Fatal(err)
			}
			b, err := game.NewState(3, "round-boundary")
			if err != nil {
				t.Fatal(err)
			}
			b.Order = []int{1, 2, 3}
			b.Active = 3
			for i := range b.Players {
				b.Players[i].History = []game.Suit{game.Clubs}
			}
			m, err := game.NewMatchLifecycle(b, 1)
			if err != nil {
				t.Fatal(err)
			}
			m.Game.RoundClosed = true
			env, err := matchstore.NewEnvelope(m, "rules-test")
			if err != nil {
				t.Fatal(err)
			}
			env.BotDifficulty = "beginner"
			actors := []string{room.Members[0].ActorID, room.Members[1].ActorID, room.Members[2].ActorID}
			if err = s.matches.Create(ctx, "round", env, actors); err != nil {
				t.Fatal(err)
			}
			// Once the human sees the round prompt, background scans must not change the
			// version underneath their first answer. Bots retain independent stay choices.
			for i := 0; i < 4; i++ {
				if _, err = s.automaticStep(ctx, "round"); err != nil {
					t.Fatal(err)
				}
			}
			env, version, err := s.matches.Restore(ctx, "round")
			if err != nil || version != 0 || len(env.Match.Game.DepartureChoices) != 0 || env.Match.Game.DeparturesResolved {
				t.Fatal("human round wait not preserved", version, err)
			}
			data := journeyRequest(t, s, 0, owner.Token, "GET", "/v1/matches/round", nil, 200)
			var view struct {
				Projection matchstore.Projection `json:"projection"`
			}
			if err = json.Unmarshal(data, &view); err != nil || !view.Projection.Online.RoundClosed || !view.Projection.Online.DeparturePending {
				t.Fatal("human departure prompt missing", err)
			}
			body := map[string]any{"command_id": "human-choice", "game_id": b.GameID, "expected_version": version, "type": "departure-choice", "payload": map[string]any{"accept": leave}}
			first := journeyRequest(t, s, 0, owner.Token, "POST", "/v1/matches/round/commands", body, 200)
			for i := 0; i < 10 && env.Match.Game.Board.Round == 1; i++ {
				if _, err = s.automaticStep(ctx, "round"); err != nil {
					t.Fatal(err)
				}
				env, version, err = s.matches.Restore(ctx, "round")
				if err != nil {
					t.Fatal(err)
				}
			}
			if version < 4 || env.Match.Game.Board.Round != 2 || env.Match.Game.RoundClosed || env.Match.Game.Board.Players[0].Departed != leave {
				t.Fatal("explicit choice did not resume the next round", version, leave, err)
			}
			// No stream projection exposes another participant's collected choice.
			events, err := s.matches.Events(ctx, "round", owner.Account.ID, 0, 10)
			if err != nil || len(events) != int(version) {
				t.Fatal("round stream", len(events), err)
			}
			for _, event := range events {
				var eventProjection matchstore.Projection
				if err = json.Unmarshal(event.Projection, &eventProjection); err != nil || eventProjection.Board.Seat != 1 || eventProjection.Online.DeparturePending {
					t.Fatal("completed own choice remains pending", err)
				}
				if bytes.Contains(event.Projection, []byte("departure_choices")) {
					t.Fatal("stream disclosed private simultaneous answers")
				}
			}
			if replay := journeyRequest(t, s, 0, owner.Token, "POST", "/v1/matches/round/commands", body, 200); !bytes.Equal(first, replay) {
				t.Fatal("choice retry changed receipt")
			}
			_, afterReplay, err := s.matches.Restore(ctx, "round")
			if err != nil || afterReplay != version {
				t.Fatal("choice replay advanced twice", afterReplay, version, err)
			}
			if err = s.matches.VerifyJournal(ctx, "round"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOnlineBotPurchaseCycleEndsAcrossRestartAndReceiptReplay(t *testing.T) {
	s := journeyServer(t)
	ctx := context.Background()
	owner := journeyGuests(t, s, 1)[0]
	room, err := s.rooms.CreateWithBots(ctx, owner.Account.ID, "purchase-cycle-room", 3, 1, "advanced")
	if err != nil {
		t.Fatal(err)
	}
	b, err := game.NewState(3, "purchase-cycle-game")
	if err != nil {
		t.Fatal(err)
	}
	b.Round, b.Turn, b.Active, b.Order = 2, 6, 2, []int{2, 1, 3}
	b.DrawOrder = nil
	paid := 0
	for i := range b.Cards {
		card := &b.Cards[i]
		card.Controller, card.Zone = 1, game.Hand
		if card.Card.Rank == 1 {
			card.Zone = game.ConcealedAce
		} else if card.Card.Suit == game.Spades && card.Card.Rank <= 10 && paid < 13 {
			card.Controller, card.Zone = 2, game.Series
			b.PublicHistory = append(b.PublicHistory, card.Card.ID)
			paid++
		}
	}
	for i := range b.Players {
		b.Players[i].History = []game.Suit{game.Spades}
	}
	m, err := game.NewMatchLifecycle(b, 1)
	if err != nil {
		t.Fatal(err)
	}
	m.Game.TurnStarted = true
	env, err := matchstore.NewEnvelope(m, "rules-test")
	if err != nil {
		t.Fatal(err)
	}
	env.BotDifficulty = "advanced"
	actors := []string{room.Members[0].ActorID, room.Members[1].ActorID, room.Members[2].ActorID}
	if err = s.matches.Create(ctx, "purchase-cycle", env, actors); err != nil {
		t.Fatal(err)
	}
	planningContext, cancelPlanning := context.WithTimeout(ctx, 2*time.Second)
	planningStarted := time.Now()
	actor, intent, ok, err := s.botIntent(planningContext, "purchase-cycle", env, 0)
	cancelPlanning()
	t.Logf("dense bot planning elapsed=%s reason=%s", time.Since(planningStarted), automaticFailureReason(err))
	if err != nil || !ok || actor != actors[1] || intent.Request.Operation == nil || intent.Request.Operation.Kind != "end-turn" {
		t.Fatal("online runtime re-entered the empty-supply purchase loop", ok, err)
	}
	// A fresh runtime has no in-memory turn counter. The same durable position
	// must produce the same legal completion and command identity after restart.
	restarted, err := New(s.cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		closeContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := restarted.Close(closeContext); err != nil {
			t.Error("restarted runtime did not stop", err)
		}
	}()
	restored, version, err := restarted.matches.Restore(ctx, "purchase-cycle")
	if err != nil || version != 0 {
		t.Fatal(err)
	}
	againActor, again, againOK, err := restarted.botIntent(ctx, "purchase-cycle", restored, version)
	if err != nil || !againOK || againActor != actor || game.Digest(again) != game.Digest(intent) {
		t.Fatal("restart changed bot progress choice or idempotency identity", err)
	}
	stepContext, cancelStep := context.WithTimeout(ctx, 2*time.Second)
	stepStarted := time.Now()
	_, err = restarted.automaticStep(stepContext, "purchase-cycle")
	cancelStep()
	t.Logf("dense automatic transaction elapsed=%s reason=%s", time.Since(stepStarted), automaticFailureReason(err))
	if err != nil {
		t.Fatal(err)
	}
	restarted.Start()
	after, version := botHumanWait(t, restarted, "purchase-cycle")
	if after.Match.Game.Board.Active != 1 || !after.Match.Game.TurnStarted || version != 2 || after.Match.Game.Board.Turn != 7 {
		t.Fatal("automatic runtime did not return usable control to the human", version)
	}
	if _, err = restarted.automaticStep(ctx, "purchase-cycle"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		receipt, err := restarted.matches.Submit(ctx, "purchase-cycle", actor, intent)
		if err != nil || receipt.Version != 1 {
			t.Fatal("acknowledgement replay repeated the bot turn", err)
		}
	}
	final, finalVersion, err := restarted.matches.Restore(ctx, "purchase-cycle")
	if err != nil || finalVersion != version || game.Digest(after) != game.Digest(final) {
		t.Fatal("bot replay or repeated scan mutated the human turn", err)
	}
	var commands, humanCommands int
	if err = s.cfg.Pool.QueryRow(ctx, `SELECT count(*),count(*) FILTER (WHERE actor_id=$2) FROM command_results WHERE match_id=$1`, "purchase-cycle", actors[0]).Scan(&commands, &humanCommands); err != nil || commands != 1 || humanCommands != 0 {
		t.Fatal("duplicate bot command or fabricated human response", commands, humanCommands, err)
	}
	if err = restarted.matches.VerifyJournal(ctx, "purchase-cycle"); err != nil {
		t.Fatal(err)
	}
}
