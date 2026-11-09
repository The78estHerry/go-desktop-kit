// Package secretstore keeps secrets in the operating system's credential
// store (Windows Credential Manager, macOS Keychain, Linux Secret Service),
// with an age-encrypted file as the fallback for Linux systems without a
// desktop session, and an in-memory store for tests.
//
// Secrets are addressed by an application ID, a service and a key, for
// example "credential-porter", "gitea-home", "token". Keep non-secret data
// such as URLs and user names in ordinary configuration files instead, so
// listing configured services never needs the secret store.
//
// The package has no API that prints or exports secrets; callers decide what
// to do with a value returned by Get and should wipe it when done.
package secretstore

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	// ErrNotFound is returned by Get and Delete when no secret exists.
	ErrNotFound = errors.New("secretstore: secret not found")
	// ErrInvalidName is returned for an application ID, service or key that
	// does not match the allowed pattern (see ValidName).
	ErrInvalidName = errors.New("secretstore: invalid name")
	// ErrUnavailable is returned by Open when no backend can be used.
	ErrUnavailable = errors.New("secretstore: no usable backend")
	// ErrTooLarge is returned by Set when the backend cannot hold the secret.
	ErrTooLarge = errors.New("secretstore: secret too large for this backend")
)

// Store is a place to keep secrets.
type Store interface {
	// Set stores secret under service and key, replacing any existing value.
	Set(service, key string, secret []byte) error
	// Get returns the secret, or ErrNotFound.
	Get(service, key string) ([]byte, error)
	// Delete removes the secret, or returns ErrNotFound.
	Delete(service, key string) error
	// List returns the sorted key names stored under service, never values.
	List(service string) ([]string, error)
}

// Backend selects where secrets are kept.
type Backend int

const (
	// Auto uses the OS credential store when it works, and otherwise the
	// age-encrypted file configured with WithAgeFile.
	Auto Backend = iota
	// Keyring is the OS credential store.
	Keyring
	// AgeFile is a single file encrypted with a passphrase (see WithAgeFile).
	AgeFile
	// Memory keeps secrets in memory only; for tests.
	Memory
)

type options struct {
	backend    Backend
	agePath    string
	passphrase func() ([]byte, error)
}

// Option configures Open.
type Option func(*options)

// WithBackend forces a backend instead of Auto.
func WithBackend(b Backend) Option {
	return func(o *options) { o.backend = b }
}

// WithAgeFile configures the age-encrypted file backend. passphrase is called
// once, when the file backend is opened; the GUI should ask the user for it.
func WithAgeFile(path string, passphrase func() ([]byte, error)) Option {
	return func(o *options) {
		o.agePath = path
		o.passphrase = passphrase
	}
}

var namePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

// ValidName reports whether s can be used as an application ID, service or
// key: 1 to 64 ASCII letters, digits, '.', '_' or '-', starting with a letter
// or digit.
func ValidName(s string) bool { return namePattern.MatchString(s) }

func checkNames(names ...string) error {
	for _, n := range names {
		if !ValidName(n) {
			return fmt.Errorf("%w: %q", ErrInvalidName, n)
		}
	}
	return nil
}

// Open returns a Store for appID.
func Open(appID string, opts ...Option) (Store, error) {
	if err := checkNames(appID); err != nil {
		return nil, err
	}
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	switch o.backend {
	case Memory:
		return NewMemory(), nil
	case Keyring:
		if err := probeKeyring(appID); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
		}
		return newKeyringStore(appID), nil
	case AgeFile:
		return openAgeFile(o)
	default:
		if probeKeyring(appID) == nil {
			return newKeyringStore(appID), nil
		}
		if o.passphrase != nil {
			return openAgeFile(o)
		}
		return nil, ErrUnavailable
	}
}

func entryName(service, key string) string { return service + "/" + key }
