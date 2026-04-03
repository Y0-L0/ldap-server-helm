package setup

import (
	"os"
	"path/filepath"
	"strings"
)

func (s *Unittest) TestRun_CreatesDirectories() {
	tmp := s.T().TempDir()
	cfg := Config{
		DataDir:    filepath.Join(tmp, "data"),
		RunDir:     filepath.Join(tmp, "run"),
		RootpwPath: filepath.Join(tmp, "etc", "rootpw.conf"),
		AdminPW:    "test",
	}

	s.Require().NoError(Run(cfg))

	for _, dir := range []string{"data", "run"} {
		info, err := os.Stat(filepath.Join(tmp, dir))
		s.Require().NoError(err, "directory %s should exist", dir)
		s.Require().True(info.IsDir())
	}
}

func (s *Unittest) TestRun_WritesRootpwConf() {
	tmp := s.T().TempDir()
	rootpwPath := filepath.Join(tmp, "etc", "rootpw.conf")
	cfg := Config{
		DataDir:    filepath.Join(tmp, "data"),
		RunDir:     filepath.Join(tmp, "run"),
		RootpwPath: rootpwPath,
		AdminPW:    "secret",
	}

	s.Require().NoError(Run(cfg))

	content := string(s.ReadFile(rootpwPath))
	s.Require().True(
		strings.HasPrefix(content, "rootpw {SSHA}"),
		"rootpw.conf should start with 'rootpw {SSHA}'",
	)

	hash := strings.TrimPrefix(strings.TrimSpace(content), "rootpw ")
	s.Require().True(verifySSHA(hash, "secret"))
}
