package cli

import (
	"bytes"
	"errors"
)

func (s *Unittest) TestMain_NoArgs() {
	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager"}, &stderr, nil)
	s.Require().Equal(1, code)
	s.Require().Contains(stderr.String(), "usage:")
}

func (s *Unittest) TestMain_InvalidSubcommand() {
	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "bogus"}, &stderr, Commands{})
	s.Require().Equal(1, code)
	s.Require().Contains(stderr.String(), "unknown command: bogus")
}

func (s *Unittest) TestMain_DispatchesCommand() {
	called := false
	var stubErr error
	cmds := Commands{"init": func() error { called = true; return stubErr }}

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "init"}, &stderr, cmds)
	s.Require().Equal(0, code)
	s.Require().True(called)
}

func (s *Unittest) TestMain_CommandError() {
	stubErr := errors.New("boom")
	cmds := Commands{"fail": func() error { return stubErr }}

	var stderr bytes.Buffer
	code := Main([]string{"ldap-manager", "fail"}, &stderr, cmds)
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
