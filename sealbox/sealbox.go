// Package sealbox encrypts data with a passphrase into the standard age
// format (scrypt recipient), so a sealed file can always be opened with the
// official age tool as well:
//
//	age -d backup.cpbak
//
// Passphrases may contain any Unicode text. They are normalized to NFC before
// use, so the same passphrase typed on Windows, macOS or Linux produces the
// same key even when an input method emits decomposed characters.
//
// The passphrase is never stored. Losing it means the data cannot be
// recovered; this is intentional.
package sealbox

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"filippo.io/age"
	"golang.org/x/text/unicode/norm"
)

var (
	// ErrWrongPassphrase is returned by Open when the passphrase does not
	// match the one used to seal the data.
	ErrWrongPassphrase = errors.New("sealbox: wrong passphrase")
	// ErrNotSealed is returned by Open when the input is not age-encrypted data.
	ErrNotSealed = errors.New("sealbox: input is not a sealed (age) file")
	// ErrEmptyPassphrase is returned when the passphrase is empty.
	ErrEmptyPassphrase = errors.New("sealbox: empty passphrase")
)

// header is the first line of every binary age file.
const header = "age-encryption.org/v1\n"

// workFactor is the scrypt work factor (log2 N) used by Seal. age's default
// is 18, about one second on a typical computer. Tests lower it.
var workFactor = 18

// maxWorkFactor is the highest work factor Open accepts, matching age's default.
const maxWorkFactor = 22

// normalize returns the NFC form of the passphrase.
func normalize(passphrase []byte) string {
	return norm.NFC.String(string(passphrase))
}

// Seal encrypts plaintext with passphrase and writes age-format data to w.
func Seal(w io.Writer, plaintext, passphrase []byte) error {
	if len(passphrase) == 0 {
		return ErrEmptyPassphrase
	}
	r, err := age.NewScryptRecipient(normalize(passphrase))
	if err != nil {
		return fmt.Errorf("sealbox: %w", err)
	}
	r.SetWorkFactor(workFactor)
	wc, err := age.Encrypt(w, r)
	if err != nil {
		return fmt.Errorf("sealbox: %w", err)
	}
	if _, err := wc.Write(plaintext); err != nil {
		return fmt.Errorf("sealbox: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("sealbox: %w", err)
	}
	return nil
}

// Open decrypts age-format data read from r with passphrase.
// It returns ErrWrongPassphrase or ErrNotSealed for those two cases.
func Open(r io.Reader, passphrase []byte) ([]byte, error) {
	if len(passphrase) == 0 {
		return nil, ErrEmptyPassphrase
	}
	br := bufio.NewReader(r)
	if prefix, err := br.Peek(len(header)); err != nil || string(prefix) != header {
		return nil, ErrNotSealed
	}
	id, err := age.NewScryptIdentity(normalize(passphrase))
	if err != nil {
		return nil, fmt.Errorf("sealbox: %w", err)
	}
	id.SetMaxWorkFactor(maxWorkFactor)
	dr, err := age.Decrypt(br, id)
	if err != nil {
		var noMatch *age.NoIdentityMatchError
		if errors.As(err, &noMatch) {
			return nil, ErrWrongPassphrase
		}
		return nil, fmt.Errorf("sealbox: %w", err)
	}
	plaintext, err := io.ReadAll(dr)
	if err != nil {
		return nil, fmt.Errorf("sealbox: %w", err)
	}
	return plaintext, nil
}
