package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func (s *Unittest) TestInit_CreatesDirectories() {
	tmp := s.T().TempDir()
	s.T().Setenv("LDAP_ADMIN_PW", "test")
	s.T().Setenv("LDAP_DATA_DIR", filepath.Join(tmp, "data"))
	s.T().Setenv("LDAP_RUN_DIR", filepath.Join(tmp, "run"))
	s.T().Setenv("LDAP_ROOTPW_PATH", filepath.Join(tmp, "etc", "rootpw.conf"))
	s.T().Setenv("LDAP_CONF_DIR", filepath.Join(tmp, "conf.d"))

	s.Require().NoError(runInit())

	for _, dir := range []string{"data", "run", "conf.d"} {
		info, err := os.Stat(filepath.Join(tmp, dir))
		s.Require().NoError(err, "directory %s should exist", dir)
		s.Require().True(info.IsDir())
	}
}

func (s *Unittest) TestInit_WritesRootpwConf() {
	tmp := s.T().TempDir()
	rootpwPath := filepath.Join(tmp, "etc", "rootpw.conf")
	s.T().Setenv("LDAP_ADMIN_PW", "secret")
	s.T().Setenv("LDAP_DATA_DIR", filepath.Join(tmp, "data"))
	s.T().Setenv("LDAP_RUN_DIR", filepath.Join(tmp, "run"))
	s.T().Setenv("LDAP_ROOTPW_PATH", rootpwPath)
	s.T().Setenv("LDAP_CONF_DIR", filepath.Join(tmp, "conf.d"))

	s.Require().NoError(runInit())

	content := string(s.ReadFile(rootpwPath))
	s.Require().True(strings.HasPrefix(content, "rootpw {SSHA}"), "rootpw.conf should start with 'rootpw {SSHA}'")

	hash := strings.TrimPrefix(strings.TrimSpace(content), "rootpw ")
	s.Require().True(VerifySSHA(hash, "secret"))
}

func (s *Unittest) TestInit_WritesEmptyReplicationFragments() {
	tmp := s.T().TempDir()
	confDir := filepath.Join(tmp, "conf.d")
	s.T().Setenv("LDAP_ADMIN_PW", "test")
	s.T().Setenv("LDAP_DATA_DIR", filepath.Join(tmp, "data"))
	s.T().Setenv("LDAP_RUN_DIR", filepath.Join(tmp, "run"))
	s.T().Setenv("LDAP_ROOTPW_PATH", filepath.Join(tmp, "rootpw.conf"))
	s.T().Setenv("LDAP_CONF_DIR", confDir)

	s.Require().NoError(runInit())

	for _, name := range []string{"serverid.conf", "syncrepl-config.conf", "syncrepl-data.conf"} {
		data := s.ReadFile(filepath.Join(confDir, name))
		s.Require().Empty(data, "%s should be empty", name)
	}
}

func (s *Unittest) TestInit_MissingPassword() {
	s.T().Setenv("LDAP_ADMIN_PW", "")
	err := runInit()
	s.Require().ErrorIs(err, errMissingAdminPW)
}
