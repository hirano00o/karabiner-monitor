package keyboard

import (
	"context"
	"testing"
	"time"
)

// TestCheckAccessibilityPermission tests accessibility permission check
func TestCheckAccessibilityPermission(t *testing.T) {
	// This test just verifies the function doesn't panic
	// Actual result depends on system accessibility settings
	hasPermission := CheckAccessibilityPermission()
	t.Logf("Accessibility permission: %v", hasPermission)
}

// TestWaitForIdle tests waiting for idle keyboard state
func TestWaitForIdle(t *testing.T) {
	t.Run("immediate timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// With very short duration, should return quickly
		wasIdle := WaitForIdle(ctx, 100*time.Millisecond)
		// Result depends on whether keys are being pressed during test
		t.Logf("Was idle: %v", wasIdle)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		wasIdle := WaitForIdle(ctx, 5*time.Second)
		// Should return false due to context cancellation
		if wasIdle {
			t.Error("WaitForIdle should return false when context is cancelled")
		}
	})
}

// TestKeyboardMonitoring tests the basic structure (not actual monitoring)
func TestKeyboardMonitoring(t *testing.T) {
	// This is a structural test to ensure the package can be imported
	// and basic functions exist
	_ = CheckAccessibilityPermission
	_ = WaitForIdle
}
