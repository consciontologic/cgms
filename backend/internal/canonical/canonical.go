// Package canonical implements cgms-canonical-v1. Domain JSON contains integers only.
package canonical

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const Version = "cgms-canonical-v1"
const MaxBytes = 4 << 20
const MaxDepth = 64
const MaxArray = 100000

// DecodeLimit changes only the byte ceiling; all canonical validation is retained.
// Callers must explicitly bind an extended limit to a versioned artifact contract.
func DecodeLimit(b []byte, v any, maxBytes int) error {
	if maxBytes < 1 || maxBytes > 64<<20 {
		return fmt.Errorf("invalid canonical byte limit")
	}
	if len(b) > maxBytes || !utf8.Valid(b) {
		return fmt.Errorf("JSON size or UTF-8 invalid")
	}
	if e := surrogates(b); e != nil {
		return e
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	x, e := value(d, 0)
	if e != nil {
		return e
	}
	if _, e = d.Token(); e != io.EOF {
		return fmt.Errorf("trailing JSON")
	}
	if e = shape(x, reflect.TypeOf(v)); e != nil {
		return e
	}
	raw, e := json.Marshal(x)
	if e != nil {
		return e
	}
	dd := json.NewDecoder(bytes.NewReader(raw))
	dd.UseNumber()
	dd.DisallowUnknownFields()
	return dd.Decode(v)
}
func Decode(b []byte, v any) error { return DecodeLimit(b, v, MaxBytes) }

func value(d *json.Decoder, depth int) (any, error) {
	if depth > MaxDepth {
		return nil, fmt.Errorf("depth exceeded")
	}
	t, e := d.Token()
	if e != nil {
		return nil, e
	}
	switch x := t.(type) {
	case json.Delim:
		switch x {
		case '{':
			m := map[string]any{}
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return nil, e
				}
				s, ok := k.(string)
				if !ok {
					return nil, fmt.Errorf("key")
				}
				for _, c := range s {
					if c > 127 {
						return nil, fmt.Errorf("non ASCII key")
					}
				}
				if _, ok = m[s]; ok {
					return nil, fmt.Errorf("duplicate %s", s)
				}
				v, e := value(d, depth+1)
				if e != nil {
					return nil, e
				}
				m[s] = v
				if len(m) > MaxArray {
					return nil, fmt.Errorf("object size")
				}
			}
			_, e = d.Token()
			return m, e
		case '[':
			a := []any{}
			for d.More() {
				v, e := value(d, depth+1)
				if e != nil {
					return nil, e
				}
				a = append(a, v)
				if len(a) > MaxArray {
					return nil, fmt.Errorf("array size")
				}
			}
			_, e = d.Token()
			return a, e
		}
		return nil, fmt.Errorf("delimiter")
	case json.Number:
		n, e := strconv.ParseInt(string(x), 10, 64)
		if e != nil || n > 9007199254740991 || n < -9007199254740991 || strings.ContainsAny(string(x), ".eE") {
			return nil, fmt.Errorf("integer outside domain")
		}
		return x, nil
	}
	return t, nil
}
func shape(x any, t reflect.Type) error {
	if t == nil {
		return fmt.Errorf("nil target")
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if reflect.PointerTo(t).Implements(reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()) {
		return nil
	}
	if x == nil {
		return nil
	}
	switch t.Kind() {
	case reflect.Struct:
		m, ok := x.(map[string]any)
		if !ok {
			return nil
		}
		fields := map[string]reflect.Type{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = f.Name
			}
			fields[name] = f.Type
		}
		for k, v := range m {
			ft, ok := fields[k]
			if !ok {
				return fmt.Errorf("unknown exact-case field %s", k)
			}
			if e := shape(v, ft); e != nil {
				return e
			}
		}
	case reflect.Slice, reflect.Array:
		if a, ok := x.([]any); ok {
			for _, v := range a {
				if e := shape(v, t.Elem()); e != nil {
					return e
				}
			}
		}
	case reflect.Map:
		if m, ok := x.(map[string]any); ok {
			for _, v := range m {
				if e := shape(v, t.Elem()); e != nil {
					return e
				}
			}
		}
	}
	return nil
}
func surrogates(b []byte) error {
	for i := 0; i < len(b); i++ {
		if b[i] != '\\' {
			continue
		}
		i++
		if i >= len(b) || b[i] != 'u' {
			continue
		}
		if i+4 >= len(b) {
			return fmt.Errorf("escape")
		}
		n, e := strconv.ParseUint(string(b[i+1:i+5]), 16, 16)
		if e != nil {
			return e
		}
		i += 4
		if n >= 0xdc00 && n <= 0xdfff {
			return fmt.Errorf("unpaired low surrogate")
		}
		if n >= 0xd800 && n <= 0xdbff {
			if i+6 >= len(b) || b[i+1] != '\\' || b[i+2] != 'u' {
				return fmt.Errorf("unpaired high surrogate")
			}
			n, e = strconv.ParseUint(string(b[i+3:i+7]), 16, 16)
			if e != nil || n < 0xdc00 || n > 0xdfff {
				return fmt.Errorf("surrogate pair")
			}
			i += 6
		}
	}
	return nil
}
func Marshal(v any) ([]byte, error) { return MarshalLimit(v, MaxBytes) }

