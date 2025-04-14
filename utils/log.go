package utils

import (
	"log/slog"

	"github.com/pterm/pterm"
)

// Logger is the global logger instance used throughout the application.
var Logger *slog.Logger

// DefaultLogger is the default logger configuration.
// It uses pterm for pretty terminal output.
var DefaultLogger = pterm.DefaultLogger

// DefaultHander is the default slog handler for logging.
// It uses the default logger configuration.
var DefaultHander = pterm.NewSlogHandler(&DefaultLogger)

func init() {
	// Initialize with default handler
	LogWithDefaultHandler()
	// Set the default logger level to Info
	DefaultLogger.Level = pterm.LogLevelInfo
}

// InitLogger can be called later to change the logger configuration
func InitLogger(debug bool) {
	// TODO: Remove InitLogger and its test and add a method to set the logger level.
	// InitLogger is used to set the logger level, it is better to change the method name
	if debug {
		DefaultLogger.Level = pterm.LogLevelDebug
	}
}

// LogWithDefaultHandler initializes the logger with the default handler.
func LogWithDefaultHandler() {
	Logger = slog.New(DefaultHander)
}

// LogWithCustomHandler initializes the logger with a custom handler.
func LogWithCustomHandler(handler slog.Handler) {
	Logger = slog.New(handler)
	// slog.SetLogLoggerLevel(slog.LevelDebug)
}
