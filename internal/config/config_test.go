package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad tests loading configuration from a JSON file
func TestLoad(t *testing.T) {
	t.Run("load existing config", func(t *testing.T) {
		// Create temporary config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		configJSON := `{
			"process_name": "test_process",
			"memory_threshold_mb": 100,
			"check_interval_seconds": 30,
			"idle_wait_seconds": 5,
			"log_max_size_mb": 20,
			"log_max_age_days": 14
		}`

		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.ProcessName != "test_process" {
			t.Errorf("ProcessName = %v, want test_process", cfg.ProcessName)
		}
		if cfg.MemoryThresholdMB != 100 {
			t.Errorf("MemoryThresholdMB = %v, want 100", cfg.MemoryThresholdMB)
		}
		if cfg.CheckIntervalSeconds != 30 {
			t.Errorf("CheckIntervalSeconds = %v, want 30", cfg.CheckIntervalSeconds)
		}
		if cfg.IdleWaitSeconds != 5 {
			t.Errorf("IdleWaitSeconds = %v, want 5", cfg.IdleWaitSeconds)
		}
		if cfg.LogMaxSizeMB != 20 {
			t.Errorf("LogMaxSizeMB = %v, want 20", cfg.LogMaxSizeMB)
		}
		if cfg.LogMaxAgeDays != 14 {
			t.Errorf("LogMaxAgeDays = %v, want 14", cfg.LogMaxAgeDays)
		}
	})

	t.Run("create default config if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Check default values
		if cfg.ProcessName != "karabiner_grabber" {
			t.Errorf("ProcessName = %v, want karabiner_grabber", cfg.ProcessName)
		}
		if cfg.MemoryThresholdMB != 50 {
			t.Errorf("MemoryThresholdMB = %v, want 50", cfg.MemoryThresholdMB)
		}
		if cfg.CheckIntervalSeconds != 60 {
			t.Errorf("CheckIntervalSeconds = %v, want 60", cfg.CheckIntervalSeconds)
		}
		if cfg.IdleWaitSeconds != 10 {
			t.Errorf("IdleWaitSeconds = %v, want 10", cfg.IdleWaitSeconds)
		}
		if cfg.LogMaxSizeMB != 10 {
			t.Errorf("LogMaxSizeMB = %v, want 10", cfg.LogMaxSizeMB)
		}
		if cfg.LogMaxAgeDays != 7 {
			t.Errorf("LogMaxAgeDays = %v, want 7", cfg.LogMaxAgeDays)
		}

		// Verify file was created
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file should be created")
		}
	})
}

// TestValidate tests configuration validation
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				ProcessName:          "test",
				MemoryThresholdMB:    50,
				CheckIntervalSeconds: 60,
				IdleWaitSeconds:      10,
				LogMaxSizeMB:         10,
				LogMaxAgeDays:        7,
			},
			wantErr: false,
		},
		{
			name: "invalid memory threshold",
			config: Config{
				ProcessName:          "test",
				MemoryThresholdMB:    0,
				CheckIntervalSeconds: 60,
				IdleWaitSeconds:      10,
				LogMaxSizeMB:         10,
				LogMaxAgeDays:        7,
			},
			wantErr: true,
		},
		{
			name: "invalid check interval",
			config: Config{
				ProcessName:          "test",
				MemoryThresholdMB:    50,
				CheckIntervalSeconds: 0,
				IdleWaitSeconds:      10,
				LogMaxSizeMB:         10,
				LogMaxAgeDays:        7,
			},
			wantErr: true,
		},
		{
			name: "empty process name",
			config: Config{
				ProcessName:          "",
				MemoryThresholdMB:    50,
				CheckIntervalSeconds: 60,
				IdleWaitSeconds:      10,
				LogMaxSizeMB:         10,
				LogMaxAgeDays:        7,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
