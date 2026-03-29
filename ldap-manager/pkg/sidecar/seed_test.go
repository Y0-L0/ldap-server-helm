package sidecar

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

type addedEntry struct {
	dn    string
	attrs map[string][]string
}

type fakeSeedBackend struct {
	entries  []addedEntry
	addErr   error
	addCalls int
}

func (f *fakeSeedBackend) Check(_ context.Context) error { return nil }

func (f *fakeSeedBackend) Add(dn string, attrs map[string][]string) error {
	f.addCalls++
	if f.addErr != nil {
		return f.addErr
	}
	f.entries = append(f.entries, addedEntry{dn: dn, attrs: attrs})
	return nil
}

func (s *Unittest) TestSeed_SkipsWhenSentinelExists() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()
	s.WriteFile(filepath.Join(dataDir, sentinelFile), []byte("done"))

	backend := &fakeSeedBackend{}
	err := seed(backend, seedDir, dataDir)

	s.Require().NoError(err)
	s.Require().Equal(0, backend.addCalls)
}

func (s *Unittest) TestSeed_SeedsAndCreatesSentinel() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()
	s.WriteFile(filepath.Join(seedDir, "base.ldif"), []byte(`dn: dc=example,dc=org
objectClass: top
objectClass: dcObject
objectClass: organization
o: Example Organization
dc: example

dn: ou=people,dc=example,dc=org
objectClass: top
objectClass: organizationalUnit
ou: people
`))

	backend := &fakeSeedBackend{}
	err := seed(backend, seedDir, dataDir)

	s.Require().NoError(err)
	s.Require().Equal(2, backend.addCalls)

	s.Require().Equal("dc=example,dc=org", backend.entries[0].dn)
	s.Require().Equal([]string{"top", "dcObject", "organization"}, backend.entries[0].attrs["objectClass"])
	s.Require().Equal([]string{"Example Organization"}, backend.entries[0].attrs["o"])

	s.Require().Equal("ou=people,dc=example,dc=org", backend.entries[1].dn)

	// Sentinel file should exist.
	_, err = os.Stat(filepath.Join(dataDir, sentinelFile))
	s.Require().NoError(err)
}

func (s *Unittest) TestSeed_PropagatesLDAPError() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()
	s.WriteFile(filepath.Join(seedDir, "base.ldif"), []byte(`dn: dc=example,dc=org
objectClass: top
`))

	backend := &fakeSeedBackend{addErr: errors.New("ldap: connection refused")}
	err := seed(backend, seedDir, dataDir)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "ldap: connection refused")

	// Sentinel should NOT be created on error.
	_, err = os.Stat(filepath.Join(dataDir, sentinelFile))
	s.Require().True(os.IsNotExist(err))
}

func (s *Unittest) TestSeed_EmptySeedDir() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()

	backend := &fakeSeedBackend{}
	err := seed(backend, seedDir, dataDir)

	s.Require().NoError(err)
	s.Require().Equal(0, backend.addCalls)

	// Sentinel should be created even with no LDIF files.
	_, err = os.Stat(filepath.Join(dataDir, sentinelFile))
	s.Require().NoError(err)
}

func (s *Unittest) TestSeed_MultipleLDIFFiles() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()
	s.WriteFile(filepath.Join(seedDir, "01-base.ldif"), []byte(`dn: dc=example,dc=org
objectClass: top
`))
	s.WriteFile(filepath.Join(seedDir, "02-people.ldif"), []byte(`dn: ou=people,dc=example,dc=org
objectClass: organizationalUnit
`))

	backend := &fakeSeedBackend{}
	err := seed(backend, seedDir, dataDir)

	s.Require().NoError(err)
	s.Require().Equal(2, backend.addCalls)
}

func (s *Unittest) TestSeed_MalformedLineSkipped() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()
	s.WriteFile(filepath.Join(seedDir, "base.ldif"), []byte(`dn: dc=example,dc=org
objectClass: top
garbage_no_colon
dc: example
`))

	backend := &fakeSeedBackend{}
	err := seed(backend, seedDir, dataDir)

	s.Require().NoError(err)
	s.Require().Equal(1, backend.addCalls)
	s.Require().Equal([]string{"top"}, backend.entries[0].attrs["objectClass"])
	s.Require().Equal([]string{"example"}, backend.entries[0].attrs["dc"])
}

func (s *Unittest) TestSeed_UnreadableLDIFFile() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()

	// Create a directory named trick.ldif — glob matches it, os.Open fails.
	s.Require().NoError(os.Mkdir(filepath.Join(seedDir, "trick.ldif"), 0o750))

	backend := &fakeSeedBackend{}
	err := seed(backend, seedDir, dataDir)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "trick.ldif")
}

func (s *Unittest) TestSeed_SentinelWriteError() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()

	// Make dataDir read-only so sentinel write fails.
	s.Require().NoError(os.Chmod(dataDir, 0o555)) //nolint:gosec // intentionally restrictive for test
	s.T().Cleanup(func() {
		_ = os.Chmod(dataDir, 0o700) //nolint:gosec // restore for cleanup
	})

	backend := &fakeSeedBackend{}
	err := seed(backend, seedDir, dataDir)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "writing sentinel")
}

func (s *Unittest) TestSeed_IgnoresNonLDIFFiles() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()
	s.WriteFile(filepath.Join(seedDir, "readme.txt"), []byte("not an ldif"))
	s.WriteFile(filepath.Join(seedDir, "base.ldif"), []byte(`dn: dc=example,dc=org
objectClass: top
`))

	backend := &fakeSeedBackend{}
	err := seed(backend, seedDir, dataDir)

	s.Require().NoError(err)
	s.Require().Equal(1, backend.addCalls)
}
