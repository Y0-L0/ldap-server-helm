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

// fullConfig returns a Config with all required fields populated.
// Individual tests override only the fields they care about.
func fullConfig() Config {
	return Config{
		LogLevel:    "info",
		DataDir:     "/var/lib/ldap",
		RunDir:      "/var/run/slapd",
		RootpwPath:  "/etc/ldap/auth/rootpw.conf",
		LdifSeedDir: "/seed",
		Connection: ConnectionConfig{
			URI:    "ldapi:///",
			BindDN: "cn=admin,dc=example,dc=org",
		},
	}
}

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
	cfgPath := s.writeConfig(fullConfig())
	ri := initFunc(func(setup.Config) error { return errors.New("boom") })

	var stderr bytes.Buffer
	code := Main(append([]string{cmdName}, append(configArgs(cfgPath), "setup")...), &stderr, ri, stubSidecar)
	s.Require().Equal(1, code)
}

func (s *Unittest) TestMain_InitCustomConfig() {
	tmp := s.T().TempDir()
	s.T().Setenv("LDAP_ADMIN_PW", "secret")

	cfg := fullConfig()
	cfg.DataDir = filepath.Join(tmp, "data")
	cfg.RunDir = filepath.Join(tmp, "run")
	cfg.RootpwPath = filepath.Join(tmp, "rootpw.conf")
	cfgPath := s.writeConfig(cfg)

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

	cfg := fullConfig()
	cfg.DataDir = "/custom/data"
	cfg.LdifSeedDir = "/custom/seed"
	cfg.Connection = ConnectionConfig{
		URI:    "ldap://remote:389",
		BindDN: "cn=admin,dc=test,dc=org",
	}
	cfgPath := s.writeConfig(cfg)

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
	s.Require().Equal(healthAddr, gotCfg.HealthAddr)
	s.Require().Equal("/custom/seed", gotCfg.SeedDir)
	s.Require().Equal("/custom/data", gotCfg.DataDir)
	s.Require().Equal("ldap://remote:389", gotLDAP.uri)
	s.Require().Equal("cn=admin,dc=test,dc=org", gotLDAP.bindDN)
	s.Require().Equal("pw", gotLDAP.bindPW)
}
