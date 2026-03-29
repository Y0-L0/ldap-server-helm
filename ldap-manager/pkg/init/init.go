// Package init implements the ldap-manager init container logic.
package init

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Config holds the init container configuration.
type Config struct {
	DataDir    string
	RunDir     string
	RootpwPath string
	AdminPW    string
}

// Run executes the init container logic: creates directories and writes rootpw.conf.
func Run(cfg Config) error {
	for _, dir := range []string{cfg.DataDir, cfg.RunDir} {
		slog.Info("creating directory", "path", dir)
		if mkErr := os.MkdirAll(dir, 0o750); mkErr != nil {
			return fmt.Errorf("creating %s: %w", dir, mkErr)
		}
	}

	hash, err := generateSSHA(cfg.AdminPW)
	if err != nil {
		return fmt.Errorf("generating SSHA hash: %w", err)
	}

	slog.Info("writing rootpw.conf", "path", cfg.RootpwPath)
	if err := os.MkdirAll(filepath.Dir(cfg.RootpwPath), 0o750); err != nil {
		return fmt.Errorf("creating parent for %s: %w", cfg.RootpwPath, err)
	}
	if err := os.WriteFile(
		cfg.RootpwPath,
		[]byte("rootpw "+hash+"\n"),
		0o600,
	); err != nil {
		return fmt.Errorf("writing %s: %w", cfg.RootpwPath, err)
	}

	slog.Info("init complete")
	return nil
}
