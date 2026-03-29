package cli

import (
	"bytes"
	"errors"
	"path/filepath"

	initpkg "github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/init"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

func stubApp() App {
	return App{
		RunInit:    func(initpkg.Config) error { return nil },
		RunSidecar: func(sidecar.Config, ldapConfig) error { return nil },
	}
}

func (s *Unittest) TestMain_NoArgs() {
	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager"}, &stderr, stubApp())
	s.Require().Equal(1, code)
	s.Require().Contains(stderr.String(), "usage:")
}

func (s *Unittest) TestMain_InvalidSubcommand() {
	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "bogus"}, &stderr, stubApp())
	s.Require().Equal(1, code)
	s.Require().Contains(stderr.String(), "unknown command: bogus")
}

func (s *Unittest) TestMain_DispatchesInit() {
	s.T().Setenv("LDAP_ADMIN_PW", "test")

	var got initpkg.Config
	app := stubApp()
	app.RunInit = func(cfg initpkg.Config) error { got = cfg; return nil }

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "init"}, &stderr, app)
	s.Require().Equal(0, code)
	s.Require().Equal("test", got.AdminPW)
}

func (s *Unittest) TestMain_DispatchesSidecar() {
	var gotCfg sidecar.Config
	var gotLDAP ldapConfig
	app := stubApp()
	app.RunSidecar = func(cfg sidecar.Config, lcfg ldapConfig) error {
		gotCfg = cfg
		gotLDAP = lcfg
		return nil
	}

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "sidecar"}, &stderr, app)
	s.Require().Equal(0, code)
	s.Require().Equal(":8080", gotCfg.HealthAddr)
	s.Require().Equal("ldapi:///", gotLDAP.uri)
}

func (s *Unittest) TestMain_CommandError() {
	app := stubApp()
	app.RunInit = func(initpkg.Config) error { return errors.New("boom") }
	s.T().Setenv("LDAP_ADMIN_PW", "test")

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "init"}, &stderr, app)
	s.Require().Equal(1, code)
}

func (s *Unittest) TestMain_InitMissingPassword() {
	s.T().Setenv("LDAP_ADMIN_PW", "")

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "init"}, &stderr, stubApp())
	s.Require().Equal(1, code)
}

func (s *Unittest) TestParseInitConfig() {
	s.T().Setenv("LDAP_ADMIN_PW", "secret")
	s.T().Setenv("LDAP_DATA_DIR", "/custom/data")
	s.T().Setenv("LDAP_RUN_DIR", "/custom/run")
	s.T().Setenv("LDAP_ROOTPW_PATH", "/custom/rootpw.conf")

	cfg, err := parseInitConfig()
	s.Require().NoError(err)
	s.Require().Equal("/custom/data", cfg.DataDir)
	s.Require().Equal("/custom/run", cfg.RunDir)
	s.Require().Equal("/custom/rootpw.conf", cfg.RootpwPath)
	s.Require().Equal("secret", cfg.AdminPW)
}

func (s *Unittest) TestParseInitConfig_Defaults() {
	s.T().Setenv("LDAP_ADMIN_PW", "secret")

	cfg, err := parseInitConfig()
	s.Require().NoError(err)
	s.Require().Equal("/var/lib/ldap", cfg.DataDir)
	s.Require().Equal("/var/run/slapd", cfg.RunDir)
	s.Require().Equal("/etc/ldap/rootpw.conf", cfg.RootpwPath)
}

func (s *Unittest) TestParseInitConfig_MissingPassword() {
	s.T().Setenv("LDAP_ADMIN_PW", "")
	_, err := parseInitConfig()
	s.Require().ErrorIs(err, errMissingAdminPW)
}

