package canonical

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestVector(t *testing.T) {
	var v any
	if e := Decode([]byte(`{ "b":2, "a":1 }`), &v); e != nil {
		t.Fatal(e)
	}
	b, e := Marshal(v)
	if e != nil || string(b) != `{"a":1,"b":2}` {
		t.Fatalf("canonical %s %v", b, e)
	}
	h, _ := Hash(v)
	if h != "43258cff783fe7036d8a43033f830adfc60ec037382473548ac742b888292777" {
		t.Fatal(h)
	}
}
func TestReject(t *testing.T) {
	for _, s := range []string{`{"a":1,"a":2}`, `1.0`, `1e2`, `9007199254740992`, `"\ud800"`, `"\udc00"`, `{} {}`, string([]byte{34, 255, 34}), strings.Repeat("[", 65) + "0" + strings.Repeat("]", 65)} {
		var v any
		if Decode([]byte(s), &v) == nil {
			t.Errorf("accepted %q", s)
		}
	}
	var x struct {
		Flag bool `json:"flag"`
	}
	for _, s := range []string{`{"Flag":true}`, `{"other":true}`, `{"flag":"true"}`} {
		if Decode([]byte(s), &x) == nil {
			t.Errorf("accepted %s", s)
		}
	}
}
func TestUnicode(t *testing.T) {
	b, e := Marshal(map[string]any{"text": "<>&/\u2028\u2029\n\x01"})
	if e != nil || string(b) != "{\"text\":\"<>&/\u2028\u2029\\n\\u0001\"}" {
		t.Fatalf("%s %v", b, e)
	}
}
func FuzzDecode(f *testing.F) {
	f.Add(`{"a":1}`)
	f.Fuzz(func(t *testing.T, s string) {
		var v any
		if Decode([]byte(s), &v) == nil {
			b, e := Marshal(v)
			if e != nil {
				t.Fatal(e)
			}
			var w any
			if e = Decode(b, &w); e != nil {
				t.Fatal(e)
			}
		}
	})
}
func TestMarshalRejectLossy(t *testing.T) {
	for _, v := range []any{1.0, string([]byte{255}), map[string]any{"x": 2.0}} {
		if _, e := Marshal(v); e == nil {
			t.Errorf("accepted lossy %v", v)
		}
	}
}

// The prior implementation is retained as an independent byte/error oracle.
func legacyMarshal(v any) ([]byte, error) {
	if e := domain(reflect.ValueOf(v), 0); e != nil {
		return nil, e
	}
	raw, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	var x any
	if e = Decode(raw, &x); e != nil {
		return nil, e
	}
	var b bytes.Buffer
	encode(&b, x)
	return b.Bytes(), nil
}
func TestMarshalLegacyEquivalence(t *testing.T) {
	values := []any{nil, true, 0, -1, int64(9007199254740991), int64(9007199254740992), 1.2, json.Number("-0"), json.Number("1e2"), "<>&\u2028\n", map[string]any{"z": []any{nil, 1, "two"}, "a": false}, map[string]int{"é": 1}, json.RawMessage(`{"a":1,"a":2}`), json.RawMessage(`"\ud800"`), make([]int, MaxArray+1), strings.Repeat("x", MaxBytes)}
	for i, v := range values {
		got, err := Marshal(v)
		want, oldErr := legacyMarshal(v)
		if (err == nil) != (oldErr == nil) || !bytes.Equal(got, want) {
			t.Fatalf("case %d differs: errors %v / %v", i, err, oldErr)
		}
	}
}
func FuzzMarshalLegacyEquivalence(f *testing.F) {
	f.Add(`{"a":[1,null,"x"],"b":true}`)
	f.Fuzz(func(t *testing.T, raw string) {
		v := json.RawMessage(raw)
		got, e := Marshal(v)
		want, old := legacyMarshal(v)
		if (e == nil) != (old == nil) || !bytes.Equal(got, want) {
			t.Fatal("legacy canonical mismatch")
		}
	})
}

func TestQuoteByteEquivalence(t *testing.T) {
	s := "é🎲\u2028\u2029"
	for i := 0; i < 128; i++ {
		s += string(rune(i))
	}
	var got, want bytes.Buffer
	quote(&got, s)
	want.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"', '\\':
			want.WriteByte('\\')
			want.WriteRune(r)
		case '\b':
			want.WriteString(`\b`)
		case '\t':
			want.WriteString(`\t`)
		case '\n':
			want.WriteString(`\n`)
		case '\f':
			want.WriteString(`\f`)
		case '\r':
			want.WriteString(`\r`)
		default:
			if r < 32 {
				fmt.Fprintf(&want, `\u%04x`, r)
			} else {
				want.WriteRune(r)
			}
		}
	}
	want.WriteByte('"')
	if !bytes.Equal(got.Bytes(), want.Bytes()) {
		t.Fatal("quote differs from legacy byte encoding")
	}
}

func TestExplicitCanonicalByteLimit(t *testing.T) {
	value := strings.Repeat("x", MaxBytes)
	if _, e := Marshal(value); e == nil {
		t.Fatal("default limit widened")
	}
	raw, e := MarshalLimit(value, 8<<20)
	if e != nil {
		t.Fatal(e)
	}
	var decoded string
	if e = DecodeLimit(raw, &decoded, 8<<20); e != nil || decoded != value {
		t.Fatal("extended roundtrip", e)
	}
	if Decode(raw, &decoded) == nil {
		t.Fatal("default decode widened")
	}
	for _, invalid := range [][]byte{[]byte(`{"x":1,"x":2}`), []byte(`1e2`), []byte(`"\ud800"`), []byte(`9007199254740992`)} {
		var x any
		if DecodeLimit(invalid, &x, 8<<20) == nil {
			t.Fatal("extended limit relaxed canonical validation")
		}
	}
}
