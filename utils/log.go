package utils

import (
	"log/slog"

	"github.com/pterm/pterm"
)

// Logger is the global logger instance used throughout the application.
var Logger *slog.Logger

func init() {
	// Initialize with default settings (Info level)
	handler := pterm.NewSlogHandler(&pterm.DefaultLogger)
	Logger = slog.New(handler)
}

// InitLogger can be called later to change the logger configuration
func InitLogger(debug bool) {
	// var level slog.Level
	if debug {
		pterm.DefaultLogger.Level = pterm.LogLevelDebug
		// level = slog.LevelDebug
	}
	// Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	// 	Level: level,
	// }))
}
