package config

import (
	"sync"
)

// Manager manages configuration with thread-safe access.
// It provides concurrent-safe read and reload operations for the application configuration.
//
// Example:
//
//	manager := NewManager("/path/to/config.json")
//	cfg := manager.Get()
//	fmt.Printf("Process: %s\n", cfg.ProcessName)
//
//	// Reload configuration
//	if err := manager.Reload(); err != nil {
//		log.Printf("reload failed: %v", err)
//	}
type Manager struct {
	mu   sync.RWMutex
	cfg  *Config
	path string
}

// NewManager creates a new configuration manager and loads the initial configuration.
// If the configuration file doesn't exist, it creates a default one.
// Returns an error if the initial load fails.
//
// Example:
//
//	manager, err := NewManager("/Library/Application Support/karabiner-monitor/config.json")
//	if err != nil {
//		log.Fatalf("failed to create manager: %v", err)
//	}
func NewManager(path string) (*Manager, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}

	return &Manager{
		cfg:  cfg,
		path: path,
	}, nil
}

// Get returns a copy of the current configuration.
// This method is safe for concurrent use and provides read-only access to the configuration.
//
// Note: The returned Config is a copy, so modifications to it will not affect the manager's state.
//
// Example:
//
//	cfg := manager.Get()
//	fmt.Printf("Memory threshold: %d MB\n", cfg.MemoryThresholdMB)
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modifications
	cfgCopy := *m.cfg
	return &cfgCopy
}

// Reload reloads the configuration from the file.
// If the new configuration is invalid, the old configuration is retained and an error is returned.
// This method is safe for concurrent use.
//
// The reload process:
//  1. Load and parse the configuration file
//  2. Validate the new configuration
//  3. If valid, atomically replace the old configuration
//  4. If invalid, keep the old configuration and return error
//
// Example:
//
//	if err := manager.Reload(); err != nil {
//		log.Printf("config reload failed, keeping old config: %v", err)
//	} else {
//		log.Println("config reloaded successfully")
//	}
func (m *Manager) Reload() error {
	newCfg, err := Load(m.path)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.cfg = newCfg
	m.mu.Unlock()

	return nil
}

// Path returns the path to the configuration file.
//
// Example:
//
//	path := manager.Path()
//	fmt.Printf("Config file: %s\n", path)
func (m *Manager) Path() string {
	return m.path
}
