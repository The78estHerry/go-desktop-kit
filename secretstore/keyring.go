package secretstore

// Status: tested on Windows (Credential Manager). macOS (Keychain) and Linux
// (Secret Service) use the same code through go-keyring and are implemented,
// not yet tested on real hardware.

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/zalando/go-keyring"
)

// keyringStore maps an entry to go-keyring service=appID and
// user="<service>/<key>". On Windows the credential's target name is
// "<appID>:<service>/<key>".
//
// The OS stores cannot list entries portably, so the store keeps an index of
// entry names (never values) in one more entry, user=indexUser. The index is
// limited by the backend's size limit (about 2.5 KB on Windows, roughly 100
// entries).
type keyringStore struct {
	appID string
	mu    sync.Mutex
}

// indexUser cannot collide with an entry: entry names always contain '/'.
const indexUser = "_index"

func newKeyringStore(appID string) *keyringStore { return &keyringStore{appID: appID} }

// probeKeyring reports whether the OS credential store answers at all.
func probeKeyring(appID string) error {
	_, err := keyring.Get(appID, "_probe")
	if err == nil || errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func mapKeyringErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, keyring.ErrNotFound):
		return ErrNotFound
	case errors.Is(err, keyring.ErrSetDataTooBig):
		return ErrTooLarge
	default:
		return fmt.Errorf("secretstore: %w", err)
	}
}

func (s *keyringStore) Set(service, key string, secret []byte) error {
	if err := checkNames(service, key); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	name := entryName(service, key)
	if err := mapKeyringErr(keyring.Set(s.appID, name, string(secret))); err != nil {
		return err
	}
	return s.updateIndex(func(idx []string) []string {
		if !slices.Contains(idx, name) {
			idx = append(idx, name)
		}
		return idx
	})
}

func (s *keyringStore) Get(service, key string) ([]byte, error) {
	if err := checkNames(service, key); err != nil {
		return nil, err
	}
	v, err := keyring.Get(s.appID, entryName(service, key))
	if err != nil {
		return nil, mapKeyringErr(err)
	}
	return []byte(v), nil
}

func (s *keyringStore) Delete(service, key string) error {
	if err := checkNames(service, key); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	name := entryName(service, key)
	if err := mapKeyringErr(keyring.Delete(s.appID, name)); err != nil {
		return err
	}
	return s.updateIndex(func(idx []string) []string {
		return slices.DeleteFunc(idx, func(n string) bool { return n == name })
	})
}

func (s *keyringStore) List(service string) ([]string, error) {
	if err := checkNames(service); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	idx, err := s.readIndex()
	if err != nil {
		return nil, err
	}
	m := make(map[string]struct{}, len(idx))
	for _, n := range idx {
		m[n] = struct{}{}
	}
	return keysUnder(m, service), nil
}

func (s *keyringStore) readIndex() ([]string, error) {
	v, err := keyring.Get(s.appID, indexUser)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, mapKeyringErr(err)
	}
	var idx []string
	if err := json.Unmarshal([]byte(v), &idx); err != nil {
		return nil, fmt.Errorf("secretstore: corrupt index: %w", err)
	}
	return idx, nil
}

// updateIndex must be called with s.mu held.
func (s *keyringStore) updateIndex(change func([]string) []string) error {
	idx, err := s.readIndex()
	if err != nil {
		return err
	}
	idx = change(idx)
	if len(idx) == 0 {
		err := keyring.Delete(s.appID, indexUser)
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return mapKeyringErr(err)
	}
	slices.Sort(idx)
	b, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	return mapKeyringErr(keyring.Set(s.appID, indexUser, string(b)))
}
