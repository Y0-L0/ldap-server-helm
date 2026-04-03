package e2e

import (
	"context"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/ldap"
)

func (s *E2E) TestCheckRootDSE() {
	err := s.backend.Check(context.Background())
	s.Require().NoError(err)
}

func (s *E2E) TestCheckUnreachable() {
	bad := &ldap.RealLDAP{
		URI:    "ldap://127.0.0.1:1",
		BindDN: adminDN,
		BindPW: adminPW,
	}
	err := bad.Check(context.Background())
	s.Require().Error(err)
}
