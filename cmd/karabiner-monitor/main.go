package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hirano00o/karabiner-monitor/internal/config"
	"github.com/hirano00o/karabiner-monitor/internal/keyboard"
	"github.com/hirano00o/karabiner-monitor/internal/logger"
	"github.com/hirano00o/karabiner-monitor/internal/monitor"
	"github.com/hirano00o/karabiner-monitor/internal/notifier"
)

func main() {
	// Load configuration
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(homeDir, ".config", "karabiner-monitor", "config.json")
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logPath := filepath.Join(homeDir, "Library", "Logs", "karabiner-monitor", "monitor.log")
	log := logger.New(logPath, cfg)

	log.Info("karabiner-monitor starting", "config", configPath)

	// Check accessibility permission
	if !keyboard.CheckAccessibilityPermission() {
		log.Error("accessibility permission not granted")
		fmt.Fprintln(os.Stderr, "ERROR: Accessibility permission required.")
		fmt.Fprintln(os.Stderr, "Please grant permission in: System Preferences > Security & Privacy > Privacy > Accessibility")
		os.Exit(1)
	}

	log.Info("accessibility permission granted")

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

	// Main monitoring loop
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
			if err := checkAndKillIfNeeded(ctx, cfg, log); err != nil {
				log.Error("error during check", "error", err)
			}
		}
	}
}

// checkAndKillIfNeeded checks the process memory usage and kills it if necessary.
// It implements the core business logic for monitoring and restarting processes.
//
// Process flow:
//  1. Find the target process by name
//  2. Check if it exists (skip if not found)
//  3. Get memory usage
//  4. If memory usage exceeds threshold:
//     a. Wait for keyboard idle period
//     b. Kill the process if still idle
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

	// Wait for keyboard idle
	idleCtx, idleCancel := context.WithTimeout(ctx, time.Duration(cfg.IdleWaitSeconds+5)*time.Second)
	defer idleCancel()

	log.Info("waiting for keyboard idle", "duration_seconds", cfg.IdleWaitSeconds)
	wasIdle := keyboard.WaitForIdle(idleCtx, time.Duration(cfg.IdleWaitSeconds)*time.Second)

	if !wasIdle {
		log.Info("keyboard activity detected or timeout, skipping kill")
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
	if err := notifier.SendNotification(memoryMB, killTime); err != nil {
		log.Error("failed to send notification", "error", err)
		// Don't return error, notification failure shouldn't stop the monitor
	}

	return nil
}
