package sealbox

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"filippo.io/age"
	"golang.org/x/text/unicode/norm"
)

func init() { workFactor = 10 } // keep tests fast

func seal(t *testing.T, plaintext, passphrase string) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := Seal(&buf, []byte(plaintext), []byte(passphrase)); err != nil {
		t.Fatalf("Seal: %v", err)
	}
	return buf.Bytes()
}

func TestRoundTrip(t *testing.T) {
	for _, plaintext := range []string{"", "test-token-not-real", strings.Repeat("x", 100000)} {
		sealed := seal(t, plaintext, "correct horse battery staple")
		got, err := Open(bytes.NewReader(sealed), []byte("correct horse battery staple"))
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		if string(got) != plaintext {
			t.Fatalf("got %d bytes, want %d", len(got), len(plaintext))
		}
	}
}

func TestWrongPassphrase(t *testing.T) {
	sealed := seal(t, "test-token-not-real", "right")
	if _, err := Open(bytes.NewReader(sealed), []byte("wrong")); !errors.Is(err, ErrWrongPassphrase) {
		t.Fatalf("got %v, want ErrWrongPassphrase", err)
	}
}

func TestNotSealed(t *testing.T) {
	for _, in := range []string{"", "hello", `{"version":1}`} {
		if _, err := Open(strings.NewReader(in), []byte("x")); !errors.Is(err, ErrNotSealed) {
			t.Errorf("%q: got %v, want ErrNotSealed", in, err)
		}
	}
}

func TestEmptyPassphrase(t *testing.T) {
	if err := Seal(io.Discard, []byte("x"), nil); !errors.Is(err, ErrEmptyPassphrase) {
		t.Errorf("Seal: got %v", err)
	}
	if _, err := Open(strings.NewReader(header), nil); !errors.Is(err, ErrEmptyPassphrase) {
		t.Errorf("Open: got %v", err)
	}
}

func TestTampered(t *testing.T) {
	sealed := seal(t, "test-token-not-real", "pw")
	sealed[len(sealed)-1] ^= 1
	_, err := Open(bytes.NewReader(sealed), []byte("pw"))
	if err == nil || errors.Is(err, ErrWrongPassphrase) {
		t.Fatalf("got %v, want an integrity error", err)
	}
}

// A passphrase typed as NFD on one system must open data sealed with the NFC
// form on another, and the other way round.
func TestUnicodeNormalization(t *testing.T) {
	for _, p := range []string{"café", "密碼是 Ǎo 與 ü", "パスワード ガギグ"} {
		nfc, nfd := norm.NFC.String(p), norm.NFD.String(p)
		if nfc == nfd {
			t.Fatalf("%q: test input has no decomposable characters", p)
		}
		for _, pair := range [][2]string{{nfc, nfd}, {nfd, nfc}} {
			sealed := seal(t, "test-token-not-real", pair[0])
			if _, err := Open(bytes.NewReader(sealed), []byte(pair[1])); err != nil {
				t.Errorf("%q: sealed with %+q, opened with %+q: %v", p, pair[0], pair[1], err)
			}
		}
	}
}

// Data sealed by sealbox must open with plain age and the NFC passphrase,
// which is what `age -d` does.
func TestStandardAgeFormat(t *testing.T) {
	sealed := seal(t, "test-token-not-real", norm.NFD.String("café"))
	id, err := age.NewScryptIdentity(norm.NFC.String("café"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := age.Decrypt(bytes.NewReader(sealed), id)
	if err != nil {
		t.Fatalf("age.Decrypt: %v", err)
	}
	got, _ := io.ReadAll(r)
	if string(got) != "test-token-not-real" {
		t.Fatalf("got %q", got)
	}
}
