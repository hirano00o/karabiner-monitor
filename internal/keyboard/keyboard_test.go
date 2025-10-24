package keyboard

import (
	"context"
	"testing"
	"time"
)

// TestWaitForIdle tests waiting for the specified duration
func TestWaitForIdle(t *testing.T) {
	t.Run("completes after duration", func(t *testing.T) {
		ctx := context.Background()
		start := time.Now()

		// Wait for 100ms
		completed := WaitForIdle(ctx, 100*time.Millisecond)
		elapsed := time.Since(start)

		if !completed {
			t.Error("WaitForIdle should return true when wait completes")
		}

		// Should take approximately 100ms (with some tolerance)
		if elapsed < 90*time.Millisecond || elapsed > 200*time.Millisecond {
			t.Errorf("Wait duration unexpected: got %v, want ~100ms", elapsed)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		completed := WaitForIdle(ctx, 5*time.Second)

		// Should return false due to context cancellation
		if completed {
			t.Error("WaitForIdle should return false when context is cancelled")
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// Try to wait for 5 seconds but context will timeout at 50ms
		completed := WaitForIdle(ctx, 5*time.Second)

		// Should return false due to context timeout
		if completed {
			t.Error("WaitForIdle should return false when context times out")
		}
	})
}
