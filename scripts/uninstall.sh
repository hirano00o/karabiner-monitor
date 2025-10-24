#!/bin/bash

# Karabiner Monitor Uninstallation Script
# This script removes the karabiner-monitor service (LaunchDaemon)

set -e

echo "==== Karabiner Monitor Uninstallation ===="

PLIST_DEST="/Library/LaunchDaemons/com.karabiner.monitor.plist"

# Unload LaunchDaemon
if [ -f "$PLIST_DEST" ]; then
    echo "Unloading LaunchDaemon..."
    sudo launchctl unload "$PLIST_DEST" 2>/dev/null || true
    echo "✓ LaunchDaemon unloaded"

    echo "Removing LaunchDaemon plist..."
    sudo rm "$PLIST_DEST"
    echo "✓ LaunchDaemon plist removed"
else
    echo "LaunchDaemon plist not found, skipping..."
fi

# Remove binary
if [ -f "/usr/local/bin/karabiner-monitor" ]; then
    echo "Removing binary..."
    sudo rm /usr/local/bin/karabiner-monitor
    echo "✓ Binary removed"
else
    echo "Binary not found, skipping..."
fi

echo ""
echo "==== Uninstallation Complete ===="
echo ""
echo "Configuration and logs have been preserved."
echo "To remove them manually:"
echo "  Config: sudo rm -rf '/Library/Application Support/karabiner-monitor'"
echo "  Logs:   sudo rm -rf /var/log/karabiner-monitor"
echo "  Stdout: sudo rm -f /var/log/karabiner-monitor.stdout"
echo "  Stderr: sudo rm -f /var/log/karabiner-monitor.stderr"
echo ""
