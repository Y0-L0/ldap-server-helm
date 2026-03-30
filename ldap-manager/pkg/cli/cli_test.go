package cli

import (
	"bytes"
	"errors"
	"path/filepath"

	initpkg "github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/init"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

var (
	stubInit    initFunc    = func(initpkg.Config) error { return nil }
	stubSidecar sidecarFunc = func(sidecar.Config, ldapConfig) error { return nil }
)

func (s *Unittest) TestMain_NoArgs() {
	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager"}, &stderr, stubInit, stubSidecar)
	s.Require().Equal(1, code)
	s.Require().Contains(stderr.String(), "usage:")
}

func (s *Unittest) TestMain_InvalidSubcommand() {
	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "bogus"}, &stderr, stubInit, stubSidecar)
	s.Require().Equal(1, code)
	s.Require().Contains(stderr.String(), "unknown command: bogus")
}

func (s *Unittest) TestMain_CommandError() {
	s.T().Setenv("LDAP_ADMIN_PW", "test")
	ri := initFunc(func(initpkg.Config) error { return errors.New("boom") })

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "init"}, &stderr, ri, stubSidecar)
	s.Require().Equal(1, code)
}

func (s *Unittest) TestMain_InitMissingPassword() {
	s.T().Setenv("LDAP_ADMIN_PW", "")

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "init"}, &stderr, stubInit, stubSidecar)
	s.Require().Equal(1, code)
}

func (s *Unittest) TestMain_InitCustomEnv() {
	tmp := s.T().TempDir()
	s.T().Setenv("LDAP_ADMIN_PW", "secret")
	s.T().Setenv("LDAP_DATA_DIR", filepath.Join(tmp, "data"))
	s.T().Setenv("LDAP_RUN_DIR", filepath.Join(tmp, "run"))
	s.T().Setenv("LDAP_ROOTPW_PATH", filepath.Join(tmp, "rootpw.conf"))

	var got initpkg.Config
	ri := initFunc(func(cfg initpkg.Config) error { got = cfg; return nil })

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "init"}, &stderr, ri, stubSidecar)
	s.Require().Equal(0, code)
	s.Require().Equal(filepath.Join(tmp, "data"), got.DataDir)
	s.Require().Equal(filepath.Join(tmp, "run"), got.RunDir)
	s.Require().Equal(filepath.Join(tmp, "rootpw.conf"), got.RootpwPath)
	s.Require().Equal("secret", got.AdminPW)
}

func (s *Unittest) TestMain_InitDefaults() {
	s.T().Setenv("LDAP_ADMIN_PW", "secret")

	var got initpkg.Config
	ri := initFunc(func(cfg initpkg.Config) error { got = cfg; return nil })

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "init"}, &stderr, ri, stubSidecar)
	s.Require().Equal(0, code)
	s.Require().Equal("/var/lib/ldap", got.DataDir)
	s.Require().Equal("/var/run/slapd", got.RunDir)
	s.Require().Equal("/etc/ldap/rootpw.conf", got.RootpwPath)
	s.Require().Equal("secret", got.AdminPW)
}

func (s *Unittest) TestMain_SidecarCustomEnv() {
	s.T().Setenv("HEALTH_ADDR", ":9090")
	s.T().Setenv("SEED_DIR", "/custom/seed")
	s.T().Setenv("LDAP_DATA_DIR", "/custom/data")
	s.T().Setenv("LDAP_URI", "ldap://remote:389")
	s.T().Setenv("LDAP_BASE_DN", "dc=test,dc=org")
	s.T().Setenv("LDAP_ADMIN_PW", "pw")

	var gotCfg sidecar.Config
	var gotLDAP ldapConfig
	rs := sidecarFunc(func(cfg sidecar.Config, lcfg ldapConfig) error {
		gotCfg = cfg
		gotLDAP = lcfg
		return nil
	})

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "sidecar"}, &stderr, stubInit, rs)
	s.Require().Equal(0, code)
	s.Require().Equal(":9090", gotCfg.HealthAddr)
	s.Require().Equal("/custom/seed", gotCfg.SeedDir)
	s.Require().Equal("/custom/data", gotCfg.DataDir)
	s.Require().Equal("ldap://remote:389", gotLDAP.uri)
	s.Require().Equal("cn=admin,dc=test,dc=org", gotLDAP.bindDN)
	s.Require().Equal("pw", gotLDAP.bindPW)
}

func (s *Unittest) TestMain_SidecarDefaults() {
	var gotCfg sidecar.Config
	var gotLDAP ldapConfig
	rs := sidecarFunc(func(cfg sidecar.Config, lcfg ldapConfig) error {
		gotCfg = cfg
		gotLDAP = lcfg
		return nil
	})

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "sidecar"}, &stderr, stubInit, rs)
	s.Require().Equal(0, code)
	s.Require().Equal(":8080", gotCfg.HealthAddr)
	s.Require().Equal("/seed", gotCfg.SeedDir)
	s.Require().Equal("/var/lib/ldap", gotCfg.DataDir)
	s.Require().Equal("ldapi:///", gotLDAP.uri)
	s.Require().Equal("cn=admin,", gotLDAP.bindDN)
	s.Require().Empty(gotLDAP.bindPW)
}
