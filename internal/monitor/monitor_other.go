//go:build !darwin

package monitor

import (
	"fmt"

	"github.com/shirou/gopsutil/v4/process"
)

// GetMemoryUsageDarwin is a placeholder for non-Darwin platforms.
// On non-macOS systems, we use RSS (Resident Set Size) instead.
//
// This function retrieves the Resident Set Size (RSS) which represents
// the portion of memory occupied by the process that is held in main memory (RAM).
//
// Returns memory usage in megabytes, or an error if the process doesn't exist
// or memory information is unavailable.
//
// Example:
//
//	memoryMB, err := GetMemoryUsageDarwin(12345)
//	if err != nil {
//		log.Fatalf("failed to get memory usage: %v", err)
//	}
//	fmt.Printf("Process is using %.2f MB (RSS)\n", memoryMB)
func GetMemoryUsageDarwin(pid int32) (float64, error) {
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
