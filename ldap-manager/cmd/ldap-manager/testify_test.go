package main

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/jlohmer/ldap-manager/internal/testsuite"
)

var update = flag.Bool("update", false, "update golden files")

type GoldenTest struct {
	testsuite.LoggingSuite

	update bool
}

func (s *GoldenTest) SetupSuite() {
	s.update = *update
}

func TestGolden(t *testing.T) { suite.Run(t, new(GoldenTest)) }

func (s *GoldenTest) goldenFile(name string) string {
	return filepath.Join("testdata", name)
}
