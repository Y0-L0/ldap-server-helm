package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/ldap"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

type ldapConfig struct {
	uri    string
	bindDN string
	bindPW string
}

func parseLDAPConfig() ldapConfig {
	baseDN := os.Getenv("LDAP_BASE_DN")
	return ldapConfig{
		uri:    envOrDefault("LDAP_URI", "ldap://localhost:389/"),
		bindDN: envOrDefault("LDAP_BIND_DN", "cn=admin,"+baseDN),
		bindPW: os.Getenv("LDAP_ADMIN_PW"),
	}
}

func parseSidecarConfig() sidecar.Config {
	return sidecar.Config{
		HealthAddr: envOrDefault("HEALTH_ADDR", ":8080"),
		SeedDir:    envOrDefault("SEED_DIR", "/seed"),
		DataDir:    envOrDefault("LDAP_DATA_DIR", "/var/lib/ldap"),
		PollDelay:  2 * time.Second,
	}
}

func RunSidecar(cfg sidecar.Config, lcfg ldapConfig) error {
	backend := &ldap.RealLDAP{
		URI:    lcfg.uri,
		BindDN: lcfg.bindDN,
		BindPW: lcfg.bindPW,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT,
	)
	defer stop()

	return sidecar.Run(ctx, cfg, backend.Check, backend.Add)
}
