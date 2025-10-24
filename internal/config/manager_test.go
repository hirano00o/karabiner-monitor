package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestNewManager(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Test: NewManager creates manager with default config if file doesn't exist
	manager, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager returned nil manager")
	}

	cfg := manager.Get()
	if cfg.ProcessName != "karabiner_grabber" {
		t.Errorf("expected default process name 'karabiner_grabber', got %s", cfg.ProcessName)
	}

	// Test: NewManager loads existing config
	customConfig := &Config{
		ProcessName:          "test_process",
		MemoryThresholdMB:    100,
		CheckIntervalSeconds: 30,
		IdleWaitSeconds:      5,
		LogMaxSizeMB:         20,
		LogMaxAgeDays:        14,
		Debug:                true,
	}

	data, err := json.MarshalIndent(customConfig, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	manager, err = NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed with existing config: %v", err)
	}

	cfg = manager.Get()
	if cfg.ProcessName != "test_process" {
		t.Errorf("expected process name 'test_process', got %s", cfg.ProcessName)
	}
	if cfg.MemoryThresholdMB != 100 {
		t.Errorf("expected memory threshold 100, got %d", cfg.MemoryThresholdMB)
	}
}

func TestManager_Get(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	manager, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Test: Get returns a copy
	cfg1 := manager.Get()
	cfg2 := manager.Get()

	if cfg1 == cfg2 {
		t.Error("Get should return different pointers (copies)")
	}

	// Modify one copy
	cfg1.ProcessName = "modified"

	// Check that the other copy is unchanged
	if cfg2.ProcessName == "modified" {
		t.Error("modifying one copy affected another copy")
	}

	// Check that the manager's config is unchanged
	cfg3 := manager.Get()
	if cfg3.ProcessName == "modified" {
		t.Error("modifying returned copy affected manager's config")
	}
}

func TestManager_Reload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	manager, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Test: Reload with valid new config
	newConfig := &Config{
		ProcessName:          "new_process",
		MemoryThresholdMB:    200,
		CheckIntervalSeconds: 120,
		IdleWaitSeconds:      20,
		LogMaxSizeMB:         30,
		LogMaxAgeDays:        30,
		Debug:                true,
	}

	data, err := json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	if err := manager.Reload(); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	cfg := manager.Get()
	if cfg.ProcessName != "new_process" {
		t.Errorf("expected process name 'new_process', got %s", cfg.ProcessName)
	}
	if cfg.MemoryThresholdMB != 200 {
		t.Errorf("expected memory threshold 200, got %d", cfg.MemoryThresholdMB)
	}

	// Test: Reload with invalid config (old config should be retained)
	invalidConfig := `{"process_name": "", "memory_threshold_mb": -1}`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	err = manager.Reload()
	if err == nil {
		t.Fatal("Reload should fail with invalid config")
	}

	// Old config should still be accessible
	cfg = manager.Get()
	if cfg.ProcessName != "new_process" {
		t.Errorf("old config not retained after failed reload, got process name: %s", cfg.ProcessName)
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	manager, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Test: Concurrent reads
	var wg sync.WaitGroup
	numReaders := 100

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := manager.Get()
			if cfg == nil {
				t.Error("Get returned nil")
			}
		}()
	}

	wg.Wait()

	// Test: Concurrent reads and writes
	// Use a mutex to serialize file writes to avoid corrupted JSON
	var writeMutex sync.Mutex
	numWriters := 10
	numReadersPerWriter := 10

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			newConfig := &Config{
				ProcessName:          "karabiner_grabber",
				MemoryThresholdMB:    50 + id,
				CheckIntervalSeconds: 60,
				IdleWaitSeconds:      10,
				LogMaxSizeMB:         10,
				LogMaxAgeDays:        7,
				Debug:                false,
			}

			data, err := json.MarshalIndent(newConfig, "", "  ")
			if err != nil {
				t.Errorf("failed to marshal config: %v", err)
				return
			}

			// Serialize file writes to prevent corrupted JSON
			writeMutex.Lock()
			if err := os.WriteFile(configPath, data, 0644); err != nil {
				t.Errorf("failed to write config: %v", err)
				writeMutex.Unlock()
				return
			}

			if err := manager.Reload(); err != nil {
				t.Errorf("Reload failed: %v", err)
			}
			writeMutex.Unlock()
		}(i)

		// Start readers concurrently with writers
		for j := 0; j < numReadersPerWriter; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cfg := manager.Get()
				if cfg == nil {
					t.Error("Get returned nil during concurrent access")
				}
			}()
		}
	}

	wg.Wait()
}

func TestManager_Path(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	manager, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if manager.Path() != configPath {
		t.Errorf("expected path %s, got %s", configPath, manager.Path())
	}
}
