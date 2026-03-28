// Package testsuite provides shared test suite helpers with logging and file utilities.
package testsuite

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

// LoggingSuite sets a default slog logger per test for diagnostics.
type LoggingSuite struct {
	suite.Suite

	logBuf bytes.Buffer
}

func (s *LoggingSuite) SetupTest() {
	s.logBuf.Reset()
	handler := slog.NewTextHandler(&s.logBuf, &slog.HandlerOptions{AddSource: false, Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))
}

func (s *LoggingSuite) TearDownTest() {
	if !s.T().Failed() || !testing.Verbose() {
		return
	}
	s.T().Log("=== Captured Production Logs ===\n")
	s.T().Log(s.logBuf.String())
}

// WriteFile creates parent directories as needed and writes data to path.
// Test fails on error.
func (s *LoggingSuite) WriteFile(path string, data []byte) {
	if err := os.MkdirAll( //nolint:gosec // test helper, relaxed permissions are fine
		filepath.Dir(path),
		0o755,
	); err != nil {
		s.Require().NoError(err, "mkdir parent for %s", path)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil { //nolint:gosec // test helper, relaxed permissions are fine
		s.Require().NoError(err, "write file %s", path)
	}
}

// ReadFile reads and returns file contents. Test fails on error.
func (s *LoggingSuite) ReadFile(path string) []byte {
	b, err := os.ReadFile(path)
	s.Require().NoError(err, "read file %s", path)
	return b
}
