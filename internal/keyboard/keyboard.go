package keyboard

import (
	"context"
	"time"
)

// WaitForIdle waits for the specified duration.
// This function simply waits for the configured duration before returning true.
//
// Since karabiner-monitor runs as a LaunchDaemon (root process), it doesn't need
// to monitor actual keyboard events. Instead, it waits for the configured duration
// to ensure the user has a chance to save their work before the process is killed.
//
// Parameters:
//   - ctx: Context for cancellation (e.g., timeout or manual cancellation)
//   - duration: How long to wait
//
// Returns:
//   - true if the wait completed successfully
//   - false if context was cancelled
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	if WaitForIdle(ctx, 10*time.Second) {
//		log.Println("Waited for 10 seconds")
//	} else {
//		log.Println("Wait was cancelled")
//	}
func WaitForIdle(ctx context.Context, duration time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(duration):
		return true
	}
}
