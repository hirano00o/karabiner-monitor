//go:build darwin

package monitor

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// GetMemoryUsageDarwin returns the memory usage of a process on macOS using phys_footprint.
// This function uses the vmmap command to retrieve the phys_footprint,
// which corresponds to the MEM column in the top command.
//
// The phys_footprint is calculated by the kernel as:
//   - internal + internal_compressed + iokit_mapped + purgeable + page_table
//   - minus alternate_accounting (for memory charged to other processes)
//
// This provides a more accurate representation of a process's memory impact on macOS
// compared to RSS (Resident Set Size).
//
// Note: This implementation uses vmmap command instead of task_for_pid() API
// because task_for_pid() requires special entitlements on modern macOS,
// even when running as root. The vmmap approach is slower but more reliable.
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
//	fmt.Printf("Process is using %.2f MB (phys_footprint)\n", memoryMB)
func GetMemoryUsageDarwin(pid int32) (float64, error) {
	// Use vmmap to get physical footprint
	// vmmap --summary <pid> outputs a summary including "Physical footprint: XXX"
	// If we're running as root (UID 0), we can use vmmap directly
	// Otherwise, we need to use sudo for root processes
	var cmd *exec.Cmd
	if os.Geteuid() == 0 {
		// Running as root, use vmmap directly
		cmd = exec.Command("vmmap", "--summary", fmt.Sprintf("%d", pid))
	} else {
		// Not running as root, use sudo
		cmd = exec.Command("sudo", "-n", "vmmap", "--summary", fmt.Sprintf("%d", pid))
	}

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run vmmap for pid %d: %w", pid, err)
	}

	// Parse the output to find "Physical footprint: XXX"
	// Example line: "Physical footprint:         797.0M"
	re := regexp.MustCompile(`Physical footprint:\s+([0-9.]+)([KMGT])?`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return 0, fmt.Errorf("failed to parse phys_footprint from vmmap output")
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse footprint value: %w", err)
	}

	// Convert to MB based on the unit
	unit := "K" // default to kilobytes if no unit specified
	if len(matches) > 2 && matches[2] != "" {
		unit = matches[2]
	}

	var memoryMB float64
	switch strings.ToUpper(unit) {
	case "K":
		memoryMB = value / 1024
	case "M":
		memoryMB = value
	case "G":
		memoryMB = value * 1024
	case "T":
		memoryMB = value * 1024 * 1024
	default:
		// Assume bytes if no unit
		memoryMB = value / 1024 / 1024
	}

	return memoryMB, nil
}
