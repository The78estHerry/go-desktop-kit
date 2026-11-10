package apppin

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// params are argon2id cost settings. Stored hashes carry their own, so the
// defaults can be raised later without breaking existing PINs.
type params struct {
	memory  uint32 // KiB
	time    uint32
	threads uint8
}

// defaultParams follow the second recommended option of RFC 9106 section 4
// (64 MiB, 3 passes) with 4 lanes.
var defaultParams = params{memory: 64 * 1024, time: 3, threads: 4}

const (
	saltLen = 16
	keyLen  = 32
)

var b64 = base64.RawStdEncoding

// hash returns the PHC string "$argon2id$v=19$m=…,t=…,p=…$salt$hash".
func (pr params) hash(pin []byte) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey(pin, salt, pr.time, pr.memory, pr.threads, keyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, pr.memory, pr.time, pr.threads, b64.EncodeToString(salt), b64.EncodeToString(key)), nil
}

var errFormat = errors.New("apppin: unreadable PIN hash")

// verifyHash compares pin with a PHC string in constant time.
func verifyHash(phc string, pin []byte) (bool, error) {
	f := strings.Split(phc, "$")
	if len(f) != 6 || f[0] != "" || f[1] != "argon2id" {
		return false, errFormat
	}
	var version int
	if _, err := fmt.Sscanf(f[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, errFormat
	}
	var pr params
	if _, err := fmt.Sscanf(f[3], "m=%d,t=%d,p=%d", &pr.memory, &pr.time, &pr.threads); err != nil ||
		pr.memory == 0 || pr.memory > 1<<21 || pr.time == 0 || pr.time > 16 || pr.threads == 0 {
		return false, errFormat // the limits keep a tampered record from exhausting memory
	}
	salt, err1 := b64.DecodeString(f[4])
	want, err2 := b64.DecodeString(f[5])
	if err1 != nil || err2 != nil || len(salt) < 8 || len(want) < 16 {
		return false, errFormat
	}
	got := argon2.IDKey(pin, salt, pr.time, pr.memory, pr.threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
