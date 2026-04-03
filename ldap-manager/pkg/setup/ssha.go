package setup

import (
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // SHA1 is mandated by the LDAP SSHA spec (RFC 3112)
	"encoding/base64"
)

const (
	sshaPrefix = "{SSHA}"
	saltLen    = 8
	sha1Len    = 20
)

// generateSSHA returns an SSHA hash string: {SSHA}<base64(sha1(password+salt)+salt)>.
func generateSSHA(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	h := sha1.New() //nolint:gosec // SHA1 is mandated by the LDAP SSHA spec (RFC 3112)
	h.Write([]byte(password))
	h.Write(salt)
	digest := h.Sum(nil)

	payload := make([]byte, 0, sha1Len+saltLen)
	payload = append(payload, digest...)
	payload = append(payload, salt...)

	return sshaPrefix + base64.StdEncoding.EncodeToString(payload), nil
}
