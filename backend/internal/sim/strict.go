package sim

import (
	"encoding/json"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"reflect"
	"strings"
)

func decodeRequired(b []byte, v any) error {
	limit := canonical.MaxBytes
	switch v.(type) {
	case *GameplayRun:
		limit = 32 << 20
	case *Scenario:
		limit = (32 << 20) + 65536
	}
	if e := canonical.DecodeLimit(b, v, limit); e != nil {
		return e
	}
	var raw any
	if e := canonical.DecodeLimit(b, &raw, limit); e != nil {
		return e
	}
	if e := requiredValue(raw, reflect.TypeOf(v)); e != nil {
		return e
	}
	return validateExtendedArtifact(b, v)
}
func requiredValue(v any, t reflect.Type) error {
	for t.Kind() == reflect.Pointer {
		if v == nil {
			return nil
		}
		t = t.Elem()
	}
	if reflect.PointerTo(t).Implements(reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()) {
		return nil
	}
	if v == nil {
		if t.Kind() == reflect.Slice || t.Kind() == reflect.Map || t.Kind() == reflect.Interface {
			return nil
		}
		return fmt.Errorf("null scalar")
	}
	switch t.Kind() {
	case reflect.Struct:
		m, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("object required")
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			tags := strings.Split(f.Tag.Get("json"), ",")
			name := tags[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = f.Name
			}
			x, ok := m[name]
			if !ok {
				if len(tags) > 1 && tags[1] == "omitempty" {
					continue
				}
				return fmt.Errorf("missing required field %s", name)
			}
			if e := requiredValue(x, f.Type); e != nil {
				return e
			}
		}
	case reflect.Slice, reflect.Array:
		a, ok := v.([]any)
		if !ok {
			return fmt.Errorf("array required")
		}
		for _, x := range a {
			if e := requiredValue(x, t.Elem()); e != nil {
				return e
			}
		}
	case reflect.Map:
		m, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("object required")
		}
		for _, x := range m {
			if e := requiredValue(x, t.Elem()); e != nil {
				return e
			}
		}
	}
	return nil
}
