package notifier

import (
	"fmt"
	"os/exec"
	"time"
)

// FormatMessage creates a notification message with memory usage and timestamp.
// The message includes the memory usage in megabytes and the time when the process was killed.
//
// Example:
//
//	now := time.Now()
//	msg := FormatMessage(75.5, now)
//	// Output: "Memory usage: 75.50 MB at 2025-10-12 10:30:45"
func FormatMessage(memoryMB float64, killTime time.Time) string {
	return fmt.Sprintf(
		"Memory usage: %.2f MB at %s",
		memoryMB,
		killTime.Format("2006-01-02 15:04:05"),
	)
}

// SendNotification sends a macOS notification using terminal-notifier.
// The notification includes the title, subtitle, and message with memory usage and timestamp.
//
// Requirements:
//   - terminal-notifier installed (brew install terminal-notifier)
//   - For LaunchDaemon usage, terminal-notifier works better than osascript
//
// The notification is sent asynchronously and this function returns immediately
// after executing the terminal-notifier command.
//
// Example:
//
//	err := SendNotification(50.5, time.Now())
//	if err != nil {
//		log.Printf("failed to send notification: %v", err)
//	}
func SendNotification(memoryMB float64, killTime time.Time) error {
	message := FormatMessage(memoryMB, killTime)

	// Try terminal-notifier first (recommended for LaunchDaemon)
	cmd := exec.Command("terminal-notifier",
		"-title", "Karabiner Monitor",
		"-subtitle", "karabiner_grabber restarted",
		"-message", message,
		"-sound", "default")

	if err := cmd.Run(); err != nil {
		// Fallback to osascript if terminal-notifier is not available
		script := fmt.Sprintf(`display notification "%s" with title "Karabiner Monitor" subtitle "karabiner_grabber restarted"`,
			message)

		fallbackCmd := exec.Command("osascript", "-e", script)
		if fallbackErr := fallbackCmd.Run(); fallbackErr != nil {
			return fmt.Errorf("failed to send notification (terminal-notifier: %v, osascript: %v)", err, fallbackErr)
		}
	}

	return nil
}
