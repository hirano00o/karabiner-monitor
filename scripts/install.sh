#!/bin/bash

# Karabiner Monitor Installation Script
# This script installs the karabiner-monitor service as a LaunchDaemon (runs as root)

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

# Create config directory (owned by root since daemon runs as root)
CONFIG_DIR="/Library/Application Support/karabiner-monitor"
echo "Creating config directory: $CONFIG_DIR"
sudo mkdir -p "$CONFIG_DIR"
echo "✓ Config directory created"

# Create default config if it doesn't exist
CONFIG_FILE="$CONFIG_DIR/config.json"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Creating default configuration..."
    sudo tee "$CONFIG_FILE" > /dev/null << 'EOF'
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
LOG_DIR="/var/log/karabiner-monitor"
echo "Creating log directory: $LOG_DIR"
sudo mkdir -p "$LOG_DIR"
echo "✓ Log directory created"

# Install LaunchDaemon
PLIST_SRC="./configs/com.karabiner.monitor.daemon.plist"
PLIST_DEST="/Library/LaunchDaemons/com.karabiner.monitor.plist"

if [ ! -f "$PLIST_SRC" ]; then
    echo "Error: LaunchDaemon plist not found at $PLIST_SRC"
    exit 1
fi

echo "Installing LaunchDaemon..."
sudo cp "$PLIST_SRC" "$PLIST_DEST"
sudo chown root:wheel "$PLIST_DEST"
sudo chmod 644 "$PLIST_DEST"
echo "✓ LaunchDaemon plist installed"

# Load LaunchDaemon
echo "Loading LaunchDaemon..."
sudo launchctl unload "$PLIST_DEST" 2>/dev/null || true
sudo launchctl load "$PLIST_DEST"
echo "✓ LaunchDaemon loaded"

echo ""
echo "==== Installation Complete ===="
echo ""
echo "📝 Service Information"
echo ""
echo "This service runs as a LaunchDaemon with root privileges, which allows it to:"
echo "  • Monitor the karabiner_grabber process (which also runs as root)"
echo "  • Access accurate memory usage via phys_footprint"
echo "  • Kill and restart the process when memory threshold is exceeded"
echo ""
echo "No additional permissions are required."
echo ""
echo "Configuration file: $CONFIG_FILE"
echo "Log directory: $LOG_DIR"
echo "Standard output: /var/log/karabiner-monitor.stdout"
echo "Standard error: /var/log/karabiner-monitor.stderr"
echo ""
echo "To check status: sudo launchctl list | grep karabiner.monitor"
echo "To view logs: sudo tail -f $LOG_DIR/monitor.log"
echo ""
