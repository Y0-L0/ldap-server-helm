package cli

import "log/slog"

func (s *Unittest) TestParseLogLevel() {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"", slog.LevelInfo},
		{"bogus", slog.LevelInfo},
	}
	for _, tc := range tests {
		s.Run(tc.input, func() {
			s.Require().Equal(tc.want, parseLogLevel(tc.input))
		})
	}
}
