// Package log provides a centralized logging configuration for mosychlos
package log

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// LogFormat represents the output format of logs
type LogFormat string

const (
	// Text outputs logs in a human-readable text format
	Text LogFormat = "text"
	// JSON outputs logs in structured JSON format for machine processing
	JSON LogFormat = "json"
)

// LogLevel represents the minimum severity of logs to output
type LogLevel string

const (
	// Debug includes detailed information for troubleshooting
	Debug LogLevel = "debug"
	// Info includes general operational information
	Info LogLevel = "info"
	// Warn includes potential issues that don't prevent operation
	Warn LogLevel = "warn"
	// Error includes errors that prevent specific operations
	Error LogLevel = "error"
)

// Config holds the configuration for the logger
type Config struct {
	// Level is the minimum log level to output
	Level LogLevel
	// Format is the output format (text or json)
	Format LogFormat
	// Output is where the logs are written
	Output io.Writer
	// AddSource adds source code location to log entries when true
	AddSource bool
}

// DefaultConfig returns the default logging configuration
func DefaultConfig() Config {
	return Config{
		Level:     Info,
		Format:    Text,
		Output:    os.Stderr,
		AddSource: false,
	}
}

// Init initializes the logger with the default configuration and environment overrides
func Init() {
	config := DefaultConfig()

	// Check for environment variable overrides
	if envLevel := os.Getenv("SLOG_LEVEL"); envLevel != "" {
		config.Level = LogLevel(strings.ToLower(envLevel))
	}

	if envFormat := os.Getenv("SLOG_FORMAT"); envFormat != "" {
		config.Format = LogFormat(strings.ToLower(envFormat))
	}

	if os.Getenv("SLOG_SOURCE") == "true" {
		config.AddSource = true
	}

	InitWithConfig(config)
}

// InitWithConfig initializes the logger with a specific configuration
func InitWithConfig(config Config) {
	var level slog.Level
	switch config.Level {
	case Debug:
		level = slog.LevelDebug
	case Info:
		level = slog.LevelInfo
	case Warn:
		level = slog.LevelWarn
	case Error:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
	}

	var handler slog.Handler
	switch config.Format {
	case JSON:
		handler = slog.NewJSONHandler(config.Output, opts)
	default:
		handler = slog.NewTextHandler(config.Output, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// With creates a new logger with the given attributes added to each log entry
func With(attrs ...any) *slog.Logger {
	return slog.With(attrs...)
}

// WithGroup creates a new logger with the given group added to each log entry
func WithGroup(name string) *slog.Logger {
	return slog.Default().WithGroup(name)
}
