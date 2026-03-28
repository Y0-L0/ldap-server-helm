package main

import (
	"bytes"
	"os"
)

func interceptMain(args []string) (restore func(), outBuf *bytes.Buffer, errBuf *bytes.Buffer, code *int) {
	oldExit, oldOut, oldErr := osExit, stdout, stderr
	outBuf, errBuf = &bytes.Buffer{}, &bytes.Buffer{}
	stdout, stderr = outBuf, errBuf

	var exitCode int
	osExit = func(c int) { exitCode = c }
	os.Args = args

	return func() { osExit, stdout, stderr = oldExit, oldOut, oldErr }, outBuf, errBuf, &exitCode
}

func (s *GoldenTest) EqualGoldenFile(goldenFileName string, actual []byte) {
	goldenFilePath := s.goldenFile(goldenFileName)
	if s.update {
		s.WriteFile(goldenFilePath, actual)
	}
	expected := s.ReadFile(goldenFilePath)
	exp, act := string(expected), string(actual)
	s.Require().Equal(exp, act)
}

func (s *GoldenTest) TestCLI_NoArgs() {
	restore, _, _, code := interceptMain([]string{"ldap-manager"})
	defer restore()

	main()

	s.Require().Equal(0, *code)
}

func (s *GoldenTest) TestCLI_Help() {
	restore, stdout, _, code := interceptMain([]string{"ldap-manager", "--help"})
	defer restore()

	main()

	s.Require().Equal(0, *code)
	s.EqualGoldenFile("help.golden.txt", stdout.Bytes())
}

func (s *GoldenTest) TestCLI_InitHelp() {
	restore, stdout, _, code := interceptMain([]string{"ldap-manager", "init", "--help"})
	defer restore()

	main()

	s.Require().Equal(0, *code)
	s.EqualGoldenFile("init-help.golden.txt", stdout.Bytes())
}

func (s *GoldenTest) TestCLI_SidecarHelp() {
	restore, stdout, _, code := interceptMain([]string{"ldap-manager", "sidecar", "--help"})
	defer restore()

	main()

	s.Require().Equal(0, *code)
	s.EqualGoldenFile("sidecar-help.golden.txt", stdout.Bytes())
}

func (s *GoldenTest) TestCLI_InvalidSubcommand() {
	restore, _, stderr, code := interceptMain([]string{"ldap-manager", "bogus"})
	defer restore()

	main()

	s.Require().Equal(1, *code)
	s.EqualGoldenFile("invalid-subcommand.golden.txt", stderr.Bytes())
}
