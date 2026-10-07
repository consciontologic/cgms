package game

import (
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"testing"
)

// These synthetic vectors were measured with pinned Go ChaCha8 and the specified
// descending Fisher-Yates sampler. A redeal advances the same stream.
func TestDealSeededRedealVectors(t *testing.T) {
	for _, v := range []struct{ seed, redeals int }{{0, 0}, {19, 1}, {1097, 2}} {
		s := position(t)
		rng := randomstream.New([32]byte{byte(v.seed), byte(v.seed >> 8)})
		count := 0
		for tries := 0; tries < 5; tries++ {
			order := append([]string(nil), s.DrawOrder...)
			if e := rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] }); e != nil {
				t.Fatal(e)
			}
			n, redo, e := DealAttempt(s, order)
			if e != nil {
				t.Fatal(e)
			}
			if redo {
				count++
				continue
			}
			if e = n.Validate(); e != nil {
				t.Fatal(e)
			}
			if len(n.CurrentSeries(1)) == 0 {
				t.Fatal("valid initial diamonds missing")
			}
			break
		}
		if count != v.redeals {
			t.Fatalf("synthetic vector %d redeals=%d want %d", v.seed, count, v.redeals)
		}
	}
}
