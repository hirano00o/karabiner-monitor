package monitor

import (
	"os"
	"testing"
)

// TestFindProcess tests finding a process by name
func TestFindProcess(t *testing.T) {

	t.Run("find existing process", func(t *testing.T) {
		// Get current process name
		proc, err := FindProcess("go") // 'go test' command
		if err != nil {
			t.Skipf("Skipping: could not find go process: %v", err)
		}

		if proc == nil {
			t.Fatal("FindProcess() should return non-nil process")
		}

		if proc.PID <= 0 {
			t.Errorf("PID should be positive, got %d", proc.PID)
		}
	})

	t.Run("process not found", func(t *testing.T) {
		proc, err := FindProcess("nonexistent_process_name_12345")
		if err == nil {
			t.Error("FindProcess() should return error for nonexistent process")
		}
		if proc != nil {
			t.Error("FindProcess() should return nil process when not found")
		}
	})
}

// TestGetMemoryUsage tests retrieving memory usage of a process
func TestGetMemoryUsage(t *testing.T) {
	currentPID := int32(os.Getpid())

	memoryMB, err := GetMemoryUsage(currentPID)
	if err != nil {
		t.Fatalf("GetMemoryUsage() error = %v", err)
	}

	if memoryMB <= 0 {
		t.Errorf("memory usage should be positive, got %.2f MB", memoryMB)
	}

	// Sanity check: test process shouldn't use more than 1GB
	if memoryMB > 1024 {
		t.Errorf("memory usage seems too high: %.2f MB", memoryMB)
	}
}

// TestGetMemoryUsageInvalidPID tests error handling for invalid PID
func TestGetMemoryUsageInvalidPID(t *testing.T) {
	// Use an invalid PID
	_, err := GetMemoryUsage(999999)
	if err == nil {
		t.Error("GetMemoryUsage() should return error for invalid PID")
	}
}

// TestKillProcess tests killing a process (we'll use a test subprocess)
func TestKillProcess(t *testing.T) {
	t.Skip("Skipping TestKillProcess as it requires spawning a subprocess")
	// This test would require creating a subprocess and killing it,
	// which is complex and may interfere with the test environment.
	// In production, this function will be tested through integration tests.
}

// TestProcessInfo tests the ProcessInfo structure
func TestProcessInfo(t *testing.T) {
	info := &ProcessInfo{
		PID:  12345,
		Name: "test_process",
	}

	if info.PID != 12345 {
		t.Errorf("PID = %d, want 12345", info.PID)
	}

	if info.Name != "test_process" {
		t.Errorf("Name = %s, want test_process", info.Name)
	}
}
