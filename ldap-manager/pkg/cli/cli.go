// Package cli implements the ldap-manager command-line interface.
package cli

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/setup"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

type (
	initFunc    func(setup.Config) error
	sidecarFunc func(sidecar.Config, ldapConfig) error
)

// Main runs the ldap-manager CLI. Returns an exit code.
func Main(args []string, stderr io.Writer, runInit initFunc, runSidecar sidecarFunc) int {
	setupLogging()

	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: ldap-manager <setup|sidecar>")
		return 1
	}

	cmd := args[1]
	var err error

	switch cmd {
	case "setup":
		cfg, parseErr := parseInitConfig()
		if parseErr != nil {
			err = parseErr
		} else {
			err = runInit(cfg)
		}
	case "sidecar":
		err = runSidecar(parseSidecarConfig(), parseLDAPConfig())
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

var errMissingAdminPW = errors.New("LDAP_ADMIN_PW is required")

func parseInitConfig() (setup.Config, error) {
	adminPW := os.Getenv("LDAP_ADMIN_PW")
	if adminPW == "" {
		return setup.Config{}, errMissingAdminPW
	}
	return setup.Config{
		DataDir:    envOrDefault("LDAP_DATA_DIR", "/var/lib/ldap"),
		RunDir:     envOrDefault("LDAP_RUN_DIR", "/var/run/slapd"),
		RootpwPath: envOrDefault("LDAP_ROOTPW_PATH", "/etc/ldap/auth/rootpw.conf"),
		AdminPW:    adminPW,
	}, nil
}
