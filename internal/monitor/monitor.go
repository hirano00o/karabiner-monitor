package monitor

import (
	"fmt"

	"github.com/shirou/gopsutil/v4/process"
)

// ProcessInfo contains information about a monitored process.
// It includes the process ID and name for identification and tracking.
//
// Example:
//
//	info := &ProcessInfo{
//		PID:  12345,
//		Name: "karabiner_grabber",
//	}
type ProcessInfo struct {
	// PID is the process ID
	PID int32

	// Name is the process name
	Name string
}

// FindProcess finds a process by name and returns its information.
// If multiple processes with the same name exist, it returns the first one found.
// Returns an error if the process is not found or if there's an error accessing process information.
//
// Example:
//
//	proc, err := FindProcess("karabiner_grabber")
//	if err != nil {
//		log.Printf("process not found: %v", err)
//		return
//	}
//	fmt.Printf("Found process: PID=%d, Name=%s\n", proc.PID, proc.Name)
func FindProcess(name string) (*ProcessInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to get process list: %w", err)
	}

	for _, p := range processes {
		procName, err := p.Name()
		if err != nil {
			// Skip processes we can't access
			continue
		}

		if procName == name {
			return &ProcessInfo{
				PID:  p.Pid,
				Name: procName,
			}, nil
		}
	}

	return nil, fmt.Errorf("process %s not found", name)
}

// GetMemoryUsage returns the memory usage of a process in megabytes.
// It retrieves the Resident Set Size (RSS) which represents the portion of memory
// occupied by the process that is held in main memory (RAM).
//
// Returns an error if the process doesn't exist or memory information is unavailable.
//
// Example:
//
//	memoryMB, err := GetMemoryUsage(12345)
//	if err != nil {
//		log.Fatalf("failed to get memory usage: %v", err)
//	}
//	fmt.Printf("Process is using %.2f MB\n", memoryMB)
func GetMemoryUsage(pid int32) (float64, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return 0, fmt.Errorf("failed to get process: %w", err)
	}

	memInfo, err := p.MemoryInfo()
	if err != nil {
		return 0, fmt.Errorf("failed to get memory info: %w", err)
	}

	// Convert RSS from bytes to megabytes
	memoryMB := float64(memInfo.RSS) / 1024 / 1024

	return memoryMB, nil
}

// KillProcess terminates a process by its PID.
// It sends a SIGTERM signal to gracefully terminate the process.
// The process may handle the signal and clean up before exiting.
//
// Returns an error if the process doesn't exist or cannot be terminated.
//
// Example:
//
//	if err := KillProcess(12345); err != nil {
//		log.Fatalf("failed to kill process: %v", err)
//	}
//	log.Println("Process terminated successfully")
func KillProcess(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to get process: %w", err)
	}

	if err := p.Terminate(); err != nil {
		return fmt.Errorf("failed to terminate process: %w", err)
	}

	return nil
}
