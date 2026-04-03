package e2e

import (
	"context"
	"os"
	"path/filepath"
	"time"

	ldapadapter "github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/ldap"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

func (s *E2E) TestSeedFlow() {
	// Fresh directories so no sentinel exists
	seedDataDir := s.T().TempDir()
	seedDir := s.T().TempDir()

	// Write seed LDIF with entries that don't conflict with slapadd-loaded data
	ldif := `dn: cn=seeduser,ou=people,` + baseDN + `
objectClass: inetOrgPerson
cn: seeduser
sn: Seeded
`
	s.Require().NoError(os.WriteFile(
		filepath.Join(seedDir, "seed.ldif"),
		[]byte(ldif),
		0o600,
	))

	// sidecar.Run blocks — run in goroutine with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- sidecar.Run(ctx, sidecar.Config{
			HealthAddr: "127.0.0.1:0",
			SeedDir:    seedDir,
			DataDir:    seedDataDir,
			PollDelay:  100 * time.Millisecond,
		}, s.backend.Check, s.backend.Add)
	}()

	// Wait for sentinel file to appear (seed completed)
	sentinel := filepath.Join(seedDataDir, ".initialized")
	s.waitForFile(sentinel, 10*time.Second)

	// Cancel context to stop sidecar.Run
	cancel()
	s.Require().NoError(<-errCh)

	// Verify the sentinel was written
	_, err := os.Stat(sentinel)
	s.Require().NoError(err, "sentinel file should exist")
}

func (s *E2E) TestSeedSkipsWhenAlreadyInitialized() {
	seedDataDir := s.T().TempDir()
	seedDir := s.T().TempDir()

	// Pre-create sentinel
	sentinel := filepath.Join(seedDataDir, ".initialized")
	s.Require().NoError(os.WriteFile(sentinel, []byte("initialized\n"), 0o600))

	// Write seed LDIF that would fail if actually loaded (duplicate base DN)
	ldif := `dn: ` + baseDN + `
objectClass: top
objectClass: dcObject
objectClass: organization
o: Example Organization
dc: example
`
	s.Require().NoError(os.WriteFile(
		filepath.Join(seedDir, "seed.ldif"),
		[]byte(ldif),
		0o600,
	))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a separate backend to avoid interfering with other tests
	backend := &ldapadapter.RealLDAP{
		URI:    s.ldapURI,
		BindDN: adminDN,
		BindPW: adminPW,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- sidecar.Run(ctx, sidecar.Config{
			HealthAddr: "127.0.0.1:0",
			SeedDir:    seedDir,
			DataDir:    seedDataDir,
			PollDelay:  100 * time.Millisecond,
		}, backend.Check, backend.Add)
	}()

	// Give sidecar time to start and (not) seed
	time.Sleep(500 * time.Millisecond)
	cancel()
	s.Require().NoError(<-errCh)
}

func (s *E2E) waitForFile(path string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	s.Require().Failf("timeout", "file %s did not appear within %s", path, timeout)
}
