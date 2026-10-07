package randomstream

import (
	"encoding/hex"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"os"
	"slices"
	"strconv"
	"testing"
)

func TestSeedVector(t *testing.T) {
	s, e := Derive(Identity{Root: "42", Population: 3, Pairing: "pilot-3p", Block: "dev-000", Rotation: "r0", Kind: "deal"})
	if e != nil {
		t.Fatal(e)
	}
	if hex.EncodeToString(s[:]) != "a639efa27dd3c8af295e7928f17d4fb493e15347a126caadf2d353d4c6a1b9a3" {
		t.Fatal(s)
	}
}
func TestResumeAndRejection(t *testing.T) {
	s := New([32]byte{})
	for i := 0; i < 37; i++ {
		s.Uint64()
	}
	p, e := s.Snapshot()
	if e != nil {
		t.Fatal(e)
	}
	r, e := Restore(p)
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 200; i++ {
		if s.Uint64() != r.Uint64() {
			t.Fatal("resume differs")
		}
	}
	if s.Counter != r.Counter {
		t.Fatal("counter")
	}
	words := []uint64{0, 1, 9223372036854775807, 9223372036854775809}
	j := 0
	v, e := Sample(func() uint64 { x := words[j]; j++; return x }, 9223372036854775809)
	if e != nil || j != 3 || v != 9223372036854775807 {
		t.Fatalf("%d %d %v", v, j, e)
	}
}
func TestShuffle(t *testing.T) {
	s := New([32]byte{})
	for n := 0; n < 105; n++ {
		a := make([]int, n)
		for i := range a {
			a[i] = i
		}
		before := s.Counter
		if e := s.Shuffle(n, func(i, j int) { a[i], a[j] = a[j], a[i] }); e != nil {
			t.Fatal(e)
		}
		seen := map[int]bool{}
		for _, v := range a {
			if v < 0 || v >= n || seen[v] {
				t.Fatal("not permutation")
			}
			seen[v] = true
		}
		if n < 2 && before != s.Counter {
			t.Fatal("empty consumed")
		}
	}
}
func TestChaChaAndShuffleGolden(t *testing.T) {
	s := New([32]byte{})
	expected := []uint64{12432809566553409497, 7766219816435805978, 2207117295956785947, 8400577010494958244, 17544015947949813325, 6729310184567665740, 4793391916538700715, 16392521451148261596}
	for _, want := range expected {
		if got := s.Uint64(); got != want {
			t.Fatalf("word %d want %d", got, want)
		}
	}
	a := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	_ = s.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	want := []int{1, 8, 7, 3, 2, 0, 4, 9, 5, 6}
	for i := range a {
		if a[i] != want[i] {
			t.Fatal(a)
		}
	}
	if s.Counter != 17 {
		t.Fatal(s.Counter)
	}
}
func TestInvalidIdentities(t *testing.T) {
	base := Identity{Root: "42", Population: 3, Pairing: "pilot", Block: "dev", Rotation: "r0", Kind: "deal"}
	for _, root := range []string{"-1", "01", "+1", "1.0", "115792089237316195423570985008687907853269984665640564039457584007913129639936"} {
		x := base
		x.Root = root
		if _, e := Derive(x); e == nil {
			t.Fatal("invalid root")
		}
	}
	for _, kind := range []string{"unknown", "bot"} {
		x := base
		x.Kind = kind
		if _, e := Derive(x); e == nil {
			t.Fatal("invalid kind/seat")
		}
	}
	if _, e := Sample(func() uint64 { return 1 }, 0); e == nil {
		t.Fatal("zero bound")
	}
	s := New([32]byte{})
	p, _ := s.Snapshot()
	p.Algorithm = "future"
	if _, e := Restore(p); e == nil {
		t.Fatal("future snapshot")
	}
}
func TestCuratedVectorFile(t *testing.T) {
	b, e := os.ReadFile("../../testdata/randomstream/chacha8-v1.json")
	if e != nil {
		t.Fatal(e)
	}
	var x struct {
		Algorithm string   `json:"algorithm"`
		Seed      string   `json:"seed_hex"`
		Words     []string `json:"words_decimal"`
		Input     []int    `json:"shuffle_input"`
		Output    []int    `json:"shuffle_output"`
		Counter   string   `json:"final_counter"`
	}
	if e = canonical.Decode(b, &x); e != nil {
		t.Fatal(e)
	}
	raw, e := hex.DecodeString(x.Seed)
	if e != nil || len(raw) != 32 {
		t.Fatal("seed")
	}
	var seed [32]byte
	copy(seed[:], raw)
	s := New(seed)
	for _, text := range x.Words {
		word, e := strconv.ParseUint(text, 10, 64)
		if e != nil || s.Uint64() != word {
			t.Fatal("word vector")
		}
	}
	if e = s.Shuffle(len(x.Input), func(i, j int) { x.Input[i], x.Input[j] = x.Input[j], x.Input[i] }); e != nil {
		t.Fatal(e)
	}
	if !slices.Equal(x.Input, x.Output) || strconv.FormatUint(s.Counter, 10) != x.Counter {
		t.Fatal("shuffle vector")
	}
}
