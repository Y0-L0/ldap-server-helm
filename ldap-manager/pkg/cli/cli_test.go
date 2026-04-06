package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/setup"
	"github.com/y0-l0/ldap-server-helm/ldap-manager/pkg/sidecar"
)

const cmdName = "ldap-manager"

var (
	stubInit    initFunc    = func(setup.Config) error { return nil }
	stubSidecar sidecarFunc = func(sidecar.Config, ldapConfig) error { return nil }
)

// writeConfig serialises cfg to a temp file and returns its path.
func (s *Unittest) writeConfig(cfg Config) string {
	path := filepath.Join(s.T().TempDir(), "config.json")
	data, err := json.Marshal(cfg)
	s.Require().NoError(err)
	s.WriteFile(path, data)

	return path
}

// configArgs returns --config <path> args for use in Main calls.
func configArgs(path string) []string { return []string{"--config", path} }

func (s *Unittest) TestMain_NoArgs() {
	// Argument validation happens before config loading — no config file needed.
	var stderr bytes.Buffer
	code := Main([]string{cmdName}, &stderr, stubInit, stubSidecar)
	s.Require().Equal(1, code)
	s.Require().Contains(stderr.String(), "usage:")
}

func (s *Unittest) TestMain_InvalidSubcommand() {
	// Argument validation happens before config loading — no config file needed.
	var stderr bytes.Buffer
	code := Main([]string{cmdName, "bogus"}, &stderr, stubInit, stubSidecar)
	s.Require().Equal(1, code)
	s.Require().Contains(stderr.String(), "unknown command: bogus")
}

func (s *Unittest) TestMain_MissingConfigFile() {
	var stderr bytes.Buffer
	code := Main([]string{cmdName, "--config", "/nonexistent/config.json", "setup"}, &stderr, stubInit, stubSidecar)
	s.Require().Equal(1, code)
	s.Require().Contains(stderr.String(), "config:")
}

func (s *Unittest) TestMain_CommandError() {
	s.T().Setenv("LDAP_ADMIN_PW", "test")
	cfgPath := s.writeConfig(Config{})
	ri := initFunc(func(setup.Config) error { return errors.New("boom") })

	var stderr bytes.Buffer
	code := Main(append([]string{cmdName}, append(configArgs(cfgPath), "setup")...), &stderr, ri, stubSidecar)
	s.Require().Equal(1, code)
}

func (s *Unittest) TestMain_InitMissingPassword() {
	s.T().Setenv("LDAP_ADMIN_PW", "")
	cfgPath := s.writeConfig(Config{})

	var stderr bytes.Buffer
	code := Main(append([]string{cmdName}, append(configArgs(cfgPath), "setup")...), &stderr, stubInit, stubSidecar)
	s.Require().Equal(1, code)
}

func (s *Unittest) TestMain_InitCustomConfig() {
	tmp := s.T().TempDir()
	s.T().Setenv("LDAP_ADMIN_PW", "secret")

	cfgPath := s.writeConfig(Config{
		DataDir:    filepath.Join(tmp, "data"),
		RunDir:     filepath.Join(tmp, "run"),
		RootpwPath: filepath.Join(tmp, "rootpw.conf"),
	})

	var got setup.Config
	ri := initFunc(func(cfg setup.Config) error { got = cfg; return nil })

	var stderr bytes.Buffer
	code := Main(append([]string{cmdName}, append(configArgs(cfgPath), "setup")...), &stderr, ri, stubSidecar)
	s.Require().Equal(0, code)
	s.Require().Equal(filepath.Join(tmp, "data"), got.DataDir)
	s.Require().Equal(filepath.Join(tmp, "run"), got.RunDir)
	s.Require().Equal(filepath.Join(tmp, "rootpw.conf"), got.RootpwPath)
	s.Require().Equal("secret", got.AdminPW)
}

func (s *Unittest) TestMain_SidecarCustomConfig() {
	s.T().Setenv("LDAP_ADMIN_PW", "pw")

	cfgPath := s.writeConfig(Config{
		DataDir: "/custom/data",
		Connection: ConnectionConfig{
			URI:    "ldap://remote:389",
			BaseDN: "dc=test,dc=org",
			BindDN: "cn=admin,dc=test,dc=org",
		},
		Sidecar: SidecarConfig{
			HealthAddr: ":9090",
			SeedDir:    "/custom/seed",
		},
	})

	var gotCfg sidecar.Config
	var gotLDAP ldapConfig
	rs := sidecarFunc(func(cfg sidecar.Config, lcfg ldapConfig) error {
		gotCfg = cfg
		gotLDAP = lcfg

		return nil
	})

	var stderr bytes.Buffer
	code := Main(append([]string{cmdName}, append(configArgs(cfgPath), "sidecar")...), &stderr, stubInit, rs)
	s.Require().Equal(0, code)
	s.Require().Equal(":9090", gotCfg.HealthAddr)
	s.Require().Equal("/custom/seed", gotCfg.SeedDir)
	s.Require().Equal("/custom/data", gotCfg.DataDir)
	s.Require().Equal("ldap://remote:389", gotLDAP.uri)
	s.Require().Equal("cn=admin,dc=test,dc=org", gotLDAP.bindDN)
	s.Require().Equal("pw", gotLDAP.bindPW)
}

func (s *Unittest) TestMain_SidecarBindDNDerivedFromBaseDN() {
	cfgPath := s.writeConfig(Config{
		Connection: ConnectionConfig{
			BaseDN: "dc=example,dc=org",
			// BindDN intentionally omitted — should be derived from BaseDN.
		},
	})

	var gotLDAP ldapConfig
	rs := sidecarFunc(func(_ sidecar.Config, lcfg ldapConfig) error {
		gotLDAP = lcfg

		return nil
	})

	var stderr bytes.Buffer
	Main(append([]string{cmdName}, append(configArgs(cfgPath), "sidecar")...), &stderr, stubInit, rs)
	s.Require().Equal("cn=admin,dc=example,dc=org", gotLDAP.bindDN)
}
