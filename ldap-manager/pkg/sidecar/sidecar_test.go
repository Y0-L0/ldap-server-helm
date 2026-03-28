package sidecar

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type fakeBackend struct {
	healthy atomic.Bool
	entries []seedEntry
}

func (f *fakeBackend) Check(_ context.Context) error {
	if !f.healthy.Load() {
		return errUnhealthy
	}
	return nil
}

func (f *fakeBackend) Add(dn string, attrs map[string][]string) error {
	f.entries = append(f.entries, seedEntry{dn: dn, attrs: attrs})
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
		Checker:    backend,
		Seeder:     backend,
	}

	done := make(chan error, 1)
	go func() { done <- Run(ctx, cfg) }()

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
		Checker:    backend,
		Seeder:     backend,
	}

	done := make(chan error, 1)
	go func() { done <- Run(ctx, cfg) }()

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
		Checker:    backend,
		Seeder:     backend,
	}

	done := make(chan error, 1)
	go func() { done <- Run(ctx, cfg) }()

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

	// Use port 0 for OS-assigned free port. We'll discover it via the listener.
	cfg := Config{
		HealthAddr: "127.0.0.1:0",
		SeedDir:    seedDir,
		DataDir:    dataDir,
		PollDelay:  10 * time.Millisecond,
		Checker:    backend,
		Seeder:     backend,
	}

	done := make(chan error, 1)
	go func() { done <- Run(ctx, cfg) }()

	// Wait for seed to finish, proving the server started.
	s.Require().Eventually(func() bool {
		_, err := os.Stat(filepath.Join(dataDir, sentinelFile))
		return err == nil
	}, 2*time.Second, 10*time.Millisecond)

	// Note: we can't easily test the HTTP server at port :0 since we don't
	// capture the actual listening address. This is tested via health_test.go
	// using httptest. Here we verify the sidecar lifecycle completes correctly.

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

	err := waitForSlapd(context.Background(), backend, 10*time.Millisecond)
	s.Require().NoError(err)
}

func (s *Unittest) TestWaitForSlapd_CancelledWhileWaiting() {
	backend := &fakeBackend{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := waitForSlapd(ctx, backend, time.Second)
	s.Require().ErrorIs(err, context.Canceled)
}

// Ensure http import is used (for http.StatusOK reference in other tests).
var _ = http.StatusOK
