// Package randomstream provides offline deterministic research streams, not online entropy.
package randomstream

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"math/big"
	"math/rand/v2"
)

const Algorithm = "go-chacha8-v1"
const Derivation = "cgms-seed-v1"
const Sampler = "cgms-u64-reject-v1"
const ShuffleVersion = "cgms-fisher-yates-v1"

type Identity struct {
	Root, Pairing, Block, Rotation, Kind       string
	Population, Match, Slot, Replacement, Seat int
}

func validRoot(s string) bool {
	n, ok := new(big.Int).SetString(s, 10)
	return ok && n.Sign() >= 0 && n.BitLen() <= 256 && n.String() == s
}
func Derive(i Identity) ([32]byte, error) {
	if !validRoot(i.Root) || (i.Population != 3 && i.Population != 4) || i.Match < 0 || i.Slot < 0 || i.Replacement < 0 || i.Seat < 0 || i.Seat > i.Population {
		return [32]byte{}, fmt.Errorf("invalid stream identity")
	}
	switch i.Kind {
	case "bot":
		if i.Seat == 0 {
			return [32]byte{}, fmt.Errorf("bot seat required")
		}
	case "deal", "initiative", "effect":
		if i.Seat != 0 {
			return [32]byte{}, fmt.Errorf("environment seat must be zero")
		}
	default:
		return [32]byte{}, fmt.Errorf("unknown stream kind")
	}
	b, e := canonical.Marshal([]any{Derivation, i.Root, i.Population, i.Pairing, i.Block, i.Rotation, i.Match, i.Slot, i.Replacement, i.Kind, i.Seat})
	if e != nil {
		return [32]byte{}, e
	}
	return sha256.Sum256(b), nil
}
func AnalysisSeed(root string, population int, pairing string) ([32]byte, error) {
	if !validRoot(root) || (population != 3 && population != 4) {
		return [32]byte{}, fmt.Errorf("analysis identity")
	}
	b, e := canonical.Marshal([]any{"cgms-analysis-v1", root, population, pairing})
	if e != nil {
		return [32]byte{}, e
	}
	return sha256.Sum256(b), nil
}

type Stream struct {
	rng     *rand.ChaCha8
	Counter uint64
}
type Snapshot struct {
	Algorithm string `json:"algorithm"`
	State     string `json:"state"`
	Counter   string `json:"counter"`
}

func New(seed [32]byte) *Stream  { return &Stream{rng: rand.NewChaCha8(seed)} }
func (s *Stream) Uint64() uint64 { s.Counter++; return s.rng.Uint64() }
func Sample(next func() uint64, n uint64) (uint64, error) {
	if n == 0 {
		return 0, fmt.Errorf("empty sample")
	}
	threshold := -n % n
	for {
		x := next()
		if x >= threshold {
			return x % n, nil
		}
	}
}
func (s *Stream) Sample(n uint64) (uint64, error) { return Sample(s.Uint64, n) }
func (s *Stream) Shuffle(n int, swap func(int, int)) error {
	if n < 0 {
		return fmt.Errorf("negative length")
	}
	for i := n - 1; i > 0; i-- {
		j, e := s.Sample(uint64(i + 1))
		if e != nil {
			return e
		}
		swap(i, int(j))
	}
	return nil
}
func (s *Stream) Snapshot() (Snapshot, error) {
	b, e := s.rng.MarshalBinary()
	return Snapshot{Algorithm, hex.EncodeToString(b), fmt.Sprint(s.Counter)}, e
}
func Restore(p Snapshot) (*Stream, error) {
	if p.Algorithm != Algorithm {
		return nil, fmt.Errorf("PRNG version")
	}
	n, ok := new(big.Int).SetString(p.Counter, 10)
	if !ok || n.Sign() < 0 || n.BitLen() > 64 || n.String() != p.Counter {
		return nil, fmt.Errorf("counter")
	}
	b, e := hex.DecodeString(p.State)
	if e != nil {
		return nil, e
	}
	s := New([32]byte{})
	if e = s.rng.UnmarshalBinary(b); e != nil {
		return nil, e
	}
	s.Counter = n.Uint64()
	return s, nil
}
