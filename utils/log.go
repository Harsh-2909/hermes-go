package utils

import (
	"log/slog"
	"os"
)

// Logger is the global logger instance used throughout the application.
var Logger *slog.Logger

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
