//go:build integration

package matchstore

import (
	"context"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestReputationFinalizedExactBoundaryAndIdempotency(t *testing.T) {
	for _, amount := range []int64{39, 40, 100} {
		t.Run(fmt.Sprint(amount), func(t *testing.T) {
			s := testStore(t)
			ctx := context.Background()
			env := envelopeFixture(t)
			env.RulesHash = "rules-test"
			b := env.Match.Game.Promises
			var err error
			b, err = b.Offer(game.Promise{ID: "p", Payer: 0, Recipient: 1, AwardID: "award", Mode: "fixed", Amount: game.IntAmount(100)})
			if err != nil {
				t.Fatal(err)
			}
			b, err = b.Accept("p", 1)
			if err != nil {
				t.Fatal(err)
			}
			b, err = b.RecordAward(game.PromiseAward{ID: "award", Payer: 0, Amount: game.IntAmount(100)})
			if err != nil {
				t.Fatal(err)
			}
			b, err = b.CloseBoard()
			if err != nil {
				t.Fatal(err)
			}
			l := env.Match.Game.Ledger
			l, err = game.SettleReceipts(l, []game.FinancialReceipt{{Seat: 0, Amount: game.IntAmount(100)}})
			if err != nil {
				t.Fatal(err)
			}
			l.Phase = "settlement"
			if amount < 100 {
				b, err = b.OfferFinal("p", 0, game.IntAmount(amount))
				if err != nil {
					t.Fatal(err)
				}
				b, err = b.AnswerFinal("p", 1, true)
				if err != nil {
					t.Fatal(err)
				}
			}
			b, l, err = b.Pay("p", "pay", 0, l)
			if err != nil {
				t.Fatal(err)
			}
			for seat := 0; seat < 3; seat++ {
				b, err = b.Finish(seat)
				if err != nil {
					t.Fatal(err)
				}
				l, err = l.FinishSettlement(seat)
				if err != nil {
					t.Fatal(err)
				}
			}
			b, err = b.Finalize()
			if err != nil {
				t.Fatal(err)
			}
			l, err = l.Finalize()
			if err != nil {
				t.Fatal(err)
			}
			board, err := game.CloseCardBoard(env.Match.Game.Board)
			if err != nil {
				t.Fatal(err)
			}
			env.Match.Game.Board = board
			env.Match.Game.Promises = b
			env.Match.Game.Ledger = l
			env.Match.Game.Ending = "ordinary"
			if err = env.Validate(); err != nil {
				t.Fatal(err)
			}
			if err = s.Create(ctx, "m", env, []string{"a", "b", "c"}); err != nil {
				t.Fatal(err)
			}
			tx, err := s.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			for i := 0; i < 2; i++ {
				if err = syncReputation(ctx, tx, "m", env); err != nil {
					t.Fatal(err)
				}
			}
			if err = tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			var count int
			if err = s.pool.QueryRow(ctx, "SELECT count(*) FROM reputation_outcomes WHERE match_id='m'").Scan(&count); err != nil {
				t.Fatal(err)
			}
			want := 1
			if amount == 39 {
				want = 0
			}
			if count != want {
				t.Fatal("duplicate orwrong threshold", count)
			}
			if want == 1 {
				var outcome string
				if err = s.pool.QueryRow(ctx, "SELECT outcome FROM reputation_outcomes WHERE match_id='m'").Scan(&outcome); err != nil {
					t.Fatal(err)
				}
				label := "trust"
				if amount == 40 {
					label = "anyhoo"
				}
				if outcome != label {
					t.Fatal(outcome)
				}
			}
		})
	}
}

func TestReputationPendingVoidAndScamPrecedence(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	env := envelopeFixture(t)
	env.RulesHash = "rules-test"
	if err := s.Create(ctx, "m", env, []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	b := env.Match.Game.Promises
	var err error
	for _, id := range []string{"paid", "refused"} {
		b, err = b.Offer(game.Promise{ID: id, Payer: 0, Recipient: 1, AwardID: id, Mode: "fixed", Amount: game.IntAmount(100)})
		if err != nil {
			t.Fatal(err)
		}
		b, err = b.Accept(id, 1)
		if err != nil {
			t.Fatal(err)
		}
		b, err = b.RecordAward(game.PromiseAward{ID: id, Payer: 0, Amount: game.IntAmount(100)})
		if err != nil {
			t.Fatal(err)
		}
	}
	l := env.Match.Game.Ledger
	l, err = l.Charge("recipient-debt", 1, 2, game.IntAmount(100))
	if err != nil {
		t.Fatal(err)
	}
	l, err = game.SettleReceipts(l, []game.FinancialReceipt{{Seat: 0, Amount: game.IntAmount(100)}})
	if err != nil {
		t.Fatal(err)
	}
	b, l, err = b.Pay("paid", "transfer", 0, l)
	if err != nil {
		t.Fatal(err)
	}
	if l.Cash[1].Sign() != 0 || b.Promises[0].Paid.Cmp(game.IntAmount(100)) != 0 {
		t.Fatal("gross delivered amount lost to repayment")
	}
	b, err = b.CloseBoard()
	if err != nil {
		t.Fatal(err)
	}
	env.Match.Game.Promises = b
	write := func(e Envelope) {
		t.Helper()
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		if err = syncReputation(ctx, tx, "m", e); err != nil {
			t.Fatal(err)
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	count := func() int {
		var n int
		if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM reputation_outcomes").Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	write(env)
	if count() != 0 {
		t.Fatal("pending reputation increased")
	}
	void := env.Clone()
	void.Match.Game.Promises = void.Match.Game.Promises.Nullify()
	write(void)
	if count() != 0 {
		t.Fatal("void reputation increased")
	}
	b, err = b.Refuse("refused", 0)
	if err != nil {
		t.Fatal(err)
	}
	for seat := 0; seat < 3; seat++ {
		b, err = b.Finish(seat)
		if err != nil {
			t.Fatal(err)
		}
	}
	b, err = b.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	env.Match.Game.Promises = b
	write(env)
	write(env)
	var label string
	if err = s.pool.QueryRow(ctx, "SELECT outcome FROM reputation_outcomes").Scan(&label); err != nil || label != "scam" || count() != 1 {
		t.Fatal("exclusive scam precedence/idempotency", label, err)
	}
}
