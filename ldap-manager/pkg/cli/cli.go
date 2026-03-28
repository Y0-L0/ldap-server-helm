package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/ldap"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

var errMissingAdminPW = errors.New("LDAP_ADMIN_PW is required")

const (
	defaultDataDir    = "/var/lib/ldap"
	defaultRunDir     = "/var/run/slapd"
	defaultRootpwPath = "/etc/ldap/rootpw.conf" //nolint:gosec // not a credential, just a file path
	defaultConfDir    = "/etc/ldap/slapd.conf.d"
	defaultLDAPURI    = "ldapi:///"
	defaultHealthAddr = ":8080"
	defaultSeedDir    = "/seed"
	defaultPollDelay  = 2 * time.Second
)

// Commands maps subcommand names to their run functions.
type Commands map[string]func() error

// NewCommands returns the production command map.
func NewCommands() Commands {
	return Commands{
		"init":    runInit,
		"sidecar": runSidecar,
	}
}

// Main runs the ldap-manager CLI. Returns an exit code.
func Main(args []string, stderr io.Writer, cmds Commands) int {
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

func runInit() error {
	dataDir := envOrDefault("LDAP_DATA_DIR", defaultDataDir)
	runDir := envOrDefault("LDAP_RUN_DIR", defaultRunDir)
	rootpwPath := envOrDefault("LDAP_ROOTPW_PATH", defaultRootpwPath)
	confDir := envOrDefault("LDAP_CONF_DIR", defaultConfDir)
	adminPW := os.Getenv("LDAP_ADMIN_PW")

	if adminPW == "" {
		return errMissingAdminPW
	}

	// Create directories.
	for _, dir := range []string{dataDir, runDir, confDir} {
		slog.Info("creating directory", "path", dir)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}

	// Generate SSHA hash and write rootpw.conf.
	hash, err := GenerateSSHA(adminPW)
	if err != nil {
		return fmt.Errorf("generating SSHA hash: %w", err)
	}

	slog.Info("writing rootpw.conf", "path", rootpwPath)
	if err := os.MkdirAll(
		filepath.Dir(rootpwPath),
		0o750,
	); err != nil {
		return fmt.Errorf("creating parent for %s: %w", rootpwPath, err)
	}
	if err := os.WriteFile(rootpwPath, []byte("rootpw "+hash+"\n"), 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", rootpwPath, err)
	}

	// Write empty replication fragment files.
	for _, name := range []string{"serverid.conf", "syncrepl-config.conf", "syncrepl-data.conf"} {
		path := filepath.Join(confDir, name)
		slog.Info("writing empty replication fragment", "path", path)
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}

	slog.Info("init complete")
	return nil
}

func runSidecar() error {
	ldapURI := envOrDefault("LDAP_URI", defaultLDAPURI)
	healthAddr := envOrDefault("HEALTH_ADDR", defaultHealthAddr)
	seedDir := envOrDefault("SEED_DIR", defaultSeedDir)
	dataDir := envOrDefault("LDAP_DATA_DIR", defaultDataDir)
	baseDN := os.Getenv("LDAP_BASE_DN")
	adminPW := os.Getenv("LDAP_ADMIN_PW")
	bindDN := envOrDefault("LDAP_BIND_DN", "cn=admin,"+baseDN)

	backend := &ldap.RealLDAP{
		URI:    ldapURI,
		BindDN: bindDN,
		BindPW: adminPW,
	}

	cfg := sidecar.Config{
		HealthAddr: healthAddr,
		SeedDir:    seedDir,
		DataDir:    dataDir,
		PollDelay:  defaultPollDelay,
		Checker:    backend,
		Seeder:     backend,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	return sidecar.Run(ctx, cfg)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
