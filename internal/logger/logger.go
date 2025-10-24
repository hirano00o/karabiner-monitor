package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/hirano00o/karabiner-monitor/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger wraps slog.Logger for structured logging with rotation support.
// It provides convenient methods for different log levels and automatically
// manages log file rotation based on size and age.
//
// Example:
//
//	cfg := &config.Config{LogMaxSizeMB: 10, LogMaxAgeDays: 7}
//	logger := New("/var/log/app.log", cfg)
//	logger.Info("application started", "version", "1.0.0")
//	logger.Error("failed to connect", "error", err)
type Logger struct {
	*slog.Logger
}

// New creates a new Logger with log rotation configured.
// The log directory is automatically created if it doesn't exist.
//
// Log rotation settings:
//   - MaxSize: Maximum size in megabytes before rotation (from cfg.LogMaxSizeMB)
//   - MaxAge: Maximum number of days to retain old logs (from cfg.LogMaxAgeDays)
//   - Compress: Old logs are not compressed (disabled for simplicity)
//
// The logger uses JSON format for structured logging with timestamp, level, message, and attributes.
//
// Example:
//
//	cfg := &config.Config{LogMaxSizeMB: 10, LogMaxAgeDays: 7}
//	logger := New("~/.logs/app.log", cfg)
//	logger.Info("service started", "port", 8080)
//	// Output (in log file): {"time":"2025-01-15T10:30:00Z","level":"INFO","msg":"service started","port":8080}
func New(logPath string, cfg *config.Config) *Logger {
	// Determine log level based on debug configuration
	logLevel := slog.LevelInfo
	if cfg.Debug {
		logLevel = slog.LevelDebug
	}

	// Create log directory if not exists
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// Fall back to stderr if directory creation fails
		handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		})
		return &Logger{slog.New(handler)}
	}

	// Configure log rotation using lumberjack
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    cfg.LogMaxSizeMB,  // megabytes
		MaxAge:     cfg.LogMaxAgeDays, // days
		MaxBackups: 0,                 // keep all backups within MaxAge
		Compress:   false,             // don't compress old logs
	}

	// Use multi-writer to write to both file and stderr for better observability
	multiWriter := io.MultiWriter(lumberjackLogger, os.Stderr)

	// Create JSON handler for structured logging
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: logLevel,
	})

	return &Logger{slog.New(handler)}
}

// Info logs an informational message with optional key-value pairs.
//
// Example:
//
//	logger.Info("user logged in", "user_id", 123, "ip", "192.168.1.1")
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Debug logs a debug message with optional key-value pairs.
// Debug messages are not logged by default (level is set to Info in New).
//
// Example:
//
//	logger.Debug("processing item", "item_id", 456)
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Warn logs a warning message with optional key-value pairs.
//
// Example:
//
//	logger.Warn("high memory usage", "usage_mb", 450, "threshold_mb", 500)
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message with optional key-value pairs.
//
// Example:
//
//	logger.Error("failed to kill process", "error", err, "pid", pid)
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}
