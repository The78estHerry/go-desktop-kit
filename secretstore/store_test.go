package secretstore

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/The78estHerry/go-desktop-kit/sealbox"
	"github.com/zalando/go-keyring"
)

// testStore runs the behavior every backend must share.
func testStore(t *testing.T, s Store) {
	t.Helper()
	if _, err := s.Get("svc", "token"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get missing: got %v, want ErrNotFound", err)
	}
	if err := s.Delete("svc", "token"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete missing: got %v, want ErrNotFound", err)
	}

	secret := []byte("test-token-not-real")
	if err := s.Set("svc", "token", secret); err != nil {
		t.Fatalf("Set: %v", err)
	}
	secret[0] = 'X' // the store must keep its own copy
	got, err := s.Get("svc", "token")
	if err != nil || string(got) != "test-token-not-real" {
		t.Fatalf("Get: %q, %v", got, err)
	}

	if err := s.Set("svc", "token", []byte("test-token-not-real-2")); err != nil {
		t.Fatalf("Set replace: %v", err)
	}
	if got, _ := s.Get("svc", "token"); string(got) != "test-token-not-real-2" {
		t.Fatalf("Get after replace: %q", got)
	}

	if err := s.Set("svc", "user.secret", []byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("other", "token", []byte("y")); err != nil {
		t.Fatal(err)
	}
	keys, err := s.List("svc")
	if err != nil || !slices.Equal(keys, []string{"token", "user.secret"}) {
		t.Fatalf("List: %v, %v", keys, err)
	}
	if keys, _ := s.List("none"); len(keys) != 0 {
		t.Fatalf("List empty service: %v", keys)
	}

	for _, n := range [][2]string{{"svc", "token"}, {"svc", "user.secret"}, {"other", "token"}} {
		if err := s.Delete(n[0], n[1]); err != nil {
			t.Fatalf("Delete %v: %v", n, err)
		}
	}
	if keys, _ := s.List("svc"); len(keys) != 0 {
		t.Fatalf("List after delete: %v", keys)
	}

	for _, bad := range []string{"", "a/b", "a:b", "_x", "with space", "名前"} {
		if err := s.Set(bad, "k", nil); !errors.Is(err, ErrInvalidName) {
			t.Errorf("Set(%q): got %v, want ErrInvalidName", bad, err)
		}
		if _, err := s.Get("svc", bad); !errors.Is(err, ErrInvalidName) {
			t.Errorf("Get key %q: got %v, want ErrInvalidName", bad, err)
		}
	}
}

func TestMemory(t *testing.T) {
	s, err := Open("test-app", WithBackend(Memory))
	if err != nil {
		t.Fatal(err)
	}
	testStore(t, s)
}

func TestAgeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("age file writes use the full scrypt work factor")
	}
	path := filepath.Join(t.TempDir(), "secrets.age")
	pass := func() ([]byte, error) { return []byte("test passphrase 測試"), nil }
	s, err := Open("test-app", WithBackend(AgeFile), WithAgeFile(path, pass))
	if err != nil {
		t.Fatal(err)
	}
	testStore(t, s)

	// The file is a sealbox file and holds no plain text.
	if err := s.Set("svc", "token", []byte("test-token-not-real")); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := sealbox.Open(f, []byte("wrong")); !errors.Is(err, sealbox.ErrWrongPassphrase) {
		t.Fatalf("file is not a sealbox file: %v", err)
	}

	// A second store on the same file sees the first store's write.
	s2, err := Open("test-app", WithBackend(AgeFile), WithAgeFile(path, pass))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := s2.Get("svc", "token"); err != nil || string(got) != "test-token-not-real" {
		t.Fatalf("second store Get: %q, %v", got, err)
	}

	// The wrong passphrase is rejected when opening.
	_, err = Open("test-app", WithBackend(AgeFile), WithAgeFile(path, func() ([]byte, error) { return []byte("nope"), nil }))
	if !errors.Is(err, sealbox.ErrWrongPassphrase) {
		t.Fatalf("wrong passphrase: got %v", err)
	}
}

// TestKeyring uses the real OS credential store under a random application ID
// and removes everything it creates.
func TestKeyring(t *testing.T) {
	b := make([]byte, 4)
	rand.Read(b)
	appID := "go-desktop-kit-test-" + hex.EncodeToString(b)
	if err := probeKeyring(appID); err != nil {
		t.Skipf("OS credential store unavailable: %v", err)
	}
	t.Cleanup(func() { keyring.DeleteAll(appID) })
	s, err := Open(appID, WithBackend(Keyring))
	if err != nil {
		t.Fatal(err)
	}
	testStore(t, s)
	if _, err := keyring.Get(appID, indexUser); !errors.Is(err, keyring.ErrNotFound) {
		t.Errorf("index entry left behind after deleting everything: %v", err)
	}
}

func TestOpen(t *testing.T) {
	if _, err := Open("bad/app"); !errors.Is(err, ErrInvalidName) {
		t.Errorf("bad appID: got %v", err)
	}
	if _, err := Open("test-app", WithBackend(AgeFile)); !errors.Is(err, ErrUnavailable) {
		t.Errorf("age file without WithAgeFile: got %v", err)
	}
}
