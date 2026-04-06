// Package cli implements the ldap-manager command-line interface.
package cli

import (
	"flag"
	"fmt"
	"io"
	"log/slog"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/setup"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

const defaultConfigPath = "/etc/ldap-manager/config.json"

type (
	initFunc    func(setup.Config) error
	sidecarFunc func(sidecar.Config, ldapConfig) error
)

// Main runs the ldap-manager CLI. Returns an exit code.
func Main(args []string, stderr io.Writer, runInit initFunc, runSidecar sidecarFunc) int {
	fset := flag.NewFlagSet("ldap-manager", flag.ContinueOnError)
	fset.SetOutput(stderr)
	configPath := fset.String("config", defaultConfigPath, "path to JSON config file")

	if err := fset.Parse(args[1:]); err != nil {
		return 1
	}

	remaining := fset.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(stderr, "usage: ldap-manager [--config <path>] <setup|sidecar>")
		return 1
	}

	cmd := remaining[0]
	if cmd != "setup" && cmd != "sidecar" {
		fmt.Fprintf(stderr, "unknown command: %s\n", cmd)
		return 1
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintln(stderr, "config:", err)
		return 1
	}

	setupLogging(cfg.LogLevel)

	var cmdErr error

	switch cmd {
	case "setup":
		cmdErr = runInit(cfg.SetupConfig())
	case "sidecar":
		cmdErr = runSidecar(cfg.SidecarConfig(), cfg.ldapCfg())
	}

	if cmdErr != nil {
		slog.Error(cmd+" failed", "error", cmdErr)
		return 1
	}

	return 0
}
