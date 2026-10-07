package sim

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResourceRationalDigits(t *testing.T) {
	for _, raw := range []string{`{"a":{"numerator":"123","denominator":"1"}}`, `{"numerator":"-123","denominator":"1"}`, `{"numerator":"1","denominator":"123"}`, `{"numerator":123,"denominator":"1"}`, `{"numerator":"01","denominator":"1"}`, `{"numerator":"1","denominator":"0"}`, `{"numerator":"1","numerator":"2"}`} {
		if checkRationalDigits([]byte(raw), 2) == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	if e := checkRationalDigits([]byte(`{"a":{"numerator":"-12","denominator":"99"},"note":"123456"}`), 2); e != nil {
		t.Fatal(e)
	}
}
func TestResourceArtifactBoundaries(t *testing.T) {
	p := filepath.Join(t.TempDir(), "artifact")
	if e := os.WriteFile(p, []byte("1234"), 0600); e != nil {
		t.Fatal(e)
	}
	if b, e := readArtifact(context.Background(), p, 4); e != nil || string(b) != "1234" {
		t.Fatal(string(b), e)
	}
	if _, e := readArtifact(context.Background(), p, 3); !errors.Is(e, ErrResourceBudget) {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := readArtifact(ctx, p, 4); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e := readArtifact(context.Background(), filepath.Dir(p), 4); e == nil {
		t.Fatal("directory accepted")
	}
}
func TestResourceSerializationBudget(t *testing.T) {
	if e := serializationBudget(4, []byte("12"), []byte("34")); e != nil {
		t.Fatal(e)
	}
	if e := serializationBudget(3, []byte("12"), []byte("34")); !errors.Is(e, ErrResourceBudget) {
		t.Fatal(e)
	}
	if e := serializationBudget(-1); e == nil {
		t.Fatal("negative budget")
	}
}
