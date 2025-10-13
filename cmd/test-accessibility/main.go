package main

import (
	"fmt"
	"os"

	"github.com/hirano00o/karabiner-monitor/internal/keyboard"
)

func main() {
	fmt.Printf("Running as UID: %d\n", os.Getuid())
	hasPermission := keyboard.CheckAccessibilityPermission()
	fmt.Printf("Accessibility permission: %v\n", hasPermission)
}
