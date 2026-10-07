package sim

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

var ErrCompatibility = errors.New("incompatible version")

const EngineVersion = "cgms-engine-v2.2"

var SourceRevision = "development"
var DirtyPatchHash = "unrecorded"

type Artifact struct {
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
	Bytes   int    `json:"bytes"`
	Privacy string `json:"privacy"`
}
type Provenance struct {
	PRNG               string `json:"prng"`
	SeedDerivation     string `json:"seed_derivation"`
	Sampler            string `json:"sampler"`
	Shuffle            string `json:"shuffle"`
	Menu               string `json:"legal_menu"`
	Engine             string `json:"engine"`
	SourceRevision     string `json:"source_revision"`
	DirtyPatchHash     string `json:"dirty_patch_hash"`
	BinarySHA256       string `json:"binary_sha256"`
	GoVersion          string `json:"go_version"`
	Platform           string `json:"platform"`
	BuildInfo          string `json:"build_info"`
	DependencyHash     string `json:"dependency_hash"`
	RuleSourceSHA256   string `json:"rule_source_sha256"`
	RuleSourceRevision string `json:"rule_source_revision"`
}
type Manifest struct {
	Schema         string     `json:"schema_version"`
	Status         string     `json:"status"`
	ExitCode       int        `json:"exit_code"`
	Command        string     `json:"command"`
	Argv           []string   `json:"argv"`
	CWD            string     `json:"cwd"`
	RootSeed       string     `json:"root_seed"`
	ConfigHash     string     `json:"config_hash"`
	Interpretation string     `json:"interpretation"`
	Coverage       []string   `json:"coverage"`
	Exclusions     []string   `json:"exclusions"`
	Provenance     Provenance `json:"provenance"`
	Artifacts      []Artifact `json:"artifacts"`
	Workers        int        `json:"workers"`
	Planned        int        `json:"planned"`
	Started        int        `json:"started"`
	Finalized      int        `json:"finalized"`
	Canceled       int        `json:"canceled"`
	NotStarted     int        `json:"not_started"`
	Parent         string     `json:"parent"`
	Created        string     `json:"created"`
	Expires        string     `json:"expires"`
	ReviewHold     bool       `json:"review_hold"`
	Privacy        string     `json:"privacy"`
}

func checksum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func provenance(ruleBytes []byte) Provenance {
	p := Provenance{PRNG: randomstream.Algorithm, SeedDerivation: randomstream.Derivation, Sampler: randomstream.Sampler, Shuffle: randomstream.ShuffleVersion, Menu: bots.MenuVersion, Engine: EngineVersion, SourceRevision: SourceRevision, DirtyPatchHash: DirtyPatchHash, GoVersion: runtime.Version(), Platform: runtime.GOOS + "/" + runtime.GOARCH, DependencyHash: checksum([]byte("standard-library-only")), RuleSourceSHA256: checksum(ruleBytes), RuleSourceRevision: SourceRevision}
	if x, e := os.Executable(); e == nil {
		if b, e := os.ReadFile(x); e == nil {
			p.BinarySHA256 = checksum(b)
		}
	}
	if b, ok := debug.ReadBuildInfo(); ok {
		p.BuildInfo = b.String()
	}
	return p
}
func safePath(root, name string) (string, error) {
	if name == "" || filepath.IsAbs(name) || filepath.Clean(name) != name || name == ".." || strings.HasPrefix(name, "../") {
		return "", errors.New("artifact path escape")
	}
	p := filepath.Join(root, name)
	for q := p; ; q = filepath.Dir(q) {
		st, e := os.Lstat(q)
		if e == nil && st.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("symlink forbidden")
		}
		if e != nil && !os.IsNotExist(e) {
			return "", e
		}
		if q == filepath.Dir(q) {
			break
		}
	}
	return p, nil
}
func reserveOutput(dir string) error {
	abs, e := filepath.Abs(dir)
	if e != nil {
		return e
	}
	if _, e = safePath(filepath.Dir(abs), filepath.Base(abs)); e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(abs), 0700); e != nil {
		return e
	}
	return os.Mkdir(abs, 0700)
}
func atomicFile(root, name string, b []byte) error {
	p, e := safePath(root, name)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		return e
	}
	f, e := os.OpenFile(p+".partial", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	if e = os.Rename(p+".partial", p); e != nil {
		return e
	}
	d, e := os.Open(filepath.Dir(p))
	if e != nil {
		return e
	}
	defer d.Close()
	return d.Sync()
}
func writeArtifact(dir, name string, b []byte, privacy string) (Artifact, error) {
	if e := atomicFile(dir, name, b); e != nil {
		return Artifact{}, e
	}
	stored, e := readArtifact(context.Background(), filepath.Join(dir, name), len(b)+1)
	if e != nil {
		return Artifact{}, e
	}
	if checksum(stored) != checksum(b) {
		return Artifact{}, errors.New("published artifact verification failed")
	}
	return Artifact{name, checksum(b), len(b), privacy}, nil
}
func publish(dir string, m Manifest, files map[string][]byte) error {
	names := sortedKeys(files)
	for _, name := range names {
		privacy := "restricted"
		if name == "report.md" || name == "outcomes.json" || name == "metrics.json" {
			privacy = "public"
		}
		a, e := writeArtifact(dir, name, files[name], privacy)
		if e != nil {
			m.Status = "io-failure"
			m.ExitCode = 6
			failure, marshalErr := canonical.Marshal(m)
			if marshalErr != nil {
				return errors.Join(e, marshalErr)
			}
			if publishErr := atomicFile(dir, "manifest.json", failure); publishErr != nil {
				return errors.Join(e, publishErr)
			}
			return e
		}
		m.Artifacts = append(m.Artifacts, a)
	}
	b, e := canonical.Marshal(m)
	if e != nil {
		return e
	}
	return atomicFile(dir, "manifest.json", b)
}
func readManifest(path string) (Manifest, map[string][]byte, error) {
	return readManifestContext(context.Background(), path)
}
func readManifestContext(ctx context.Context, path string) (Manifest, map[string][]byte, error) {
	var m Manifest
	b, e := readArtifact(ctx, path, canonical.MaxBytes)
	if e != nil {
		return m, nil, e
	}
	if e = decodeRequired(b, &m); e != nil {
		return m, nil, e
	}
	if m.Schema != "cgms-manifest-v1" || m.Provenance.Engine != EngineVersion || m.Privacy != "restricted" || m.Provenance.PRNG != randomstream.Algorithm || m.Provenance.SeedDerivation != randomstream.Derivation || m.Provenance.Sampler != randomstream.Sampler || m.Provenance.Shuffle != randomstream.ShuffleVersion || (m.Provenance.Menu != bots.MenuVersion && m.Provenance.Menu != game.GameplayMenuVersion) {
		return m, nil, ErrCompatibility
	}
	files := map[string][]byte{}
	remaining := 256 << 20
	for _, a := range m.Artifacts {
		p, e := safePath(filepath.Dir(path), a.Path)
		if e != nil {
			return m, nil, e
		}
		if _, ok := files[a.Path]; ok {
			return m, nil, errors.New("duplicate artifact")
		}
		if a.Bytes < 0 || a.Bytes > 64<<20 || a.Bytes > remaining {
			return m, nil, ErrResourceBudget
		}
		b, e := readArtifact(ctx, p, a.Bytes+1)
		if e != nil {
			return m, nil, e
		}
		remaining -= len(b)
		if len(b) != a.Bytes || checksum(b) != a.SHA256 {
			return m, nil, fmt.Errorf("artifact integrity: %s", a.Path)
		}
		files[a.Path] = b
	}
	return m, files, nil
}
func manifestBase(args []string, command, root, hash string, workers, n int, rule []byte) Manifest {
	cwd, _ := os.Getwd()
	now := time.Now().UTC()
	return Manifest{Schema: "cgms-manifest-v1", Command: command, Argv: append([]string(nil), args...), CWD: cwd, RootSeed: root, ConfigHash: hash, Interpretation: "accepted-game-rules", Coverage: []string{"104-card-state", "setup", "narrow-combat", "Q10-accounting"}, Exclusions: []string{"full-game-loop", "full-ability-conformance", "online-admission", "stronger-play-evidence"}, Provenance: provenance(rule), Workers: workers, Planned: n, Created: now.Format(time.RFC3339), Expires: now.Add(30 * 24 * time.Hour).Format(time.RFC3339), Privacy: "restricted"}
}

