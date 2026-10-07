package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestServiceSchemaCoversRuntimeSettings(t *testing.T) {
	raw, err := os.ReadFile("../../../config/schemas/service.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties map[string]struct {
			Type string `json:"type"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	kinds := map[reflect.Kind]string{reflect.Bool: "boolean", reflect.String: "string", reflect.Int: "integer", reflect.Struct: "object"}
	typ := reflect.TypeFor[Config]()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		property, ok := schema.Properties[name]
		if !ok {
			t.Errorf("runtime setting %s is rejected by the published schema", name)
			continue
		}
		// Version is constrained by its constant rather than a type keyword.
		if name != "version" && property.Type != kinds[field.Type.Kind()] {
			t.Errorf("setting %s schema type %q differs from runtime %v", name, property.Type, field.Type)
		}
	}
}

func TestStrictAndSecrets(t *testing.T) {
	secret := filepath.Join(t.TempDir(), "database")
	if e := os.WriteFile(secret, []byte("postgres://user:private@postgres/cgms"), 0600); e != nil {
		t.Fatal(e)
	}
	base := `{"version":1,"listen":"127.0.0.1:8080","origin":"http://localhost:8080","insecure_local":true,"rules_file":"rules.json","database_url_file":"` + secret + `","observability":{"enabled":false,"endpoint":""}}`
	c, e := Decode(strings.NewReader(base))
	if e != nil {
		t.Fatal(e)
	}
	secrets, e := c.Secrets()
	if e != nil || !strings.Contains(secrets.DatabaseURL, "private") {
		t.Fatal("secret not loaded")
	}
	for _, bad := range []string{strings.Replace(base, `"version":1`, `"version":2`, 1), strings.Replace(base, `"version":1`, `"unknown":1,"version":1`, 1), strings.Replace(base, `"insecure_local":true`, `"insecure_local":false`, 1), base + `{}`, strings.Replace(base, `"enabled":false`, `"enabled":true`, 1)} {
		if _, e := Decode(strings.NewReader(bad)); e == nil {
			t.Fatal("invalid accepted")
		}
	}
	if e := os.Chmod(secret, 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := c.Secrets(); e == nil {
		t.Fatal("public secret accepted")
	}
}

func TestMissingNullDuplicateAndInvalidPort(t *testing.T) {
	base := `{"version":1,"listen":"127.0.0.1:8080","origin":"https://localhost","insecure_local":false,"rules_file":"rules.json","database_url_file":"secret","observability":{"enabled":false,"endpoint":""}}`
	for _, bad := range []string{strings.Replace(base, `"enabled":false`, `"enabled":null`, 1), strings.Replace(base, `"enabled":false,`, ``, 1), strings.Replace(base, `"version":1`, `"version":1,"version":1`, 1), strings.Replace(base, `:8080`, `:99999`, 1), strings.Replace(base, `"insecure_local":false`, `"insecure_local":null`, 1), strings.Replace(base, `"observability":{`, `"Observability":{`, 1)} {
		if _, e := Decode(strings.NewReader(bad)); e == nil {
			t.Fatal("invalid service config accepted")
		}
	}
}

func TestBoundedStreamCapacity(t *testing.T) {
	base := `{"version":1,"listen":"127.0.0.1:8080","origin":"https://localhost","insecure_local":false,"rules_file":"rules.json","database_url_file":"secret","observability":{"enabled":false,"endpoint":""}}`
	c, err := Decode(strings.NewReader(base))
	if err != nil || c.MaxStreams != 64 {
		t.Fatal("default capacity", err)
	}
	c, err = Decode(strings.NewReader(strings.Replace(base, `"version":1`, `"version":1,"max_streams":400`, 1)))
	if err != nil || c.MaxStreams != 400 {
		t.Fatal("explicit capacity", err)
	}
	for _, value := range []string{"0", "-1", "4097", "null", "1.5"} {
		if _, err = Decode(strings.NewReader(strings.Replace(base, `"version":1`, `"version":1,"max_streams":`+value, 1))); err == nil {
			t.Fatal("unbounded/invalid capacity accepted", value)
		}
	}
}

func TestOptionalSettingsRejectExplicitNull(t *testing.T) {
	base := `{"version":1,"listen":"127.0.0.1:8080","origin":"https://localhost","insecure_local":false,"rules_file":"rules.json","database_url_file":"secret","observability":{"enabled":false,"endpoint":""}}`
	for _, name := range []string{"proxy_client_ip", "redis_url_file"} {
		bad := strings.Replace(base, `"version":1`, `"version":1,"`+name+`":null`, 1)
		if _, e := Decode(strings.NewReader(bad)); e == nil {
			t.Fatal("explicit null accepted", name)
		}
	}
}

func TestEconomyActivationIsExplicit(t *testing.T) {
	base := `{"version":1,"listen":"127.0.0.1:8080","origin":"http://localhost:8080","insecure_local":true,"rules_file":"rules.json","database_url_file":"secret","observability":{"enabled":false,"endpoint":""}}`
	if _, err := Decode(strings.NewReader(strings.Replace(base, `"version":1`, `"version":1,"economy_enabled":true`, 1))); err != nil {
		t.Fatal("approved economy configuration rejected", err)
	}
	for _, value := range []string{"null", "1", `"true"`} {
		if _, err := Decode(strings.NewReader(strings.Replace(base, `"version":1`, `"version":1,"economy_enabled":`+value, 1))); err == nil {
			t.Fatal("ambiguous activation accepted", value)
		}
	}
}
