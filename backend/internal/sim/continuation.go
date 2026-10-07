package sim

import (
	"bytes"
	"context"
	"fmt"
	"strings"
)

func verifyContinuationParent(ctx context.Context, s Scenario, config string) (map[string][]byte, error) {
	m, f, e := readManifestContext(ctx, s.Parent)
	if e != nil {
		return nil, e
	}
	if m.ConfigHash != config {
		return nil, ErrCompatibility
	}
	if e = verifyPrefixContext(ctx, m, f); e != nil {
		return nil, e
	}
	checkpoint := mustCanonical(s.Checkpoint)
	suffix := "/checkpoint.json"
	if s.Gameplay != nil {
		checkpoint = mustCanonical(s.Gameplay)
		suffix = "/gameplay.json"
	}
	found := false
	for name, b := range f {
		if strings.HasSuffix(name, suffix) && bytes.Equal(checkpoint, b) {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("checkpoint not recorded in parent")
	}
	snapshots := map[string][]byte{"parent-manifest.json": mustCanonical(m), "parent-checkpoint.json": checkpoint}
	if s.Gameplay != nil {
		var limits Budgets
		if err := decodeRequired(f["execution-limits.json"], &limits); err != nil {
			return nil, err
		}
		snapshots["parent-execution-limits.json"] = f["execution-limits.json"]
	}
	return snapshots, nil
}
func verifyParentSnapshot(m Manifest, f map[string][]byte) error {
	if f["scenario.json"] == nil {
		return nil
	}
	var sc Scenario
	if e := decodeRequired(f["scenario.json"], &sc); e != nil {
		return e
	}
	if sc.Kind != "settlement-checkpoint" && sc.Kind != "gameplay-checkpoint" {
		return nil
	}
	var p Manifest
	if e := decodeRequired(f["parent-manifest.json"], &p); e != nil {
		return e
	}
	if p.ConfigHash != m.ConfigHash || p.Provenance.Engine != m.Provenance.Engine {
		return ErrCompatibility
	}
	if sc.Gameplay != nil {
		bound := false
		for _, a := range p.Artifacts {
			if a.Path == "execution-limits.json" && a.SHA256 == checksum(f["parent-execution-limits.json"]) && a.Bytes == len(f["parent-execution-limits.json"]) {
				bound = true
			}
		}
		if !bound {
			return fmt.Errorf("parent execution limits provenance")
		}
		var prior, current Budgets
		if e := decodeRequired(f["parent-execution-limits.json"], &prior); e != nil {
			return e
		}
		if e := decodeRequired(f["execution-limits.json"], &current); e != nil {
			return e
		}
		prior.MaxTransitions, prior.WallSeconds, prior.Workers = current.MaxTransitions, current.WallSeconds, current.Workers
		if hash(prior) != hash(current) {
			return fmt.Errorf("continuation policy or resource budget drift")
		}
	}
	checkpoint := mustCanonical(sc.Checkpoint)
	suffix := "/checkpoint.json"
	if sc.Gameplay != nil {
		checkpoint = mustCanonical(sc.Gameplay)
		suffix = "/gameplay.json"
	}
	if !bytes.Equal(checkpoint, f["parent-checkpoint.json"]) {
		return fmt.Errorf("parent checkpoint snapshot mismatch")
	}
	for _, a := range p.Artifacts {
		if strings.HasSuffix(a.Path, suffix) && a.SHA256 == checksum(checkpoint) && a.Bytes == len(checkpoint) {
			return nil
		}
	}
	return fmt.Errorf("parent checkpoint provenance")
}
