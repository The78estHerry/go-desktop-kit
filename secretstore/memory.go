package secretstore

import (
	"slices"
	"strings"
	"sync"
)

type memoryStore struct {
	mu sync.Mutex
	m  map[string][]byte
}

// NewMemory returns a Store that keeps secrets in memory only.
func NewMemory() Store { return &memoryStore{m: map[string][]byte{}} }

func (s *memoryStore) Set(service, key string, secret []byte) error {
	if err := checkNames(service, key); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[entryName(service, key)] = slices.Clone(secret)
	return nil
}

func (s *memoryStore) Get(service, key string) ([]byte, error) {
	if err := checkNames(service, key); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.m[entryName(service, key)]
	if !ok {
		return nil, ErrNotFound
	}
	return slices.Clone(v), nil
}

func (s *memoryStore) Delete(service, key string) error {
	if err := checkNames(service, key); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	name := entryName(service, key)
	v, ok := s.m[name]
	if !ok {
		return ErrNotFound
	}
	clear(v)
	delete(s.m, name)
	return nil
}

func (s *memoryStore) List(service string) ([]string, error) {
	if err := checkNames(service); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return keysUnder(s.m, service), nil
}

// keysUnder returns the sorted keys of entries named "<service>/<key>".
func keysUnder[V any](m map[string]V, service string) []string {
	prefix := service + "/"
	keys := []string{}
	for name := range m {
		if k, ok := strings.CutPrefix(name, prefix); ok {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	return keys
}
