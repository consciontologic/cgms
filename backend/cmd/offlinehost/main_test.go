package main

import (
	"bytes"
	"encoding/json"
	"github.com/metaphy6/cgms/backend/internal/offline"
	"strings"
	"testing"
)

func TestProtocolRejectsPrivateImportAndProjects(t *testing.T) {
	store, e := offline.OpenStore(t.TempDir(), "rules")
	if e != nil {
		t.Fatal(e)
	}
	defer store.Close()
	input := strings.NewReader("{\"kind\":\"new\",\"slot\":\"one\",\"population\":3,\"games\":1,\"difficulty\":\"beginner\"}\n{\"kind\":\"load\",\"slot\":\"one\",\"engine\":{}}\n")
	var output bytes.Buffer
	if e = serve(input, &output, offline.NewHost(store)); e != nil {
		t.Fatal(e)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Fatal("response count")
	}
	var first, second map[string]any
	if json.Unmarshal([]byte(lines[0]), &first) != nil || json.Unmarshal([]byte(lines[1]), &second) != nil {
		t.Fatal("JSON")
	}
	if first["view"] == nil || second["error"] != "invalid_request" {
		t.Fatal("protocol admission")
	}
	if strings.Contains(lines[0], "start_ledger") || strings.Contains(lines[0], "\"chance\"") {
		t.Fatal("private save escaped")
	}
}
