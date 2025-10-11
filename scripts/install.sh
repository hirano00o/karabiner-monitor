#!/bin/bash

# Karabiner Monitor Installation Script
# This script installs the karabiner-monitor service as a LaunchAgent

set -e

echo "==== Karabiner Monitor Installation ===="

# Check if binary exists
if [ ! -f "./karabiner-monitor" ]; then
    echo "Error: karabiner-monitor binary not found."
    echo "Please run 'make build' first."
    exit 1
fi

# Install binary to /usr/local/bin
echo "Installing binary to /usr/local/bin..."
sudo cp ./karabiner-monitor /usr/local/bin/karabiner-monitor
sudo chmod +x /usr/local/bin/karabiner-monitor
echo "✓ Binary installed"

# Create config directory
CONFIG_DIR="$HOME/.config/karabiner-monitor"
echo "Creating config directory: $CONFIG_DIR"
mkdir -p "$CONFIG_DIR"
echo "✓ Config directory created"

# Create default config if it doesn't exist
CONFIG_FILE="$CONFIG_DIR/config.json"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Creating default configuration..."
    cat > "$CONFIG_FILE" << 'EOF'
{
  "process_name": "karabiner_grabber",
  "memory_threshold_mb": 50,
  "check_interval_seconds": 60,
  "idle_wait_seconds": 10,
  "log_max_size_mb": 10,
  "log_max_age_days": 7
}
EOF
    echo "✓ Default configuration created at: $CONFIG_FILE"
else
    echo "✓ Configuration already exists at: $CONFIG_FILE"
fi

# Create log directory
LOG_DIR="$HOME/Library/Logs/karabiner-monitor"
echo "Creating log directory: $LOG_DIR"
mkdir -p "$LOG_DIR"
echo "✓ Log directory created"

# Install LaunchAgent
PLIST_SRC="./configs/com.karabiner.monitor.plist"
PLIST_DEST="$HOME/Library/LaunchAgents/com.karabiner.monitor.plist"

if [ ! -f "$PLIST_SRC" ]; then
    echo "Error: LaunchAgent plist not found at $PLIST_SRC"
    exit 1
fi

echo "Installing LaunchAgent..."
cp "$PLIST_SRC" "$PLIST_DEST"
echo "✓ LaunchAgent plist installed"

# Load LaunchAgent
echo "Loading LaunchAgent..."
launchctl unload "$PLIST_DEST" 2>/dev/null || true
launchctl load "$PLIST_DEST"
echo "✓ LaunchAgent loaded"

echo ""
echo "==== Installation Complete ===="
echo ""
echo "⚠️  IMPORTANT: Accessibility Permission Required"
echo ""
echo "Please grant accessibility permission to karabiner-monitor:"
echo "1. Open System Preferences > Security & Privacy > Privacy"
echo "2. Select 'Accessibility' from the left sidebar"
echo "3. Click the lock to make changes"
echo "4. Add '/usr/local/bin/karabiner-monitor' or your Terminal app"
echo "5. Check the box to enable it"
echo ""
echo "Configuration file: $CONFIG_FILE"
echo "Log file: $LOG_DIR/monitor.log"
echo ""
echo "To check status: launchctl list | grep karabiner.monitor"
echo "To view logs: tail -f $LOG_DIR/monitor.log"
echo ""
