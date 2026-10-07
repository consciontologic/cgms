//go:build integration

package matchstore

import (
	"context"
	"encoding/json"
	"testing"
)

func journalFixture(t *testing.T) *Store {
	t.Helper()
	s := testStore(t)
	fixture(t, s)
	ctx := context.Background()
	if _, err := s.Submit(ctx, "m", "a", nullify("consent", 1)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Continue(ctx, "m", ServerIntent{ID: "deal", GameID: "g1", ExpectedVersion: 1, Request: ServerRequest{GameID: "g1", Kind: "deal"}}); err != nil {
		t.Fatal(err)
	}
	return s
}
func TestPrivateJournalReplaysActorAndServer(t *testing.T) {
	s := journalFixture(t)
	if err := s.VerifyJournal(context.Background(), "m"); err != nil {
		t.Fatal(err)
	}
}
func TestPrivateJournalRejectsTampering(t *testing.T) {
	for _, change := range []string{"input", "snapshot", "chance-missing", "chance-extra", "chance-prior", "sequence", "missing", "final", "initial"} {
		t.Run(change, func(t *testing.T) {
			s := journalFixture(t)
			ctx := context.Background()
			if change == "missing" {
				if _, err := s.pool.Exec(ctx, `DELETE FROM match_events WHERE seq=1`); err != nil {
					t.Fatal(err)
				}
			} else if change == "final" {
				e, _, err := s.Restore(ctx, "m")
				if err != nil {
					t.Fatal(err)
				}
				e.Match.Game.Board.Players[0].Confined = true
				b, _ := json.Marshal(e)
				if _, err = s.pool.Exec(ctx, `UPDATE matches SET snapshot=$1`, b); err != nil {
					t.Fatal(err)
				}
			} else if change == "initial" {
				var raw []byte
				if err := s.pool.QueryRow(ctx, `SELECT config FROM matches`).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				var c persistedConfig
				if err := json.Unmarshal(raw, &c); err != nil {
					t.Fatal(err)
				}
				c.Initial.Match.Game.Board.Players[0].Confined = true
				raw, _ = json.Marshal(c)
				if _, err := s.pool.Exec(ctx, `UPDATE matches SET config=$1`, raw); err != nil {
					t.Fatal(err)
				}
			} else {
				var raw []byte
				if err := s.pool.QueryRow(ctx, `SELECT event FROM match_events WHERE seq=2`).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				var j JournalEntry
				if err := json.Unmarshal(raw, &j); err != nil {
					t.Fatal(err)
				}
				switch change {
				case "input":
					var in ServerIntent
					if err := json.Unmarshal(j.Input, &in); err != nil {
						t.Fatal(err)
					}
					in.ExpectedVersion = 0
					j.Input, _ = json.Marshal(in)
				case "snapshot":
					j.Snapshot.Match.Game.Board.Players[0].Confined = true
				case "chance-missing":
					j.Snapshot.Chance = map[string][]string{}
				case "chance-extra":
					j.Snapshot.Chance["unused"] = []string{"bogus"}
				case "chance-prior": // Previously committed outcomes cannot be replaced in a later record.
					var first []byte
					if err := s.pool.QueryRow(ctx, `SELECT event FROM match_events WHERE seq=1`).Scan(&first); err != nil {
						t.Fatal(err)
					}
					var before JournalEntry
					_ = json.Unmarshal(first, &before)
					before.Snapshot.Chance["old"] = []string{"old"}
					first, _ = json.Marshal(before)
					if _, err := s.pool.Exec(ctx, `UPDATE match_events SET event=$1 WHERE seq=1`, first); err != nil {
						t.Fatal(err)
					}
				case "sequence":
					j.Version = 3
				}
				raw, _ = json.Marshal(j)
				if _, err := s.pool.Exec(ctx, `UPDATE match_events SET event=$1 WHERE seq=2`, raw); err != nil {
					t.Fatal(err)
				}
			}
			if err := s.VerifyJournal(ctx, "m"); err == nil {
				t.Fatal("tampered journal accepted", change)
			}
		})
	}
}

func TestPrivateJournalPreservesPreviouslyCommittedChance(t *testing.T) {
	s := journalFixture(t)
	ctx := context.Background()
	in := nullify("second-consent", 2)
	in.ExpectedVersion = 2
	if _, err := s.Submit(ctx, "m", "b", in); err != nil {
		t.Fatal(err)
	}
	if err := s.VerifyJournal(ctx, "m"); err != nil {
		t.Fatal(err)
	}
	var raw []byte
	if err := s.pool.QueryRow(ctx, `SELECT event FROM match_events WHERE seq=3`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var j JournalEntry
	if err := json.Unmarshal(raw, &j); err != nil {
		t.Fatal(err)
	}
	for key, cards := range j.Snapshot.Chance {
		if len(cards) < 2 {
			t.Fatal("missing shuffle fixture")
		}
		cards[0], cards[1] = cards[1], cards[0]
		j.Snapshot.Chance[key] = cards
		break
	}
	raw, _ = json.Marshal(j)
	if _, err := s.pool.Exec(ctx, `UPDATE match_events SET event=$1 WHERE seq=3`, raw); err != nil {
		t.Fatal(err)
	}
	if err := s.VerifyJournal(ctx, "m"); err == nil {
		t.Fatal("prior committed outcome replaced")
	}
}

func TestReplayChanceDoesNotGenerateMissingOutcome(t *testing.T) {
	e := envelopeFixture(t)
	e.replay = &chanceReplay{used: map[string]bool{}}
	if _, _, err := SecurePermutation(e, "missing", []string{"a", "b"}); err == nil {
		t.Fatal("missing recorded outcome generated")
	}
}
