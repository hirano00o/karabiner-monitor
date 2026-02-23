package notifier

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/hirano00o/karabiner-monitor/internal/logger"
)

// getConsoleUserUID returns the UID of the currently logged-in GUI user.
// This is necessary for sending notifications from a LaunchDaemon (root process).
//
// Returns:
//   - uid: The console user's UID
//   - error: Error if unable to determine the console user UID
//
// Example:
//
//	uid, err := getConsoleUserUID()
//	if err != nil {
//		return fmt.Errorf("failed to get console user UID: %w", err)
//	}
func getConsoleUserUID() (string, error) {
	// Get the username first
	cmd := exec.Command("stat", "-f", "%Su", "/dev/console")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get console user: %w", err)
	}
	username := strings.TrimSpace(string(output))

	// Get the UID for the username
	cmd = exec.Command("id", "-u", username)
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get UID for user %s: %w", username, err)
	}
	uid := strings.TrimSpace(string(output))

	return uid, nil
}

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
// Parameters:
//   - processName: Name of the process that was restarted (used in the notification subtitle)
//   - memoryMB: Memory usage in megabytes
//   - killTime: Time when the process was killed
//   - log: Logger instance for debug and error logging
//
// Example:
//
//	err := SendNotification("Karabiner-Core-Service", 50.5, time.Now(), logger)
//	if err != nil {
//		logger.Error("failed to send notification", "error", err)
//	}
func SendNotification(processName string, memoryMB float64, killTime time.Time, log *logger.Logger) error {
	message := FormatMessage(memoryMB, killTime)

	log.Debug("attempting to send notification", "message", message)

	// Get the console user UID (necessary when running as root)
	uid, err := getConsoleUserUID()
	if err != nil {
		log.Debug("failed to get console user UID", "error", err)
		return fmt.Errorf("failed to get console user UID: %w", err)
	}
	log.Debug("console user UID obtained", "uid", uid)

	// Run terminal-notifier as the console user using launchctl asuser
	// This is the recommended way to run GUI commands from a LaunchDaemon
	cmd := exec.Command("launchctl", "asuser", uid,
		"/opt/homebrew/bin/terminal-notifier",
		"-title", "Karabiner Monitor",
		"-subtitle", fmt.Sprintf("%s restarted", processName),
		"-message", message,
		"-sound", "default",
		"-ignoreDnD")

	log.Debug("running notification command", "command", cmd.Path, "args", cmd.Args)

	output, err := cmd.CombinedOutput()
	log.Debug("terminal-notifier execution completed", "output", string(output), "error", err)

	if err != nil {
		log.Debug("terminal-notifier failed, falling back to osascript", "error", err)
		// Fallback to osascript if terminal-notifier is not available
		script := fmt.Sprintf(`display notification "%s" with title "Karabiner Monitor" subtitle "%s restarted"`,
			message, processName)

		fallbackCmd := exec.Command("launchctl", "asuser", uid, "osascript", "-e", script)
		log.Debug("running fallback command", "command", fallbackCmd.Path, "args", fallbackCmd.Args)

		fallbackOutput, fallbackErr := fallbackCmd.CombinedOutput()
		log.Debug("osascript execution completed", "output", string(fallbackOutput), "error", fallbackErr)

		if fallbackErr != nil {
			log.Error("both terminal-notifier and osascript failed", "terminal_notifier_error", err, "osascript_error", fallbackErr)
			return fmt.Errorf("failed to send notification (terminal-notifier: %v, osascript: %v)", err, fallbackErr)
		}
		log.Debug("notification sent successfully via osascript")
		return nil
	}

	log.Debug("notification sent successfully via terminal-notifier")
	return nil
}