func (s *Unittest) TestParseSidecarConfig() {
	s.T().Setenv("HEALTH_ADDR", ":9090")
	s.T().Setenv("SEED_DIR", "/custom/seed")
	s.T().Setenv("LDAP_DATA_DIR", "/custom/data")

	cfg := parseSidecarConfig()
	s.Require().Equal(":9090", cfg.HealthAddr)
	s.Require().Equal("/custom/seed", cfg.SeedDir)
	s.Require().Equal("/custom/data", cfg.DataDir)
}

func (s *Unittest) TestParseSidecarConfig_Defaults() {
	cfg := parseSidecarConfig()
	s.Require().Equal(":8080", cfg.HealthAddr)
	s.Require().Equal("/seed", cfg.SeedDir)
	s.Require().Equal("/var/lib/ldap", cfg.DataDir)
}

func (s *Unittest) TestParseLDAPConfig() {
	s.T().Setenv("LDAP_URI", "ldap://localhost:389")
	s.T().Setenv("LDAP_BIND_DN", "cn=custom")
	s.T().Setenv("LDAP_ADMIN_PW", "secret")

	cfg := parseLDAPConfig()
	s.Require().Equal("ldap://localhost:389", cfg.uri)
	s.Require().Equal("cn=custom", cfg.bindDN)
	s.Require().Equal("secret", cfg.bindPW)
}

func (s *Unittest) TestParseLDAPConfig_DefaultBindDN() {
	s.T().Setenv("LDAP_BASE_DN", "dc=example,dc=org")

	cfg := parseLDAPConfig()
	s.Require().Equal("ldapi:///", cfg.uri)
	s.Require().Equal("cn=admin,dc=example,dc=org", cfg.bindDN)
}

func (s *Unittest) TestMain_InitParsesEnvVars() {
	tmp := s.T().TempDir()
	s.T().Setenv("LDAP_ADMIN_PW", "secret")
	s.T().Setenv("LDAP_DATA_DIR", filepath.Join(tmp, "data"))
	s.T().Setenv("LDAP_RUN_DIR", filepath.Join(tmp, "run"))
	s.T().Setenv("LDAP_ROOTPW_PATH", filepath.Join(tmp, "rootpw.conf"))

	var got initpkg.Config
	app := stubApp()
	app.RunInit = func(cfg initpkg.Config) error { got = cfg; return nil }

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "init"}, &stderr, app)
	s.Require().Equal(0, code)
	s.Require().Equal(filepath.Join(tmp, "data"), got.DataDir)
	s.Require().Equal(filepath.Join(tmp, "run"), got.RunDir)
	s.Require().Equal(filepath.Join(tmp, "rootpw.conf"), got.RootpwPath)
	s.Require().Equal("secret", got.AdminPW)
}

func (s *Unittest) TestMain_SidecarParsesEnvVars() {
	s.T().Setenv("HEALTH_ADDR", ":9090")
	s.T().Setenv("SEED_DIR", "/custom/seed")
	s.T().Setenv("LDAP_DATA_DIR", "/custom/data")
	s.T().Setenv("LDAP_URI", "ldap://remote:389")
	s.T().Setenv("LDAP_BASE_DN", "dc=test,dc=org")
	s.T().Setenv("LDAP_ADMIN_PW", "pw")

	var gotCfg sidecar.Config
	var gotLDAP ldapConfig
	app := stubApp()
	app.RunSidecar = func(cfg sidecar.Config, lcfg ldapConfig) error {
		gotCfg = cfg
		gotLDAP = lcfg
		return nil
	}

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "sidecar"}, &stderr, app)
	s.Require().Equal(0, code)
	s.Require().Equal(":9090", gotCfg.HealthAddr)
	s.Require().Equal("/custom/seed", gotCfg.SeedDir)
	s.Require().Equal("/custom/data", gotCfg.DataDir)
	s.Require().Equal("ldap://remote:389", gotLDAP.uri)
	s.Require().Equal("cn=admin,dc=test,dc=org", gotLDAP.bindDN)
	s.Require().Equal("pw", gotLDAP.bindPW)
}
