package transport

import (
	"encoding/json"
	"os"
	"testing"
)

func TestEconomyAPISecurityReferencesAndReceiptSchema(t *testing.T) {
	b, err := os.ReadFile("../../../docs/api/online.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Paths map[string]map[string]struct {
			Security []map[string]json.RawMessage `json:"security"`
		} `json:"paths"`
		Components struct {
			SecuritySchemes map[string]json.RawMessage `json:"securitySchemes"`
			Schemas         map[string]struct {
				Required   []string `json:"required"`
				Additional bool     `json:"additionalProperties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err = json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	for path, method := range map[string]string{"/v1/economy": "get", "/v1/economy/passes": "post", "/v1/economy/passes/{operation}": "get"} {
		op, ok := doc.Paths[path][method]
		if !ok || len(op.Security) != 2 {
			t.Fatal("missing authenticated economy route", path)
		}
		for _, alternative := range op.Security {
			for name := range alternative {
				if _, ok := doc.Components.SecuritySchemes[name]; !ok {
					t.Fatal("undefined authentication scheme", path, name)
				}
			}
		}
	}
	for name, count := range map[string]int{"PassPurchase": 2, "PassReceipt": 4} {
		s, ok := doc.Components.Schemas[name]
		if !ok || s.Additional || len(s.Required) != count {
			t.Fatal("incomplete immutable pass contract", name)
		}
	}
}
