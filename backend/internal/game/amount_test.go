package game

import (
	"encoding/json"
	"strconv"
	"testing"
)

func TestAmountExactAndImmutable(t *testing.T) {
	a, err := NewAmount("1", "3")
	if err != nil {
		t.Fatal(err)
	}
	b := a.Add(a).Add(a)
	if b.String() != "1" || a.String() != "1/3" {
		t.Fatalf("exact immutable thirds: sum=%s input=%s", b, a)
	}
}
func TestAmountCanonicalJSON(t *testing.T) {
	for _, raw := range []string{`{"numerator":"2","denominator":"4"}`, `{"numerator":"-0","denominator":"1"}`, `{"numerator":"0","denominator":"2"}`, `{"numerator":"1","denominator":"0"}`, `{"Numerator":"1","denominator":"2"}`, `{"numerator":"1","numerator":"2","denominator":"1"}`, `{"numerator":"1","denominator":"2","x":1}`} {
		var a Amount
		if json.Unmarshal([]byte(raw), &a) == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	a, _ := NewAmount("-6", "4")
	b, _ := json.Marshal(a)
	if string(b) != `{"numerator":"-3","denominator":"2"}` {
		t.Fatal(string(b))
	}
}
func FuzzAmountArithmetic(f *testing.F) {
	f.Add(int64(1), int64(3))
	f.Fuzz(func(t *testing.T, n, d int64) {
		if d <= 0 {
			return
		}
		a, e := NewAmount(strconv.FormatInt(n, 10), strconv.FormatInt(d, 10))
		if e != nil {
			t.Fatal(e)
		}
		if a.Add(a.Neg()).Sign() != 0 {
			t.Fatal("inverse")
		}
		b, _ := json.Marshal(a)
		var c Amount
		if e = json.Unmarshal(b, &c); e != nil || a.Cmp(c) != 0 {
			t.Fatal("roundtrip", e)
		}
	})
}
