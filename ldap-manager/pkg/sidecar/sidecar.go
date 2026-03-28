package sidecar

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Config holds the sidecar runtime configuration and dependencies.
type Config struct {
	HealthAddr string
	SeedDir    string
	DataDir    string
	PollDelay  time.Duration

	Checker LDAPChecker
	Seeder  LDAPSeeder
}

// Run starts the sidecar. Blocks until ctx is cancelled.
func Run(ctx context.Context, cfg Config) error {
	srv := NewHealthServer(cfg.HealthAddr, cfg.Checker)

	go func() {
		slog.Info("starting health server", "addr", cfg.HealthAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("health server failed", "error", err)
		}
	}()

	if err := waitForSlapd(ctx, cfg.Checker, cfg.PollDelay); err != nil {
		return fmt.Errorf("waiting for slapd: %w", err)
	}

	if err := Seed(cfg.Seeder, cfg.SeedDir, cfg.DataDir); err != nil {
		return fmt.Errorf("seeding: %w", err)
	}

	slog.Info("sidecar ready, blocking until context cancelled")
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutting down health server: %w", err)
	}

	return nil
}

func waitForSlapd(ctx context.Context, checker LDAPChecker, delay time.Duration) error {
	for {
		if err := checker.Check(ctx); err == nil {
			slog.Info("slapd is reachable")
			return nil
		}

		slog.Info("waiting for slapd", "retry_in", delay)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}
