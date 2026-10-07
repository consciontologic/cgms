// sim-schema generates closed runtime JSON shapes from the serialized Go types.
// Semantic conservation, compatibility and accounting validators remain in Go.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/sim"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type object = map[string]any
type generator struct{ defs object }

func (g *generator) shape(t reflect.Type) any {
	if t.Kind() == reflect.Pointer {
		return object{"anyOf": []any{g.shape(t.Elem()), object{"type": "null"}}}
	}
	switch t.Kind() {
	case reflect.Struct:
		name := t.Name()
		if name == "" {
			panic("anonymous runtime structure")
		}
		ref := object{"$ref": "#/$defs/" + name}
		if _, ok := g.defs[name]; ok {
			return ref
		}
		g.defs[name] = object{}
		if t == reflect.TypeOf(game.Amount{}) {
			g.defs[name] = object{"type": "object", "additionalProperties": false, "description": "Reduced exact rational; canonicality and domain sign validated by Go.", "properties": object{"numerator": object{"type": "string", "pattern": "^(0|-[1-9][0-9]*|[1-9][0-9]*)$"}, "denominator": object{"type": "string", "pattern": "^[1-9][0-9]*$"}}, "required": []string{"numerator", "denominator"}}
			return ref
		}
		props := object{}
		required := []string{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			tag := strings.Split(f.Tag.Get("json"), ",")
			if tag[0] == "-" {
				continue
			}
			if tag[0] == "" {
				panic("missing JSON field tag: " + name + "." + f.Name)
			}
			props[tag[0]] = g.shape(f.Type)
			optional := false
			for _, part := range tag[1:] {
				if part == "omitempty" {
					optional = true
				}
			}
			if !optional {
				required = append(required, tag[0])
			}
		}
		overrides(name, props)
		g.defs[name] = object{"type": "object", "additionalProperties": false, "properties": props, "required": required}
		return ref
	case reflect.Slice:
		return object{"type": []string{"array", "null"}, "items": g.shape(t.Elem()), "maxItems": 100000}
	case reflect.Array:
		return object{"type": "array", "items": g.shape(t.Elem()), "minItems": t.Len(), "maxItems": t.Len()}
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			panic("non-string JSON map key")
		}
		return object{"type": []string{"object", "null"}, "additionalProperties": g.shape(t.Elem()), "propertyNames": object{"pattern": "^[\\x00-\\x7f]*$"}}
	case reflect.String:
		return object{"type": "string"}
	case reflect.Bool:
		return object{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return object{"type": "integer", "minimum": -9007199254740991, "maximum": 9007199254740991}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return object{"type": "integer", "minimum": 0, "maximum": 9007199254740991}
	default:
		panic("unsupported runtime schema type " + t.String())
	}
}
func overrides(name string, p object) {
	versions := map[string]string{"Metrics": "cgms-metrics-v1", "Manifest": "cgms-manifest-v1", "Outcome": "cgms-outcome-v1", "Trace": "cgms-trace-v1", "Scenario": "cgms-scenario-v1"}
	if v, ok := versions[name]; ok {
		p["schema_version"] = object{"const": v}
	}
	switch name {
	case "GameplayPublicStats":
		p["schema"] = object{"const": "cgms-gameplay-public-v2"}
	case "GameplayPrivateFinance":
		p["schema"] = object{"const": "cgms-gameplay-finance-private-v2"}
		p["privacy"] = object{"const": "restricted"}
		p["observer"] = object{"const": "omniscient"}
	case "Manifest":
		p["exit_code"] = object{"enum": []int{0, 2, 3, 4, 5, 6, 130}}
		p["privacy"] = object{"const": "restricted"}
	case "Outcome":
		p["population"] = object{"enum": []int{3, 4}}
	case "Trace":
		p["kind"] = object{"enum": []string{"command", "deal", "settlement", "gameplay-v1"}}
	case "Scenario":
		p["kind"] = object{"enum": []string{"accounting-only", "transition-sequence", "settlement-checkpoint", "gameplay-checkpoint"}}
		p["players"] = object{"enum": []int{3, 4}}
	case "SettlementCursor":
		p["version"] = object{"const": "cgms-fifo-v1"}
	case "State":
		p["schema"] = object{"const": "cgms-state-v1"}
		p["cards"] = object{"type": "array", "items": object{"$ref": "#/$defs/PlacedCard"}, "minItems": 104, "maxItems": 104}
		p["players"] = object{"type": "array", "items": object{"$ref": "#/$defs/Player"}, "minItems": 3, "maxItems": 4}
	}
}
func main() {
	out := flag.String("out", "../config/schemas", "schema output directory")
	check := flag.Bool("check", false, "refuse stale schemas without writing")
	flag.Parse()
	specs := []struct {
		name  string
		types []any
	}{{"gameplay-public", []any{sim.GameplayPublicStats{}}}, {"finance-private", []any{sim.GameplayPrivateFinance{}}}, {"gameplay", []any{sim.GameplayRun{}}}, {"metrics", []any{sim.Metrics{}}}, {"manifest", []any{sim.Manifest{}}}, {"outcome", []any{sim.Outcome{}}}, {"trace", []any{sim.Trace{}}}, {"scenario", []any{sim.Scenario{}}}, {"checkpoint", []any{game.SettlementCursor{}, game.State{}, game.MatchLifecycle{}}}, {"observation", []any{game.Observation{}}}}
	for _, spec := range specs {
		g := generator{defs: object{}}
		refs := []any{}
		for _, v := range spec.types {
			refs = append(refs, g.shape(reflect.TypeOf(v)))
		}
		var root any = refs[0]
		if len(refs) > 1 {
			root = object{"oneOf": refs}
		}
		schema := object{"$schema": "https://json-schema.org/draft/2020-12/schema", "title": "CGMS " + spec.name + " v1", "description": "Closed offline runtime shape. Go additionally validates physical conservation, lineage, compatibility and exact rational canonicality.", "allOf": []any{root}, "$defs": g.defs}
		b, e := json.MarshalIndent(schema, "", "  ")
		if e != nil {
			panic(e)
		}
		b = append(b, '\n')
		path := filepath.Join(*out, spec.name+".schema.json")
		if *check {
			old, e := os.ReadFile(path)
			if e != nil || !bytes.Equal(old, b) {
				fmt.Fprintln(os.Stderr, "stale runtime schema:", spec.name)
				os.Exit(1)
			}
		} else {
			if e = os.WriteFile(path, b, 0644); e != nil {
				panic(e)
			}
		}
	}
	fmt.Println("Runtime schemas verified")
}
