package secretstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/The78estHerry/go-desktop-kit/sealbox"
)

// ageFileStore keeps every secret in one sealbox (age) file. The file is
// decrypted into memory when opened and again whenever another process has
// changed it; every write re-encrypts the whole file and replaces it
// atomically.
type ageFileStore struct {
	mu         sync.Mutex
	path       string
	passphrase []byte
	entries    map[string][]byte
	modTime    time.Time
	size       int64
}

type ageFileContent struct {
	Version int               `json:"version"`
	Secrets map[string][]byte `json:"secrets"`
}

func openAgeFile(o options) (Store, error) {
	if o.agePath == "" || o.passphrase == nil {
		return nil, fmt.Errorf("%w: age file backend needs WithAgeFile", ErrUnavailable)
	}
	pass, err := o.passphrase()
	if err != nil {
		return nil, err
	}
	s := &ageFileStore{path: o.agePath, passphrase: slices.Clone(pass)}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// load reads the file; a missing file is an empty store.
func (s *ageFileStore) load() error {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		s.entries = map[string][]byte{}
		s.modTime, s.size = time.Time{}, -1
		return nil
	}
	if err != nil {
		return fmt.Errorf("secretstore: %w", err)
	}
	plain, err := sealbox.Open(bytes.NewReader(data), s.passphrase)
	if err != nil {
		return err
	}
	defer clear(plain)
	var c ageFileContent
	if err := json.Unmarshal(plain, &c); err != nil {
		return fmt.Errorf("secretstore: corrupt age file: %w", err)
	}
	if c.Version != 1 {
		return fmt.Errorf("secretstore: unsupported age file version %d", c.Version)
	}
	if c.Secrets == nil {
		c.Secrets = map[string][]byte{}
	}
	s.entries = c.Secrets
	return s.stamp()
}

func (s *ageFileStore) stamp() error {
	fi, err := os.Stat(s.path)
	if err != nil {
		return fmt.Errorf("secretstore: %w", err)
	}
	s.modTime, s.size = fi.ModTime(), fi.Size()
	return nil
}

// refresh reloads the file if another process changed it.
func (s *ageFileStore) refresh() error {
	fi, err := os.Stat(s.path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		if s.size == -1 {
			return nil
		}
	case err != nil:
		return fmt.Errorf("secretstore: %w", err)
	case fi.ModTime().Equal(s.modTime) && fi.Size() == s.size:
		return nil
	}
	return s.load()
}

func (s *ageFileStore) save() error {
	plain, err := json.Marshal(ageFileContent{Version: 1, Secrets: s.entries})
	if err != nil {
		return err
	}
	defer clear(plain)
	var sealed bytes.Buffer
	if err := sealbox.Seal(&sealed, plain, s.passphrase); err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("secretstore: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".secretstore-*.tmp")
	if err != nil {
		return fmt.Errorf("secretstore: %w", err)
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil && !errors.Is(err, errors.ErrUnsupported) {
		tmp.Close()
		return fmt.Errorf("secretstore: %w", err)
	}
	if _, err := tmp.Write(sealed.Bytes()); err != nil {
		tmp.Close()
		return fmt.Errorf("secretstore: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("secretstore: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("secretstore: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("secretstore: %w", err)
	}
	return s.stamp()
}

func (s *ageFileStore) Set(service, key string, secret []byte) error {
	if err := checkNames(service, key); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return err
	}
	s.entries[entryName(service, key)] = slices.Clone(secret)
	return s.save()
}

func (s *ageFileStore) Get(service, key string) ([]byte, error) {
	if err := checkNames(service, key); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return nil, err
	}
	v, ok := s.entries[entryName(service, key)]
	if !ok {
		return nil, ErrNotFound
	}
	return slices.Clone(v), nil
}

func (s *ageFileStore) Delete(service, key string) error {
	if err := checkNames(service, key); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return err
	}
	name := entryName(service, key)
	if _, ok := s.entries[name]; !ok {
		return ErrNotFound
	}
	delete(s.entries, name)
	return s.save()
}

func (s *ageFileStore) List(service string) ([]string, error) {
	if err := checkNames(service); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return nil, err
	}
	return keysUnder(s.entries, service), nil
}
