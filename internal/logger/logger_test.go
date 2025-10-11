package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hirano00o/karabiner-monitor/internal/config"
)

// TestNew tests logger creation with rotation settings
func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := &config.Config{
		LogMaxSizeMB:  10,
		LogMaxAgeDays: 7,
	}

	logger := New(logPath, cfg)
	if logger == nil {
		t.Fatal("New() returned nil")
	}

	// Test logging
	logger.Info("test message", "key", "value")

	// Verify log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("log file should be created")
	}

	// Read and verify log content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logStr := string(content)
	if !strings.Contains(logStr, "test message") {
		t.Errorf("log should contain 'test message', got: %s", logStr)
	}
	// JSON format uses "key":"value" instead of key=value
	if !strings.Contains(logStr, `"key":"value"`) {
		t.Errorf("log should contain '\"key\":\"value\"', got: %s", logStr)
	}
}

// TestLogLevels tests different log levels
func TestLogLevels(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := &config.Config{
		LogMaxSizeMB:  10,
		LogMaxAgeDays: 7,
	}

	logger := New(logPath, cfg)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logStr := string(content)
	if !strings.Contains(logStr, "info message") {
		t.Error("log should contain info message")
	}
	if !strings.Contains(logStr, "warn message") {
		t.Error("log should contain warn message")
	}
	if !strings.Contains(logStr, "error message") {
		t.Error("log should contain error message")
	}
}

// TestLogDirectory tests that log directory is created if not exists
func TestLogDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "nested", "dir", "test.log")

	cfg := &config.Config{
		LogMaxSizeMB:  10,
		LogMaxAgeDays: 7,
	}

	logger := New(logPath, cfg)
	logger.Info("test message")

	// Verify nested directory was created
	if _, err := os.Stat(filepath.Dir(logPath)); os.IsNotExist(err) {
		t.Error("log directory should be created")
	}

	// Verify log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("log file should be created")
	}
}
