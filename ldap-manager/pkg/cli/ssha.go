// Package cli implements the ldap-manager command-line interface.
package cli

import (
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // SHA1 is mandated by the LDAP SSHA spec (RFC 3112)
	"crypto/subtle"
	"encoding/base64"
	"strings"
)

const (
	sshaPrefix = "{SSHA}"
	saltLen    = 8
	sha1Len    = 20
)

// GenerateSSHA returns an SSHA hash string: {SSHA}<base64(sha1(password+salt)+salt)>.
func GenerateSSHA(password string) (string, error) {
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

// VerifySSHA checks a password against an SSHA hash.
func VerifySSHA(hash, password string) bool {
	if !strings.HasPrefix(hash, sshaPrefix) {
		return false
	}

	encoded := strings.TrimPrefix(hash, sshaPrefix)

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}

	if len(decoded) <= sha1Len {
		return false
	}

	existingDigest := decoded[:sha1Len]
	salt := decoded[sha1Len:]

	h := sha1.New() //nolint:gosec // SHA1 is mandated by the LDAP SSHA spec (RFC 3112)
	h.Write([]byte(password))
	h.Write(salt)
	computedDigest := h.Sum(nil)

	return subtle.ConstantTimeCompare(existingDigest, computedDigest) == 1
}
