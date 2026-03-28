package cli

import (
	"encoding/base64"
	"strings"
)

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
			hash, err := GenerateSSHA(tc.password)
			s.Require().NoError(err)
			s.Require().True(VerifySSHA(hash, tc.password))
		})
	}
}

func (s *Unittest) TestSSHAHash_Format() {
	hash, err := GenerateSSHA("test")
	s.Require().NoError(err)

	s.Require().True(strings.HasPrefix(hash, "{SSHA}"), "hash should start with {SSHA}")

	// Decode the base64 payload: should be 20-byte SHA1 + 8-byte salt = 28 bytes.
	encoded := strings.TrimPrefix(hash, "{SSHA}")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	s.Require().NoError(err)
	s.Require().Len(decoded, 28, "decoded hash should be 28 bytes (20 SHA1 + 8 salt)")
}

func (s *Unittest) TestSSHAHash_UniqueSalts() {
	hash1, err := GenerateSSHA("same")
	s.Require().NoError(err)
	hash2, err := GenerateSSHA("same")
	s.Require().NoError(err)

	s.Require().NotEqual(hash1, hash2, "same password should produce different hashes due to random salt")
}

func (s *Unittest) TestSSHAVerify_WrongPassword() {
	hash, err := GenerateSSHA("correct")
	s.Require().NoError(err)
	s.Require().False(VerifySSHA(hash, "wrong"))
}

func (s *Unittest) TestSSHAVerify_Malformed() {
	s.Require().False(VerifySSHA("garbage", "test"))
	s.Require().False(VerifySSHA("{SSHA}", "test"))
	s.Require().False(VerifySSHA("{SSHA}notbase64!!!", "test"))
	s.Require().False(VerifySSHA("{SSHA}dG9vc2hvcnQ=", "test")) // too short payload
}
