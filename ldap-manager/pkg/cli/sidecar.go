package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	ldapadapter "github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/ldap"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

func parseSidecarConfig() sidecar.Config {
	return sidecar.Config{
		HealthAddr: envOrDefault("HEALTH_ADDR", ":8080"),
		SeedDir:    envOrDefault("SEED_DIR", "/seed"),
		DataDir:    envOrDefault("LDAP_DATA_DIR", "/var/lib/ldap"),
		PollDelay:  2 * time.Second,
	}
}

func runSidecar() error {
	cfg := parseSidecarConfig()
	baseDN := os.Getenv("LDAP_BASE_DN")

	backend := &ldapadapter.RealLDAP{
		URI:    envOrDefault("LDAP_URI", "ldapi:///"),
		BindDN: envOrDefault("LDAP_BIND_DN", "cn=admin,"+baseDN),
		BindPW: os.Getenv("LDAP_ADMIN_PW"),
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT,
	)
	defer stop()

	return sidecar.Run(ctx, cfg, backend)
}