// RetentionPreview never deletes; held evidence and paths outside the artifact root are excluded.
func RetentionPreview(root string, now time.Time) ([]string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if _, err = safePath(filepath.Dir(abs), filepath.Base(abs)); err != nil {
		return nil, err
	}
	type candidate struct {
		path     string
		manifest Manifest
		expiry   time.Time
	}
	runs := []candidate{}
	protected := map[string]bool{}
	err = filepath.WalkDir(abs, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if d.IsDir() || d.Name() != "manifest.json" {
			return nil
		}
		m, _, e := readManifest(p)
		if e != nil {
			return e
		}
		expiry, e := time.Parse(time.RFC3339, m.Expires)
		if e != nil {
			return e
		}
		runs = append(runs, candidate{filepath.Dir(p), m, expiry})
		if m.ReviewHold || !now.After(expiry) {
			protected[filepath.Dir(p)] = true
		}
		if m.Parent == "" {
			return nil
		}
		refs := []string{m.Parent}
		if m.Command == "compare" {
			refs = strings.Split(m.Parent, " | ")
			if len(refs) != 2 {
				return errors.New("ambiguous comparison parent reference")
			}
		}
		for _, ref := range refs {
			if ref == "" || strings.Contains(ref, " | ") || filepath.Base(ref) != "manifest.json" {
				return errors.New("unsupported parent manifest reference")
			}
			if !filepath.IsAbs(ref) {
				if !filepath.IsAbs(m.CWD) {
					return errors.New("relative reference without absolute cwd")
				}
				ref = filepath.Join(m.CWD, ref)
			}
			ref = filepath.Clean(ref)
			// Resolve existing symlinks solely to protect their referent; never traverse
			// symlink directories while enumerating deletion candidates.
			if resolved, e := filepath.EvalSymlinks(ref); e == nil {
				ref = resolved
			} else if !os.IsNotExist(e) {
				return e
			}
			protected[filepath.Dir(ref)] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := []string{}
	for _, run := range runs {
		held := false
		for p := range protected {
			rel, e := filepath.Rel(run.path, p)
			if e == nil && (rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
				held = true
				break
			}
		}
		if !run.manifest.ReviewHold && !held && now.After(run.expiry) {
			out = append(out, run.path)
		}
	}
	return out, nil
}
