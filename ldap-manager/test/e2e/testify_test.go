package e2e

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/internal/testsuite"
	ldapadapter "github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/ldap"
)

type E2E struct {
	testsuite.LoggingSuite

	backend *ldapadapter.RealLDAP
	slapd   *exec.Cmd
	tmpDir  string
	dataDir string
	seedDir string
	ldapURI string
}

func TestE2E(t *testing.T) { suite.Run(t, new(E2E)) }
