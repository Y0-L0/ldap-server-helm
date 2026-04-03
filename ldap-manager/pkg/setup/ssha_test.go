package setup

import (
	"crypto/sha1" //nolint:gosec // SHA1 is mandated by the LDAP SSHA spec (RFC 3112)
	"crypto/subtle"
	"encoding/base64"
	"strings"
)

// verifySSHA checks a password against an SSHA hash (test-only helper).
func verifySSHA(hash, password string) bool {
	if !strings.HasPrefix(hash, sshaPrefix) {
		return false
	}

	decoded, err := base64.StdEncoding.DecodeString(
		strings.TrimPrefix(hash, sshaPrefix),
	)
	if err != nil || len(decoded) <= sha1Len {
		return false
	}

	h := sha1.New() //nolint:gosec // SHA1 is mandated by the LDAP SSHA spec (RFC 3112)
	h.Write([]byte(password))
	h.Write(decoded[sha1Len:])

	return subtle.ConstantTimeCompare(decoded[:sha1Len], h.Sum(nil)) == 1
}

func (s *Unittest) TestSSHAHash_RoundTrip() {
	tests := []struct {
		name     string
		password string
	}{
		{"empty", ""},
		{"simple", "admin"},
		{"unicode", "pässwörd"},
		{"long", strings.Repeat("a", 1024)},
		{"special chars", "p@ss!w0rd#$%^&*()"},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			hash, err := generateSSHA(tc.password)
			s.Require().NoError(err)
			s.Require().True(verifySSHA(hash, tc.password))
		})
	}
}

func (s *Unittest) TestSSHAHash_Format() {
	hash, err := generateSSHA("test")
	s.Require().NoError(err)

	s.Require().True(strings.HasPrefix(hash, "{SSHA}"), "hash should start with {SSHA}")

	// Decode the base64 payload: should be 20-byte SHA1 + 8-byte salt = 28 bytes.
	encoded := strings.TrimPrefix(hash, "{SSHA}")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	s.Require().NoError(err)
	s.Require().Len(decoded, 28, "decoded hash should be 28 bytes (20 SHA1 + 8 salt)")
}

func (s *Unittest) TestSSHAHash_UniqueSalts() {
	hash1, err := generateSSHA("same")
	s.Require().NoError(err)
	hash2, err := generateSSHA("same")
	s.Require().NoError(err)

	s.Require().NotEqual(hash1, hash2, "same password should produce different hashes due to random salt")
}

func (s *Unittest) TestSSHAVerify_WrongPassword() {
	hash, err := generateSSHA("correct")
	s.Require().NoError(err)
	s.Require().False(verifySSHA(hash, "wrong"))
}

func (s *Unittest) TestSSHAVerify_Malformed() {
	s.Require().False(verifySSHA("garbage", "test"))
	s.Require().False(verifySSHA("{SSHA}", "test"))
	s.Require().False(verifySSHA("{SSHA}notbase64!!!", "test"))
	s.Require().False(verifySSHA("{SSHA}dG9vc2hvcnQ=", "test")) // too short payload
}
