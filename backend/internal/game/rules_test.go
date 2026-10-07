package game

import (
	"encoding/json"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"os"
	"testing"
)

func TestRulesNormalization(t *testing.T) {
	b, e := os.ReadFile("../../../config/rules/game-rules.json")
	if e != nil {
		t.Fatal(e)
	}
	a, e := LoadRules(b, nil)
	if e != nil {
		t.Fatal(e)
	}
	c, e := LoadRules(b, []byte(`{"match_games":3}`))
	if e != nil {
		t.Fatal(e)
	}
	ha, _ := canonical.Hash(a)
	hc, _ := canonical.Hash(c)
	if ha != hc {
		t.Fatal("normalization differs")
	}
	for _, s := range []string{`{"unknown":1}`, `{"rounds":12}`, `{"match_games":0}`, `{"schema_version":"future"}`, `{"Match_games":3}`, `{"match_games":null}`} {
		if _, e = LoadRules(b, []byte(s)); e == nil {
			t.Errorf("accepted %s", s)
		}
	}
}

func TestRulesFileDrivenGolden(t *testing.T) {
	base, e := os.ReadFile("../../../config/rules/game-rules.json")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("../../testdata/config/normalization.json")
	if e != nil {
		t.Fatal(e)
	}
	var vectors struct {
		Canonicalization string            `json:"canonicalization"`
		Hash             string            `json:"canonical_defaults_sha256"`
		Equivalent       []json.RawMessage `json:"equivalent_overrides"`
		Rejected         []json.RawMessage `json:"rejected_overrides"`
		Defaults         []string          `json:"default_encodings"`
		Changed          []struct {
			Override json.RawMessage `json:"override"`
			Hash     string          `json:"sha256"`
		} `json:"changed_overrides"`
		Malformed []string `json:"rejected_json"`
	}
	if e = canonical.Decode(raw, &vectors); e != nil {
		t.Fatal(e)
	}
	if vectors.Canonicalization != canonical.Version {
		t.Fatal("vector version")
	}
	encodings := append([]string{string(base)}, vectors.Defaults...)
	for _, encoding := range encodings {
		for _, override := range vectors.Equivalent {
			r, e := LoadRules([]byte(encoding), override)
			if e != nil {
				t.Fatal(e)
			}
			h, e := canonical.Hash(r)
			if e != nil || h != vectors.Hash {
				t.Fatalf("equivalent encoding hash=%s error=%v", h, e)
			}
		}
	}
	for _, v := range vectors.Changed {
		r, e := LoadRules(base, v.Override)
		if e != nil {
			t.Fatal(e)
		}
		h, e := canonical.Hash(r)
		if e != nil || h != v.Hash || h == vectors.Hash {
			t.Fatalf("changed config hash=%s error=%v", h, e)
		}
	}
	for _, v := range vectors.Rejected {
		if _, e = LoadRules(base, v); e == nil {
			t.Errorf("accepted rejected fixture %s", v)
		}
	}
	for _, v := range vectors.Malformed {
		if _, e = LoadRules(base, []byte(v)); e == nil {
			t.Errorf("accepted malformed/nested fixture %s", v)
		}
	}
}
