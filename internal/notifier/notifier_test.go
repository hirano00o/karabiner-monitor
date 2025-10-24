package notifier

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hirano00o/karabiner-monitor/internal/config"
	"github.com/hirano00o/karabiner-monitor/internal/logger"
)

// TestFormatMessage tests notification message formatting
func TestFormatMessage(t *testing.T) {
	now := time.Date(2025, 10, 12, 10, 30, 45, 0, time.UTC)
	memoryMB := 75.5

	msg := FormatMessage(memoryMB, now)

	if !strings.Contains(msg, "75.50 MB") {
		t.Errorf("message should contain memory usage, got: %s", msg)
	}

	if !strings.Contains(msg, "2025-10-12") {
		t.Errorf("message should contain date, got: %s", msg)
	}

	if !strings.Contains(msg, "10:30:45") {
		t.Errorf("message should contain time, got: %s", msg)
	}
}

// TestSendNotification tests sending notification (mocked execution)
func TestSendNotification(t *testing.T) {
	// This test verifies that SendNotification doesn't panic
	// Actual notification won't be sent in test environment
	t.Skip("Skipping TestSendNotification as it requires osascript and may show actual notifications")

	// Create a temporary logger for testing
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	cfg := &config.Config{
		LogMaxSizeMB:  10,
		LogMaxAgeDays: 7,
		Debug:         true,
	}
	log := logger.New(logPath, cfg)

	memoryMB := 50.5
	killTime := time.Now()

	err := SendNotification(memoryMB, killTime, log)
	if err != nil {
		t.Logf("SendNotification error (expected in test environment): %v", err)
	}
}

// TestNotificationContent tests the content structure
func TestNotificationContent(t *testing.T) {
	memoryMB := 100.25
	killTime := time.Date(2025, 10, 12, 15, 45, 30, 0, time.UTC)

	msg := FormatMessage(memoryMB, killTime)

	// Verify expected content
	if msg == "" {
		t.Error("message should not be empty")
	}

	// Check for key components
	expectedComponents := []string{
		"Memory usage:",
		"100.25 MB",
		"at",
		"2025-10-12",
		"15:45:30",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(msg, component) {
			t.Errorf("message should contain '%s', got: %s", component, msg)
		}
	}
}
