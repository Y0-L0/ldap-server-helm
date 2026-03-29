// Package cli implements the ldap-manager command-line interface.
package cli

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	initpkg "github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/init"
)

// Commands maps subcommand names to their run functions.
type Commands map[string]func() error

// NewCommands returns the production command map.
func NewCommands() Commands {
	return Commands{
		"init": func() error {
			cfg, err := parseInitConfig()
			if err != nil {
				return err
			}
			return initpkg.Run(cfg)
		},
		"sidecar": func() error {
			return runSidecar(parseSidecarConfig(), parseLDAPConfig())
		},
	}
}

// Main runs the ldap-manager CLI. Returns an exit code.
func Main(args []string, stderr io.Writer, cmds Commands) int {
	setupLogging()

	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: ldap-manager <init|sidecar>")
		return 1
	}

	cmd := args[1]
	fn, ok := cmds[cmd]
	if !ok {
		fmt.Fprintf(stderr, "unknown command: %s\n", cmd)
		return 1
	}

	if err := fn(); err != nil {
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
