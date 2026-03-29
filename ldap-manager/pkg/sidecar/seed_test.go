package sidecar

import (
	"errors"
	"os"
	"path/filepath"
)

type fakeLDAPSeeder struct {
	entries  []seedEntry
	addErr   error
	addCalls int
}

type seedEntry struct {
	dn    string
	attrs map[string][]string
}

func (f *fakeLDAPSeeder) Add(dn string, attrs map[string][]string) error {
	f.addCalls++
	if f.addErr != nil {
		return f.addErr
	}
	f.entries = append(f.entries, seedEntry{dn: dn, attrs: attrs})
	return nil
}

func (s *Unittest) TestSeed_SkipsWhenSentinelExists() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()
	s.WriteFile(filepath.Join(dataDir, sentinelFile), []byte("done"))

	seeder := &fakeLDAPSeeder{}
	err := seed(seeder, seedDir, dataDir)

	s.Require().NoError(err)
	s.Require().Equal(0, seeder.addCalls)
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

	seeder := &fakeLDAPSeeder{}
	err := seed(seeder, seedDir, dataDir)

	s.Require().NoError(err)
	s.Require().Equal(2, seeder.addCalls)

	s.Require().Equal("dc=example,dc=org", seeder.entries[0].dn)
	s.Require().Equal([]string{"top", "dcObject", "organization"}, seeder.entries[0].attrs["objectClass"])
	s.Require().Equal([]string{"Example Organization"}, seeder.entries[0].attrs["o"])

	s.Require().Equal("ou=people,dc=example,dc=org", seeder.entries[1].dn)

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

	seeder := &fakeLDAPSeeder{addErr: errors.New("ldap: connection refused")}
	err := seed(seeder, seedDir, dataDir)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "ldap: connection refused")

	// Sentinel should NOT be created on error.
	_, err = os.Stat(filepath.Join(dataDir, sentinelFile))
	s.Require().True(os.IsNotExist(err))
}

func (s *Unittest) TestSeed_EmptySeedDir() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()

	seeder := &fakeLDAPSeeder{}
	err := seed(seeder, seedDir, dataDir)

	s.Require().NoError(err)
	s.Require().Equal(0, seeder.addCalls)

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

	seeder := &fakeLDAPSeeder{}
	err := seed(seeder, seedDir, dataDir)

	s.Require().NoError(err)
	s.Require().Equal(2, seeder.addCalls)
}

func (s *Unittest) TestSeed_MalformedLineSkipped() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()
	s.WriteFile(filepath.Join(seedDir, "base.ldif"), []byte(`dn: dc=example,dc=org
objectClass: top
garbage_no_colon
dc: example
`))

	seeder := &fakeLDAPSeeder{}
	err := seed(seeder, seedDir, dataDir)

	s.Require().NoError(err)
	s.Require().Equal(1, seeder.addCalls)
	s.Require().Equal([]string{"top"}, seeder.entries[0].attrs["objectClass"])
	s.Require().Equal([]string{"example"}, seeder.entries[0].attrs["dc"])
}

func (s *Unittest) TestSeed_UnreadableLDIFFile() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()

	// Create a directory named trick.ldif — glob matches it, os.Open fails.
	s.Require().NoError(os.Mkdir(filepath.Join(seedDir, "trick.ldif"), 0o750))

	seeder := &fakeLDAPSeeder{}
	err := seed(seeder, seedDir, dataDir)

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

	seeder := &fakeLDAPSeeder{}
	err := seed(seeder, seedDir, dataDir)

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

	seeder := &fakeLDAPSeeder{}
	err := seed(seeder, seedDir, dataDir)

	s.Require().NoError(err)
	s.Require().Equal(1, seeder.addCalls)
}
