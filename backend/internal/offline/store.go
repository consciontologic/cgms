package offline

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"io"
	"os"
	"regexp"
	"sync"
)

var ErrDurabilityUnknown = errors.New("save durability unknown; reload before retry")

var slotPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

type saveFile struct {
	Schema   string   `json:"schema"`
	Checksum string   `json:"checksum"`
	Session  *Session `json:"session"`
}

// Store is one host process's storage owner. os.Root prevents slot/symlink path
// escape. The parent directory must already be a private app-owned directory.
type Store struct {
	root          *os.Root
	lock          *os.File
	rules         string
	mu            sync.Mutex
	rename        func(string, string) error
	syncDirectory func() error
}

func OpenStore(directory, rules string) (*Store, error) {
	if rules == "" {
		return nil, errors.New("rules fingerprint required")
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	lock, err := lockStore(root)
	if err != nil {
		root.Close()
		return nil, err
	}
	return &Store{root: root, lock: lock, rules: rules, rename: root.Rename, syncDirectory: func() error {
		dir, e := root.Open(".")
		if e != nil {
			return e
		}
		defer dir.Close()
		return dir.Sync()
	}}, nil
}
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := s.lock.Close()
	b := s.root.Close()
	if a != nil {
		return a
	}
	return b
}
func digest(s *Session) (string, error) {
	raw, err := canonical.Marshal(s)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
func (s *Store) Save(slot string, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !slotPattern.MatchString(slot) {
		return errors.New("invalid save slot")
	}
	if err := session.Validate(s.rules); err != nil {
		return err
	}
	checksum, err := digest(session)
	if err != nil {
		return err
	}
	raw, err := canonical.Marshal(saveFile{"cgms-offline-save-v1", checksum, session})
	if err != nil {
		return err
	}
	if len(raw) > MaxSaveBytes {
		return errors.New("save too large")
	}
	id, err := newID()
	if err != nil {
		return err
	}
	tmp := "." + id + ".tmp"
	file, err := s.root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer s.root.Remove(tmp)
	_, err = file.Write(raw)
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = s.rename(tmp, slot+".json"); err != nil {
		return err
	}
	if err = s.syncDirectory(); err != nil {
		return ErrDurabilityUnknown
	}
	return nil
}
func (s *Store) Load(slot string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !slotPattern.MatchString(slot) {
		return nil, errors.New("invalid save slot")
	}
	f, err := openSave(s.root, slot+".json")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > MaxSaveBytes {
		return nil, errors.New("invalid save file")
	}
	raw, err := io.ReadAll(io.LimitReader(f, MaxSaveBytes+1))
	if err != nil {
		return nil, err
	}
	var saved saveFile
	if err = canonical.DecodeLimit(raw, &saved, MaxSaveBytes); err != nil {
		return nil, err
	}
	if saved.Schema != "cgms-offline-save-v1" {
		return nil, errors.New("incompatible save schema")
	}
	if err = saved.Session.Validate(s.rules); err != nil {
		return nil, err
	}
	checksum, err := digest(saved.Session)
	if err != nil {
		return nil, err
	}
	if checksum != saved.Checksum {
		return nil, errors.New("save checksum mismatch")
	}
	return saved.Session, nil
}
