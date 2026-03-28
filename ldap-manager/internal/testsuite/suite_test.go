package testsuite

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SelfTest struct {
	LoggingSuite
}

func TestSelf(t *testing.T) { suite.Run(t, new(SelfTest)) }

func (s *SelfTest) TestLogCapture() {
	slog.Info("captured")
	s.Contains(s.logBuf.String(), "captured")
}
