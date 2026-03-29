package cli

import (
	"errors"
	"os"

	initpkg "github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/init"
)

var errMissingAdminPW = errors.New("LDAP_ADMIN_PW is required")

func parseInitConfig() (initpkg.Config, error) {
	adminPW := os.Getenv("LDAP_ADMIN_PW")
	if adminPW == "" {
		return initpkg.Config{}, errMissingAdminPW
	}
	return initpkg.Config{
		DataDir:    envOrDefault("LDAP_DATA_DIR", "/var/lib/ldap"),
		RunDir:     envOrDefault("LDAP_RUN_DIR", "/var/run/slapd"),
		RootpwPath: envOrDefault("LDAP_ROOTPW_PATH", "/etc/ldap/rootpw.conf"),
		AdminPW:    adminPW,
	}, nil
}
