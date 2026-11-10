// Package apppin is an application PIN that confirms the user is present
// when the operating system cannot (no Windows Hello, Linux).
//
// Like an OS prompt, it is a presence check, not encryption: it stops
// someone (or an AI agent) driving the GUI, not a program that reads the
// credential store directly.
//
//   - Only an argon2id hash is stored, never the PIN.
//   - A PIN is 6 to 64 ASCII letters and digits, so input methods and
//     keyboard layouts cannot get in the way.
//   - After 5 wrong attempts in a row the PIN locks for 30 seconds,
//     doubling with each further mistake up to one hour. The count lives in
//     the store next to the hash, so restarting the program or deleting
//     its settings does not reset it.
//   - There is no way around a forgotten PIN: Reset removes it, and the
//     application decides what else to wipe.
//
// Typical use:
//
//	store, _ := secretstore.Open("my-app")
//	pin := apppin.New(store)
//	err := pin.Verify(entered)
//	var locked *apppin.LockedError
//	switch {
//	case err == nil: // the user is present
//	case errors.As(err, &locked): // wait until locked.Until
//	case errors.Is(err, apppin.ErrWrong):
//	}
package apppin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/The78estHerry/go-desktop-kit/secretstore"
)

// Errors returned by PIN methods.
var (
	ErrNotSet = errors.New("apppin: no PIN is set")
	ErrWrong  = errors.New("apppin: wrong PIN")
	// ErrInvalid is returned by Set and Change for a PIN that is not 6 to
	// 64 ASCII letters and digits.
	ErrInvalid = errors.New("apppin: a PIN must be 6 to 64 letters and digits")
)

// LockedError is returned while too many wrong attempts lock the PIN.
type LockedError struct{ Until time.Time }

func (e *LockedError) Error() string {
	return "apppin: too many wrong attempts; try again after " + e.Until.Format(time.TimeOnly)
}

// Limits on the PIN and on attempts.
const (
	MinLength    = 6
	MaxLength    = 64
	FreeAttempts = 5                // wrong attempts before the first lock
	FirstLock    = 30 * time.Second // doubled with each further mistake
	MaxLock      = time.Hour
)

// Keys under the "apppin" service of the store.
const (
	service     = "apppin"
	keyHash     = "hash"
	keyAttempts = "attempts"
)

// PIN is the application PIN kept in a secret store.
type PIN struct {
	store  secretstore.Store
	params params
	now    func() time.Time
}

// New returns the PIN kept in store; normally the application's own store
// from secretstore.Open.
func New(store secretstore.Store) *PIN {
	return &PIN{store: store, params: defaultParams, now: time.Now}
}

// Valid reports whether pin is acceptable as a new PIN.
func Valid(pin []byte) bool {
	if len(pin) < MinLength || len(pin) > MaxLength {
		return false
	}
	for _, c := range pin {
		if !('0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z') {
			return false
		}
	}
	return true
}

// IsSet reports whether a PIN has been set.
func (p *PIN) IsSet() (bool, error) {
	_, err := p.store.Get(service, keyHash)
	if errors.Is(err, secretstore.ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

// Set stores a new PIN, replacing any old one, and clears the attempt
// count. Use Change when the user must prove they know the old PIN.
func (p *PIN) Set(pin []byte) error {
	if !Valid(pin) {
		return ErrInvalid
	}
	h, err := p.params.hash(pin)
	if err != nil {
		return err
	}
	if err := p.store.Set(service, keyHash, []byte(h)); err != nil {
		return err
	}
	return p.clearAttempts()
}

// Change replaces the PIN after verifying the old one.
func (p *PIN) Change(old, pin []byte) error {
	if !Valid(pin) {
		return ErrInvalid
	}
	if err := p.Verify(old); err != nil {
		return err
	}
	return p.Set(pin)
}

// Reset removes the PIN and the attempt count. It is the only way out of
// a forgotten PIN; the application should wipe what the PIN protected.
func (p *PIN) Reset() error {
	for _, k := range []string{keyHash, keyAttempts} {
		if err := p.store.Delete(service, k); err != nil && !errors.Is(err, secretstore.ErrNotFound) {
			return err
		}
	}
	return nil
}

// Verify checks pin. It returns nil when it matches, ErrWrong when it does
// not, *LockedError while locked (without checking) and ErrNotSet when no
// PIN exists.
func (p *PIN) Verify(pin []byte) error {
	stored, err := p.store.Get(service, keyHash)
	if errors.Is(err, secretstore.ErrNotFound) {
		return ErrNotSet
	}
	if err != nil {
		return err
	}
	a, err := p.loadAttempts()
	if err != nil {
		return err
	}
	now := p.now()
	if now.Before(a.lockedUntil) {
		return &LockedError{Until: a.lockedUntil}
	}
	// Count the attempt before checking, so killing the program between
	// the check and the write cannot give free guesses.
	a = a.failed(now)
	if err := p.saveAttempts(a); err != nil {
		return err
	}
	ok, err := verifyHash(string(stored), pin)
	if err != nil {
		return err
	}
	if !ok {
		return ErrWrong
	}
	return p.clearAttempts()
}

// LockedUntil returns when the current lock ends, or the zero time when
// the PIN is not locked.
func (p *PIN) LockedUntil() (time.Time, error) {
	a, err := p.loadAttempts()
	if err != nil || !p.now().Before(a.lockedUntil) {
		return time.Time{}, err
	}
	return a.lockedUntil, nil
}

// attempts is the persisted count of consecutive wrong attempts.
type attempts struct {
	failures    int
	lockedUntil time.Time
}

// failed returns the state after one more wrong attempt at now.
func (a attempts) failed(now time.Time) attempts {
	a.failures++
	a.lockedUntil = time.Time{}
	if extra := a.failures - FreeAttempts; extra >= 0 {
		d := MaxLock
		if extra < 20 { // 30s << 20 is far beyond MaxLock
			d = min(FirstLock<<extra, MaxLock)
		}
		a.lockedUntil = now.Add(d)
	}
	return a
}

// The record is "v1 <failures> <locked until, Unix seconds>".
func (p *PIN) loadAttempts() (attempts, error) {
	b, err := p.store.Get(service, keyAttempts)
	if errors.Is(err, secretstore.ErrNotFound) {
		return attempts{}, nil
	}
	if err != nil {
		return attempts{}, err
	}
	f := strings.Fields(string(b))
	if len(f) != 3 || f[0] != "v1" {
		return attempts{}, fmt.Errorf("apppin: unreadable attempt record")
	}
	n, err1 := strconv.Atoi(f[1])
	until, err2 := strconv.ParseInt(f[2], 10, 64)
	if err1 != nil || err2 != nil || n < 0 {
		return attempts{}, fmt.Errorf("apppin: unreadable attempt record")
	}
	a := attempts{failures: n}
	if until > 0 {
		a.lockedUntil = time.Unix(until, 0)
	}
	return a, nil
}

func (p *PIN) saveAttempts(a attempts) error {
	var until int64
	if !a.lockedUntil.IsZero() {
		// Round up so the lock never ends early.
		until = a.lockedUntil.Add(time.Second - 1).Unix()
	}
	return p.store.Set(service, keyAttempts, []byte(fmt.Sprintf("v1 %d %d", a.failures, until)))
}

func (p *PIN) clearAttempts() error {
	if err := p.store.Delete(service, keyAttempts); err != nil && !errors.Is(err, secretstore.ErrNotFound) {
		return err
	}
	return nil
}
