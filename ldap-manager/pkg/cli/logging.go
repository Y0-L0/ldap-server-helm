package cli

import (
	"log/slog"
	"os"
	"strings"
)

func setupLogging(level string) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: parseLogLevel(level)})
	slog.SetDefault(slog.New(handler))
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
