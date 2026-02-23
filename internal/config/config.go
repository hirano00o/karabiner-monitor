package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration.
// It manages settings for process monitoring, memory thresholds, and logging behavior.
//
// Example:
//
//	cfg := &Config{
//		ProcessName:          "Karabiner-Core-Service",
//		MemoryThresholdMB:    50,
//		CheckIntervalSeconds: 60,
//		IdleWaitSeconds:      10,
//		LogMaxSizeMB:         10,
//		LogMaxAgeDays:        7,
//	}
type Config struct {
	// ProcessName is the name of the process to monitor
	ProcessName string `json:"process_name"`

	// MemoryThresholdMB is the memory usage threshold in megabytes
	// When exceeded, the process will be killed
	MemoryThresholdMB int `json:"memory_threshold_mb"`

	// CheckIntervalSeconds is the interval in seconds to check memory usage
	CheckIntervalSeconds int `json:"check_interval_seconds"`

	// IdleWaitSeconds is the duration in seconds to wait without keyboard input
	// before killing the process
	IdleWaitSeconds int `json:"idle_wait_seconds"`

	// LogMaxSizeMB is the maximum size of a log file in megabytes before rotation
	LogMaxSizeMB int `json:"log_max_size_mb"`

	// LogMaxAgeDays is the maximum number of days to retain old log files
	LogMaxAgeDays int `json:"log_max_age_days"`

	// Debug enables debug logging when set to true
	Debug bool `json:"debug"`
}

// Default returns a Config with default values.
//
// Default values are:
//   - ProcessName: "Karabiner-Core-Service"
//   - MemoryThresholdMB: 50
//   - CheckIntervalSeconds: 60
//   - IdleWaitSeconds: 10
//   - LogMaxSizeMB: 10
//   - LogMaxAgeDays: 7
//   - Debug: false
//
// Example:
//
//	cfg := Default()
//	fmt.Printf("Default process: %s\n", cfg.ProcessName)
//	// Output: Default process: Karabiner-Core-Service
func Default() *Config {
	return &Config{
		ProcessName:          "Karabiner-Core-Service",
		MemoryThresholdMB:    50,
		CheckIntervalSeconds: 60,
		IdleWaitSeconds:      10,
		LogMaxSizeMB:         10,
		LogMaxAgeDays:        7,
		Debug:                false,
	}
}

// Validate validates the configuration values.
// It returns an error if any value is invalid.
//
// Validation rules:
//   - ProcessName must not be empty
//   - MemoryThresholdMB must be greater than 0
//   - CheckIntervalSeconds must be greater than 0
//   - IdleWaitSeconds must be greater than or equal to 0
//   - LogMaxSizeMB must be greater than 0
//   - LogMaxAgeDays must be greater than 0
//
// Example:
//
//	cfg := &Config{ProcessName: "test", MemoryThresholdMB: 50}
//	if err := cfg.Validate(); err != nil {
//		log.Fatalf("invalid config: %v", err)
//	}
func (c *Config) Validate() error {
	if c.ProcessName == "" {
		return errors.New("process_name must not be empty")
	}
	if c.MemoryThresholdMB <= 0 {
		return errors.New("memory_threshold_mb must be greater than 0")
	}
	if c.CheckIntervalSeconds <= 0 {
		return errors.New("check_interval_seconds must be greater than 0")
	}
	if c.IdleWaitSeconds < 0 {
		return errors.New("idle_wait_seconds must be greater than or equal to 0")
	}
	if c.LogMaxSizeMB <= 0 {
		return errors.New("log_max_size_mb must be greater than 0")
	}
	if c.LogMaxAgeDays <= 0 {
		return errors.New("log_max_age_days must be greater than 0")
	}
	return nil
}

// Load loads configuration from the specified path.
// If the file does not exist, it creates a default configuration file and returns it.
// If the file exists but contains invalid JSON, it returns an error.
//
// The directory path is automatically created if it doesn't exist.
//
// Example:
//
//	cfg, err := Load("~/.config/karabiner-monitor/config.json")
//	if err != nil {
//		log.Fatalf("failed to load config: %v", err)
//	}
//	fmt.Printf("Monitoring: %s\n", cfg.ProcessName)
func Load(path string) (*Config, error) {
	// Create directory if not exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// If config file doesn't exist, create default
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := Default()
		if err := save(path, cfg); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return cfg, nil
	}

	// Read existing config
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// save saves the configuration to the specified path.
// This is an internal function used by Load to create default configuration files.
//
// The JSON output is formatted with indentation for readability.
func save(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
