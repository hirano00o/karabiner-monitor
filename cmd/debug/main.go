package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

func main() {
	processes, err := process.Processes()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Looking for karabiner_grabber (PID 99849)...")
	fmt.Println()

	// Try to directly access PID 99849
	p, err := process.NewProcess(99849)
	if err != nil {
		fmt.Printf("ERROR: Cannot create process object for PID 99849: %v\n", err)
	} else {
		procName, err := p.Name()
		if err != nil {
			fmt.Printf("ERROR: Cannot get name for PID 99849: %v\n", err)
		} else {
			fmt.Printf("Direct access to PID 99849:\n")
			fmt.Printf("  Name: %s\n", procName)
		}

		cmdline, err := p.Cmdline()
		if err != nil {
			fmt.Printf("ERROR: Cannot get cmdline for PID 99849: %v\n", err)
		} else {
			fmt.Printf("  Cmdline: %s\n", cmdline)
		}
	}
	fmt.Println()

	fmt.Println("Scanning all processes for karabiner...")
	fmt.Println()

	for _, p := range processes {
		procName, err := p.Name()
		if err != nil {
			continue
		}

		// Check if it's karabiner related
		if strings.Contains(procName, "karabiner") || strings.Contains(strings.ToLower(procName), "karabiner") {
			cmdline, _ := p.Cmdline()
			fmt.Printf("Found karabiner process:\n")
			fmt.Printf("  PID:     %d\n", p.Pid)
			fmt.Printf("  Name:    %s\n", procName)
			fmt.Printf("  Cmdline: %s\n", cmdline)
			fmt.Println()
		}

		// Also check cmdline for exact match
		cmdline, err := p.Cmdline()
		if err != nil {
			continue
		}

		if strings.Contains(cmdline, "karabiner_grabber") {
			fmt.Printf("Found by cmdline match:\n")
			fmt.Printf("  PID:     %d\n", p.Pid)
			fmt.Printf("  Name:    %s\n", procName)
			fmt.Printf("  Cmdline: %s\n", cmdline)
			fmt.Println()
		}
	}
}
