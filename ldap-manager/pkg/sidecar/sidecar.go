// Package sidecar implements the ldap-manager sidecar logic: health checks and LDAP seeding.
package sidecar

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type (
	checkFunc func(context.Context) error
	addFunc   func(string, map[string][]string) error
)

// Config holds the sidecar runtime configuration.
type Config struct {
	HealthAddr string
	SeedDir    string
	DataDir    string
	PollDelay  time.Duration
}

// Run starts the sidecar. Blocks until ctx is cancelled.
func Run(ctx context.Context, cfg Config, check checkFunc, add addFunc) error {
	srv := newHealthServer(cfg.HealthAddr, check)

	go func() {
		slog.Info("starting health server", "addr", cfg.HealthAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("health server failed", "error", err)
		}
	}()

	if err := waitForSlapd(ctx, check, cfg.PollDelay); err != nil {
		return fmt.Errorf("waiting for slapd: %w", err)
	}

	if err := seed(add, cfg.SeedDir, cfg.DataDir); err != nil {
		return fmt.Errorf("seeding: %w", err)
	}

	slog.Info("health check API running", "addr", cfg.HealthAddr)
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutting down health server: %w", err)
	}

	return nil
}

func waitForSlapd(ctx context.Context, check checkFunc, delay time.Duration) error {
	for {
		if err := check(ctx); err == nil {
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
