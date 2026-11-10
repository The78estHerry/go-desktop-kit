package apppin

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/The78estHerry/go-desktop-kit/secretstore"
)

// newTestPIN uses a memory store, cheap hash parameters and a fake clock.
func newTestPIN() (*PIN, secretstore.Store, *time.Time) {
	store := secretstore.NewMemory()
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	p := New(store)
	p.params = params{memory: 64, time: 1, threads: 1}
	p.now = func() time.Time { return now }
	return p, store, &now
}

func TestValid(t *testing.T) {
	for _, s := range []string{"123456", "abcDEF", "a1b2c3d4", strings.Repeat("9", 64)} {
		if !Valid([]byte(s)) {
			t.Errorf("Valid(%q) = false", s)
		}
	}
	for _, s := range []string{"", "12345", "12345 6", "abc-def", "pässwort", "密碼密碼密碼", strings.Repeat("9", 65)} {
		if Valid([]byte(s)) {
			t.Errorf("Valid(%q) = true", s)
		}
	}
}

func TestSetVerify(t *testing.T) {
	p, store, _ := newTestPIN()
	if err := p.Verify([]byte("123456")); !errors.Is(err, ErrNotSet) {
		t.Fatalf("Verify before Set = %v, want ErrNotSet", err)
	}
	if err := p.Set([]byte("12345")); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Set short PIN = %v, want ErrInvalid", err)
	}
	if err := p.Set([]byte("abc123")); err != nil {
		t.Fatal(err)
	}
	if on, _ := p.IsSet(); !on {
		t.Fatal("IsSet = false after Set")
	}
	stored, _ := store.Get(service, keyHash)
	if strings.Contains(string(stored), "abc123") || !strings.HasPrefix(string(stored), "$argon2id$v=19$") {
		t.Fatalf("stored value is not a hash: %q", stored)
	}
	if err := p.Verify([]byte("abc123")); err != nil {
		t.Fatalf("right PIN: %v", err)
	}
	if err := p.Verify([]byte("ABC123")); !errors.Is(err, ErrWrong) {
		t.Fatalf("wrong case = %v, want ErrWrong", err)
	}
}

func TestLockout(t *testing.T) {
	p, _, now := newTestPIN()
	p.Set([]byte("abc123"))
	for i := 1; i <= FreeAttempts-1; i++ {
		if err := p.Verify([]byte("000000")); !errors.Is(err, ErrWrong) {
			t.Fatalf("attempt %d = %v, want ErrWrong", i, err)
		}
	}
	// The fifth mistake locks for 30 seconds.
	if err := p.Verify([]byte("000000")); !errors.Is(err, ErrWrong) {
		t.Fatalf("attempt 5 = %v", err)
	}
	var locked *LockedError
	if err := p.Verify([]byte("abc123")); !errors.As(err, &locked) {
		t.Fatalf("while locked = %v, want LockedError (the right PIN must not be checked)", err)
	}
	if want := now.Add(30 * time.Second); !locked.Until.Equal(want) {
		t.Errorf("locked until %v, want %v", locked.Until, want)
	}

	// Each further mistake doubles the wait: 60 s, 120 s, ...
	for _, wait := range []time.Duration{60 * time.Second, 120 * time.Second, 240 * time.Second} {
		*now = locked.Until
		if err := p.Verify([]byte("000000")); !errors.Is(err, ErrWrong) {
			t.Fatalf("after lock = %v", err)
		}
		until, _ := p.LockedUntil()
		if want := now.Add(wait); !until.Equal(want) {
			t.Errorf("lock %v, want %v", until.Sub(*now), wait)
		}
		locked.Until = until
	}

	// A right PIN after the lock clears the count.
	*now = locked.Until
	if err := p.Verify([]byte("abc123")); err != nil {
		t.Fatalf("right PIN after lock: %v", err)
	}
	if until, _ := p.LockedUntil(); !until.IsZero() {
		t.Errorf("still locked after success: %v", until)
	}
	for i := 1; i < FreeAttempts; i++ {
		if err := p.Verify([]byte("000000")); !errors.Is(err, ErrWrong) {
			t.Fatalf("count was not reset: attempt %d = %v", i, err)
		}
	}
}

func TestLockCappedAtOneHour(t *testing.T) {
	a := attempts{}
	now := time.Unix(0, 0)
	for i := 0; i < 100; i++ {
		a = a.failed(now)
	}
	if got := a.lockedUntil.Sub(now); got != MaxLock {
		t.Errorf("lock after 100 mistakes = %v, want %v", got, MaxLock)
	}
	a = attempts{failures: FreeAttempts + 6}.failed(now) // 30 s << 7 = 64 min
	if got := a.lockedUntil.Sub(now); got != MaxLock {
		t.Errorf("lock = %v, want the cap %v", got, MaxLock)
	}
}

func TestAttemptsSurviveRestart(t *testing.T) {
	p, store, now := newTestPIN()
	p.Set([]byte("abc123"))
	for i := 0; i < FreeAttempts; i++ {
		p.Verify([]byte("000000"))
	}
	// A new PIN object over the same store, as after restarting the program.
	q := New(store)
	q.params, q.now = p.params, func() time.Time { return *now }
	var locked *LockedError
	if err := q.Verify([]byte("abc123")); !errors.As(err, &locked) {
		t.Fatalf("after restart = %v, want LockedError", err)
	}
}

func TestChangeAndReset(t *testing.T) {
	p, _, _ := newTestPIN()
	p.Set([]byte("abc123"))
	if err := p.Change([]byte("wrong1"), []byte("new456")); !errors.Is(err, ErrWrong) {
		t.Fatalf("Change with wrong old PIN = %v", err)
	}
	if err := p.Change([]byte("abc123"), []byte("new456")); err != nil {
		t.Fatal(err)
	}
	if err := p.Verify([]byte("new456")); err != nil {
		t.Fatalf("new PIN: %v", err)
	}
	if err := p.Reset(); err != nil {
		t.Fatal(err)
	}
	if on, _ := p.IsSet(); on {
		t.Fatal("IsSet after Reset")
	}
	if err := p.Reset(); err != nil {
		t.Fatalf("second Reset: %v", err)
	}
}

func TestDefaultHashFormat(t *testing.T) {
	h, err := defaultParams.hash([]byte("abc123"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "$argon2id$v=19$m=65536,t=3,p=4$") {
		t.Errorf("hash = %s", h)
	}
	if ok, err := verifyHash(h, []byte("abc123")); !ok || err != nil {
		t.Errorf("verifyHash = %v, %v", ok, err)
	}
	for _, bad := range []string{"", "plain", "$argon2i$v=19$m=64,t=1,p=1$AAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAA",
		"$argon2id$v=19$m=999999999,t=1,p=1$AAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAA"} {
		if _, err := verifyHash(bad, []byte("x")); err == nil {
			t.Errorf("verifyHash(%q) accepted a bad record", bad)
		}
	}
}
