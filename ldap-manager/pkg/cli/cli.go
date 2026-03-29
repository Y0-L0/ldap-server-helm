// Package cli implements the ldap-manager command-line interface.
package cli

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	initpkg "github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/init"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

// App holds the injected run functions for each subcommand.
type App struct {
	RunInit    func(initpkg.Config) error
	RunSidecar func(sidecar.Config, ldapConfig) error
}

// NewApp returns the production App wiring.
func NewApp() App {
	return App{
		RunInit:    initpkg.Run,
		RunSidecar: runSidecar,
	}
}

// Main runs the ldap-manager CLI. Returns an exit code.
func Main(args []string, stderr io.Writer, app App) int {
	setupLogging()

	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: ldap-manager <init|sidecar>")
		return 1
	}

	cmd := args[1]
	var err error

	switch cmd {
	case "init":
		cfg, parseErr := parseInitConfig()
		if parseErr != nil {
			err = parseErr
		} else {
			err = app.RunInit(cfg)
		}
	case "sidecar":
		err = app.RunSidecar(parseSidecarConfig(), parseLDAPConfig())
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", cmd)
		return 1
	}

	if err != nil {
		slog.Error(cmd+" failed", "error", err)
		return 1
	}

	return 0
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
