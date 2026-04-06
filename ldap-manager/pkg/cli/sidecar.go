package cli

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/ldap"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

type LDAPConfig struct {
	uri    string
	bindDN string
	bindPW string
}

// RunSidecar is the real sidecarFunc implementation used in production.
func RunSidecar(cfg sidecar.Config, lcfg LDAPConfig) error {
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
