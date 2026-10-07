package game

import (
	"encoding/json"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
)

// Rules is an effective immutable-by-convention config; loaders never alias inputs.
type Rules map[string]any

const AcceptedRulesHash = "e1cd154e5968bc701b0dcbd206098be177e07b8c4c14df55fcf86b4e6801afa6"

// LoadRules verifies the canonical defaults before merging a strict partial override.
// The pinned content digest is the compatibility allow-list; values are read only
// from the canonical JSON, never duplicated in engine or CLI defaults.
func LoadRules(defaultBytes, overrideBytes []byte) (Rules, error) {
	var base Rules
	if e := canonical.Decode(defaultBytes, &base); e != nil {
		return nil, e
	}
	h, e := canonical.Hash(base)
	if e != nil || h != AcceptedRulesHash {
		return nil, fmt.Errorf("unsupported canonical rule defaults")
	}
	if len(overrideBytes) == 0 {
		return base, nil
	}
	var over Rules
	if e = canonical.Decode(overrideBytes, &over); e != nil {
		return nil, e
	}
	if over == nil {
		return nil, fmt.Errorf("rules override must be object")
	}
	for k, v := range over {
		old, ok := base[k]
		if !ok {
			return nil, fmt.Errorf("unknown rule %s", k)
		}
		if k == "match_games" {
			n, ok := v.(json.Number)
			if !ok {
				return nil, fmt.Errorf("match_games integer required")
			}
			i, e := n.Int64()
			if e != nil || i < 1 {
				return nil, fmt.Errorf("match_games must be positive")
			}
		} else {
			a, e := canonical.Hash(old)
			if e != nil {
				return nil, e
			}
			b, e := canonical.Hash(v)
			if e != nil || a != b {
				return nil, fmt.Errorf("accepted constant %s cannot be replaced", k)
			}
		}
		base[k] = v
	}
	return base, nil
}
