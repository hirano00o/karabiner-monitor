#!/bin/bash

# Karabiner Monitor Uninstallation Script
# This script removes the karabiner-monitor service

set -e

echo "==== Karabiner Monitor Uninstallation ===="

PLIST_DEST="$HOME/Library/LaunchAgents/com.karabiner.monitor.plist"

# Unload LaunchAgent
if [ -f "$PLIST_DEST" ]; then
    echo "Unloading LaunchAgent..."
    launchctl unload "$PLIST_DEST" 2>/dev/null || true
    echo "✓ LaunchAgent unloaded"

    echo "Removing LaunchAgent plist..."
    rm "$PLIST_DEST"
    echo "✓ LaunchAgent plist removed"
else
    echo "LaunchAgent plist not found, skipping..."
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
echo "  Config: rm -rf ~/.config/karabiner-monitor"
echo "  Logs:   rm -rf ~/Library/Logs/karabiner-monitor"
echo ""
