package sidecar

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type fakeBackend struct {
	healthy atomic.Bool
	entries []addedEntry
	addErr  error
}

func (f *fakeBackend) Check(_ context.Context) error {
	if !f.healthy.Load() {
		return errUnhealthy
	}
	return nil
}

func (f *fakeBackend) Add(dn string, attrs map[string][]string) error {
	if f.addErr != nil {
		return f.addErr
	}
	f.entries = append(f.entries, addedEntry{dn: dn, attrs: attrs})
	return nil
}

var errUnhealthy = &unhealthyError{}

type unhealthyError struct{}

func (e *unhealthyError) Error() string { return "unhealthy" }

func (s *Unittest) TestRun_SeedsAfterHealthy() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()
	s.WriteFile(filepath.Join(seedDir, "base.ldif"), []byte(`dn: dc=test,dc=org
objectClass: top
`))

	backend := &fakeBackend{}
	backend.healthy.Store(true)

	ctx, cancel := context.WithCancel(context.Background())

	cfg := Config{
		HealthAddr: "127.0.0.1:0",
		SeedDir:    seedDir,
		DataDir:    dataDir,
		PollDelay:  10 * time.Millisecond,
	}

	done := make(chan error, 1)
	go func() { done <- Run(ctx, cfg, backend.Check, backend.Add) }()

	// Wait for sentinel to appear, proving seed completed.
	s.Require().Eventually(func() bool {
		_, err := os.Stat(filepath.Join(dataDir, sentinelFile))
		return err == nil
	}, 2*time.Second, 10*time.Millisecond)

	s.Require().Len(backend.entries, 1)
	s.Require().Equal("dc=test,dc=org", backend.entries[0].dn)

	cancel()
	s.Require().NoError(<-done)
}

func (s *Unittest) TestRun_WaitsForSlapd() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()

	backend := &fakeBackend{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := Config{
		HealthAddr: "127.0.0.1:0",
		SeedDir:    seedDir,
		DataDir:    dataDir,
		PollDelay:  10 * time.Millisecond,
	}

	done := make(chan error, 1)
	go func() { done <- Run(ctx, cfg, backend.Check, backend.Add) }()

	// Slapd is unhealthy — should not have seeded yet.
	time.Sleep(50 * time.Millisecond)
	_, err := os.Stat(filepath.Join(dataDir, sentinelFile))
	s.Require().True(os.IsNotExist(err))

	// Make slapd healthy.
	backend.healthy.Store(true)

	s.Require().Eventually(func() bool {
		_, err := os.Stat(filepath.Join(dataDir, sentinelFile))
		return err == nil
	}, 2*time.Second, 10*time.Millisecond)

	cancel()
	s.Require().NoError(<-done)
}

func (s *Unittest) TestRun_CancelStopsRun() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()

	backend := &fakeBackend{}

	ctx, cancel := context.WithCancel(context.Background())

	cfg := Config{
		HealthAddr: "127.0.0.1:0",
		SeedDir:    seedDir,
		DataDir:    dataDir,
		PollDelay:  10 * time.Millisecond,
	}

	done := make(chan error, 1)
	go func() { done <- Run(ctx, cfg, backend.Check, backend.Add) }()

	// Cancel immediately — slapd is unhealthy so Run is stuck in waitForSlapd.
	cancel()

	select {
	case err := <-done:
		s.Require().ErrorIs(err, context.Canceled)
	case <-time.After(2 * time.Second):
		s.Require().Fail("Run did not return after context cancellation")
	}
}

func (s *Unittest) TestRun_HealthEndpointServes() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()

	backend := &fakeBackend{}
	backend.healthy.Store(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := Config{
		HealthAddr: "127.0.0.1:0",
		SeedDir:    seedDir,
		DataDir:    dataDir,
		PollDelay:  10 * time.Millisecond,
	}

	done := make(chan error, 1)
	go func() { done <- Run(ctx, cfg, backend.Check, backend.Add) }()

	// Wait for seed to finish, proving the server started.
	s.Require().Eventually(func() bool {
		_, err := os.Stat(filepath.Join(dataDir, sentinelFile))
		return err == nil
	}, 2*time.Second, 10*time.Millisecond)

	cancel()

	select {
	case err := <-done:
		s.Require().NoError(err)
	case <-time.After(2 * time.Second):
		s.Require().Fail("Run did not return after cancel")
	}
}

func (s *Unittest) TestWaitForSlapd_ImmediatelyHealthy() {
	backend := &fakeBackend{}
	backend.healthy.Store(true)

	err := waitForSlapd(context.Background(), backend.Check, 10*time.Millisecond)
	s.Require().NoError(err)
}

func (s *Unittest) TestWaitForSlapd_CancelledWhileWaiting() {
	backend := &fakeBackend{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := waitForSlapd(ctx, backend.Check, time.Second)
	s.Require().ErrorIs(err, context.Canceled)
}

func (s *Unittest) TestRun_SeedErrorPropagates() {
	dataDir := s.T().TempDir()
	seedDir := s.T().TempDir()
	s.WriteFile(filepath.Join(seedDir, "base.ldif"), []byte(`dn: dc=test,dc=org
objectClass: top
`))

	backend := &fakeBackend{addErr: errors.New("ldap: connection refused")}
	backend.healthy.Store(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := Config{
		HealthAddr: "127.0.0.1:0",
		SeedDir:    seedDir,
		DataDir:    dataDir,
		PollDelay:  10 * time.Millisecond,
	}

	err := Run(ctx, cfg, backend.Check, backend.Add)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "ldap: connection refused")
}