// MarshalLimit preserves canonical bytes and domain checks under an explicit cap.
func MarshalLimit(v any, maxBytes int) ([]byte, error) {
	if maxBytes < 1 || maxBytes > 64<<20 {
		return nil, fmt.Errorf("invalid canonical byte limit")
	}
	if e := domain(reflect.ValueOf(v), 0); e != nil {
		return nil, e
	}
	raw, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	// Marshal needs only an untyped parse tree; Decode's typed-target roundtrip
	// would serialize and parse that tree again without adding validation.
	if len(raw) > maxBytes || !utf8.Valid(raw) {
		return nil, fmt.Errorf("JSON size or UTF-8 invalid")
	}
	if e = surrogates(raw); e != nil {
		return nil, e
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	x, e := value(d, 0)
	if e != nil {
		return nil, e
	}
	if _, e = d.Token(); e != io.EOF {
		return nil, fmt.Errorf("trailing JSON")
	}
	var b bytes.Buffer
	encode(&b, x)
	return b.Bytes(), nil
}
func encode(b *bytes.Buffer, v any) {
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if x {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case json.Number:
		if x == "-0" {
			b.WriteByte('0')
		} else {
			b.WriteString(string(x))
		}
	case string:
		quote(b, x)
	case []any:
		b.WriteByte('[')
		for i, v := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			encode(b, v)
		}
		b.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			quote(b, k)
			b.WriteByte(':')
			encode(b, x[k])
		}
		b.WriteByte('}')
	}
}
func quote(b *bytes.Buffer, s string) {
	b.WriteByte('"')
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 32 && c != '"' && c != '\\' {
			continue
		}
		b.WriteString(s[start:i])
		switch c {
		case '"', '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case '\b':
			b.WriteString(`\b`)
		case '\t':
			b.WriteString(`\t`)
		case '\n':
			b.WriteString(`\n`)
		case '\f':
			b.WriteString(`\f`)
		case '\r':
			b.WriteString(`\r`)
		default:
			const hexDigits = "0123456789abcdef"
			b.WriteString(`\u00`)
			b.WriteByte(hexDigits[c>>4])
			b.WriteByte(hexDigits[c&15])
		}
		start = i + 1
	}
	b.WriteString(s[start:])
	b.WriteByte('"')
}

func Hash(v any) (string, error) {
	b, e := Marshal(v)
	if e != nil {
		return "", e
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func domain(v reflect.Value, depth int) error {
	if depth > MaxDepth {
		return fmt.Errorf("depth exceeded")
	}
	if !v.IsValid() {
		return nil
	}
	switch v.Kind() {
	case reflect.Interface, reflect.Pointer:
		if !v.IsNil() {
			return domain(v.Elem(), depth+1)
		}
	case reflect.Float32, reflect.Float64:
		return fmt.Errorf("float forbidden")
	case reflect.String:
		if !utf8.ValidString(v.String()) {
			return fmt.Errorf("UTF-8 invalid")
		}
	case reflect.Map:
		it := v.MapRange()
		for it.Next() {
			if e := domain(it.Key(), depth+1); e != nil {
				return e
			}
			if e := domain(it.Value(), depth+1); e != nil {
				return e
			}
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if e := domain(v.Index(i), depth+1); e != nil {
				return e
			}
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).PkgPath == "" {
				if e := domain(v.Field(i), depth+1); e != nil {
					return e
				}
			}
		}
	}
	return nil
}
