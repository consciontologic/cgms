package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Future serialized fields must update the checked interchange contract in the
// same change. This deliberately checks generated artifacts, not engine behavior.
func TestCommittedRuntimeSchemasMatchTypes(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--out", t.TempDir())
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off")
	output, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("generator: %v %s", e, output)
	}
	generated := cmd.Args[len(cmd.Args)-1]
	files, e := os.ReadDir(generated)
	if e != nil {
		t.Fatal(e)
	}
	if len(files) != 10 {
		t.Fatal("missing runtime schema family")
	}
	for _, file := range files {
		got, e := os.ReadFile(filepath.Join(generated, file.Name()))
		if e != nil {
			t.Fatal(e)
		}
		want, e := os.ReadFile(filepath.Join("../../..", "config/schemas", file.Name()))
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("stale schema %s; run python3 xops/sim_schema.py", file.Name())
		}
	}
}
