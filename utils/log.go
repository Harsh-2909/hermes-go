package utils

import (
	"log/slog"
	"os"
)

// Logger is the global logger instance used throughout the application.
var Logger *slog.Logger

func init() {
	// Initialize with default settings (Info level)
	Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// InitLogger can be called later to change the logger configuration
func InitLogger(debug bool) {
	var level slog.Level
	if debug {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}
	Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}
