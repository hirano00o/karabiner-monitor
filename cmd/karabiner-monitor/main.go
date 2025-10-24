package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/hirano00o/karabiner-monitor/internal/config"
	"github.com/hirano00o/karabiner-monitor/internal/keyboard"
	"github.com/hirano00o/karabiner-monitor/internal/logger"
	"github.com/hirano00o/karabiner-monitor/internal/monitor"
	"github.com/hirano00o/karabiner-monitor/internal/notifier"
)

func main() {
	// Load configuration using ConfigManager
	// LaunchDaemon runs as root, so use system-wide config path
	configPath := "/Library/Application Support/karabiner-monitor/config.json"
	configManager, err := config.NewManager(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logPath := "/var/log/karabiner-monitor/monitor.log"
	cfg := configManager.Get()
	log := logger.New(logPath, cfg)

	log.Info("karabiner-monitor starting", "config", configPath)
	log.Info("running as LaunchDaemon with root privileges")

	// Setup file watcher for config auto-reload
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("failed to create file watcher", "error", err)
		fmt.Fprintf(os.Stderr, "failed to create file watcher: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Error("failed to close file watcher", "error", err)
		}
	}()

	// Watch the config file
	if err := watcher.Add(configPath); err != nil {
		log.Error("failed to watch config file", "error", err, "path", configPath)
		fmt.Fprintf(os.Stderr, "failed to watch config file: %v\n", err)
		os.Exit(1)
	}
	log.Info("watching config file for changes", "path", configPath)

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info("received signal, shutting down", "signal", sig)
		cancel()
	}()

	// Handle config file changes
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Filter out irrelevant events and temporary files
				if !isRelevantConfigChange(event) {
					continue
				}

				log.Info("config file changed, reloading", "event", event.Op.String())

				if err := configManager.Reload(); err != nil {
					log.Error("failed to reload config, keeping old config", "error", err)
				} else {
					newCfg := configManager.Get()
					log.Info("config reloaded successfully",
						"process", newCfg.ProcessName,
						"threshold_mb", newCfg.MemoryThresholdMB,
						"check_interval_seconds", newCfg.CheckIntervalSeconds,
						"idle_wait_seconds", newCfg.IdleWaitSeconds)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error("file watcher error", "error", err)
			}
		}
	}()

	// Main monitoring loop
	cfg = configManager.Get()
	log.Info("starting monitoring loop",
		"process", cfg.ProcessName,
		"threshold_mb", cfg.MemoryThresholdMB,
		"check_interval_seconds", cfg.CheckIntervalSeconds,
		"idle_wait_seconds", cfg.IdleWaitSeconds)

	ticker := time.NewTicker(time.Duration(cfg.CheckIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("shutdown complete")
			return

		case <-ticker.C:
			// Get latest config for each check
			currentCfg := configManager.Get()
			if err := checkAndKillIfNeeded(ctx, currentCfg, log); err != nil {
				log.Error("error during check", "error", err)
			}
		}
	}
}

// isRelevantConfigChange determines if a file system event is relevant for config reload.
// It filters out temporary files created by editors (vim, emacs, etc.) and irrelevant operations.
//
// Temporary file patterns to ignore:
//   - Vim: .swp, .swo, ~
//   - Emacs: #, ~
//   - General: .tmp, .bak
//
// Example:
//
//	event := fsnotify.Event{Name: "config.json", Op: fsnotify.Write}
//	if isRelevantConfigChange(event) {
//		// Reload config
//	}
func isRelevantConfigChange(event fsnotify.Event) bool {
	// Only care about write and create events
	if event.Op&fsnotify.Write != fsnotify.Write && event.Op&fsnotify.Create != fsnotify.Create {
		return false
	}

	// Get the base filename
	filename := filepath.Base(event.Name)

	// Ignore temporary files created by editors
	if strings.HasPrefix(filename, ".") || // Hidden files like .swp
		strings.HasSuffix(filename, "~") || // Backup files
		strings.HasPrefix(filename, "#") || // Emacs temp files
		strings.HasSuffix(filename, ".tmp") || // Temp files
		strings.HasSuffix(filename, ".bak") || // Backup files
		strings.HasSuffix(filename, ".swp") || // Vim swap files
		strings.HasSuffix(filename, ".swo") { // Vim swap files
		return false
	}

	return true
}

// checkAndKillIfNeeded checks the process memory usage and kills it if necessary.
// It implements the core business logic for monitoring and restarting processes.
//
// Process flow:
//  1. Find the target process by name
//  2. Check if it exists (skip if not found)
//  3. Get memory usage
//  4. If memory usage exceeds threshold:
//     a. Wait for the configured duration
//     b. Kill the process
//     c. Send notification
//     d. Log the action
func checkAndKillIfNeeded(ctx context.Context, cfg *config.Config, log *logger.Logger) error {
	// Find process
	proc, err := monitor.FindProcess(cfg.ProcessName)
	if err != nil {
		// Process not found, nothing to do
		log.Info("process not found", "process", cfg.ProcessName, "error", err)
		return nil
	}

	log.Info("process found", "pid", proc.PID, "name", proc.Name)

	// Get memory usage
	memoryMB, err := monitor.GetMemoryUsage(proc.PID)
	if err != nil {
		return fmt.Errorf("failed to get memory usage: %w", err)
	}

	log.Info("memory usage", "pid", proc.PID, "memory_mb", memoryMB, "threshold_mb", cfg.MemoryThresholdMB)

	// Check if memory exceeds threshold
	if memoryMB < float64(cfg.MemoryThresholdMB) {
		return nil
	}

	log.Info("memory threshold exceeded",
		"pid", proc.PID,
		"memory_mb", memoryMB,
		"threshold_mb", cfg.MemoryThresholdMB)

	// Wait for the configured duration before killing
	waitCtx, waitCancel := context.WithTimeout(ctx, time.Duration(cfg.IdleWaitSeconds+5)*time.Second)
	defer waitCancel()

	log.Info("waiting before kill", "duration_seconds", cfg.IdleWaitSeconds)
	completed := keyboard.WaitForIdle(waitCtx, time.Duration(cfg.IdleWaitSeconds)*time.Second)

	if !completed {
		log.Info("wait cancelled or timeout, skipping kill")
		return nil
	}

	// Kill the process
	log.Info("killing process", "pid", proc.PID, "name", proc.Name, "memory_mb", memoryMB)
	killTime := time.Now()

	if err := monitor.KillProcess(proc.PID); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	log.Info("process killed successfully", "pid", proc.PID, "memory_mb", memoryMB)

	// Send notification
	if err := notifier.SendNotification(memoryMB, killTime, log); err != nil {
		log.Error("failed to send notification", "error", err)
		// Don't return error, notification failure shouldn't stop the monitor
	}

	return nil
}
