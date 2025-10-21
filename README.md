# Karabiner Monitor

A memory monitoring and auto-restart tool for the karabiner_grabber process on macOS

[日本語版 README](README.ja.md)

## Overview

The `karabiner_grabber` process in Karabiner-Elements can increase memory usage after events like screen sleep, causing keyboard shortcuts to stop working. This tool monitors memory usage and automatically restarts the process when it exceeds a threshold.

## Features

- **Memory Monitoring**: Check `karabiner_grabber` process memory usage at configurable intervals
- **Auto Restart**: Automatically kill the process when memory threshold is exceeded (Karabiner restarts it automatically)
- **Wait Period**: Configurable wait time before killing the process (default 10 seconds)
- **macOS Notifications**: Uses terminal-notifier (recommended) with osascript fallback
- **Detailed Logging**: Records all monitoring and restart actions
- **Log Rotation**: Automatic log rotation with configurable size and retention
- **LaunchDaemon**: Starts automatically at system boot (runs with root privileges to monitor root processes)
- **Process Search**: Finds processes by both name and command line arguments

## Requirements

- macOS 12 or later
- Go 1.21 or later (for building)
- Karabiner-Elements
- terminal-notifier (optional, recommended for notifications)

**Note**: This tool runs as a LaunchDaemon with root privileges, so no additional permission settings are required.

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/hirano00o/karabiner-monitor.git
cd karabiner-monitor
```

### 2. Install terminal-notifier (Optional, Recommended)

If you want to receive notifications:

```bash
brew install terminal-notifier
```

terminal-notifier supports notifications from LaunchDaemons and can send notifications to users even from root processes.

### 3. Build and Install

```bash
make install
```

This command will:
- Build the binary
- Install to `/usr/local/bin`
- Create configuration file (`/Library/Application Support/karabiner-monitor/config.json`)
- Register and start the LaunchDaemon (runs with root privileges)

### 4. Verify Installation

Once installed, the service starts as a LaunchDaemon.

**Important specifications**:
- This service runs as a LaunchDaemon with root privileges
- Root privileges are required because `karabiner_grabber` runs as a root process
- No additional permission settings are required as it runs as root
- Waits for the configured number of seconds (default 10) before killing the process
- Uses `terminal-notifier` for notifications (if installed), falls back to `osascript`

Check the service operation in the logs:

```bash
sudo tail -f /var/log/karabiner-monitor/monitor.log
```

When operating normally, you should see logs like:

```
{"level":"INFO","msg":"karabiner-monitor starting","config":"/Library/Application Support/karabiner-monitor/config.json"}
{"level":"INFO","msg":"running as LaunchDaemon with root privileges"}
{"level":"INFO","msg":"starting monitoring loop","process":"karabiner_grabber","threshold_mb":50}
{"level":"INFO","msg":"process found","pid":99849,"name":"karabiner_grabber"}
{"level":"INFO","msg":"memory usage","pid":99849,"memory_mb":797.5,"threshold_mb":50}
```

## Configuration

Configuration file: `/Library/Application Support/karabiner-monitor/config.json`

```json
{
  "process_name": "karabiner_grabber",
  "memory_threshold_mb": 50,
  "check_interval_seconds": 60,
  "idle_wait_seconds": 10,
  "log_max_size_mb": 10,
  "log_max_age_days": 7
}
```

### Configuration Options

- **process_name**: Process name to monitor (default: `karabiner_grabber`)
- **memory_threshold_mb**: Memory usage threshold in MB (default: 50)
- **check_interval_seconds**: Memory check interval in seconds (default: 60)
- **idle_wait_seconds**: Wait time before killing process in seconds (default: 10)
  - Waits this many seconds before killing the process
  - Provides a grace period for users to save their work
- **log_max_size_mb**: Maximum log file size in MB (default: 10)
- **log_max_age_days**: Log file retention days (default: 7)

## Usage

### Check Service Status

```bash
sudo launchctl list | grep karabiner.monitor
```

### View Logs

```bash
# Application log
sudo tail -f /var/log/karabiner-monitor/monitor.log

# Standard output
sudo tail -f /var/log/karabiner-monitor.stdout

# Standard error
sudo tail -f /var/log/karabiner-monitor.stderr
```

### Stop Service

```bash
sudo launchctl unload /Library/LaunchDaemons/com.karabiner.monitor.plist
```

### Start Service

```bash
sudo launchctl load /Library/LaunchDaemons/com.karabiner.monitor.plist
```

### Run Locally (for debugging)

```bash
# Must run as root
sudo /usr/local/bin/karabiner-monitor
```

## Uninstallation

```bash
make uninstall
```

To completely remove configuration files and logs:

```bash
sudo rm -rf "/Library/Application Support/karabiner-monitor"
sudo rm -rf /var/log/karabiner-monitor
sudo rm -f /var/log/karabiner-monitor.stdout
sudo rm -f /var/log/karabiner-monitor.stderr
```

## Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Run Linter

```bash
make lint
```

### Format Code

```bash
make fmt
```

### Run All Checks

```bash
make check
```

## Architecture

Project structure:

```
karabiner-monitor/
├── cmd/karabiner-monitor/   # Main application
├── internal/
│   ├── config/               # Configuration management
│   ├── logger/               # Logging system
│   ├── monitor/              # Process monitoring
│   ├── keyboard/             # Wait time management
│   └── notifier/             # macOS notifications
├── scripts/                  # Install/uninstall scripts
├── configs/                  # LaunchDaemon plist
└── Makefile                  # Build and deployment automation
```

### Key Components

1. **Config**: Load and validate JSON configuration files
2. **Logger**: Log rotation using lumberjack
3. **Monitor**: Process monitoring and macOS-specific memory measurement using gopsutil/v4
   - Exact process name matching
   - Partial command line argument matching
   - macOS: Accurate memory measurement using `phys_footprint` (same as top command's MEM column)
   - Other OS: Uses RSS (Resident Set Size)
4. **Keyboard**: Wait time management
   - Provides wait time before killing processes
5. **Notifier**: macOS notifications using terminal-notifier (with osascript fallback)

### Technical Implementation Details

#### Running as Root Process

`karabiner_grabber` runs as a root process and cannot be accessed from regular user processes. Attempting to get root process information with the gopsutil library results in an "invalid argument" error.

To solve this problem, karabiner-monitor runs as a LaunchDaemon with root privileges:

1. **Process Monitoring**: Can access root process (karabiner_grabber) information by running as root
2. **Memory Measurement**: Uses vmmap command to accurately get phys_footprint
3. **Wait Time**: Waits the configured number of seconds before killing the process

#### Process Search

The FindProcess function searches in two stages:

1. **Exact Match**: Check if the process name matches exactly
2. **Command Line Search**: Check if the search string is contained in command line arguments

This allows detection of processes running with full paths like `/Library/Application Support/org.pqrs/Karabiner-Elements/bin/karabiner_grabber`.

#### Notification System

Special handling is required to send notifications from a LaunchDaemon:

1. **terminal-notifier First**: First tries `terminal-notifier` which works from LaunchDaemons
2. **osascript Fallback**: Uses `osascript` if `terminal-notifier` is not available
3. **Error Handling**: Returns error only if both fail

`terminal-notifier` can be installed with `brew install terminal-notifier` and is ideal for notifications from LaunchDaemons.

#### Memory Measurement

Different memory measurement methods are used for macOS and other operating systems:

##### macOS (darwin)

On macOS, `phys_footprint` is used to measure memory usage. This is the same value displayed in the MEM column of the top command and represents an accurate view of the process's actual memory usage.

**What is phys_footprint**:
- Physical memory footprint calculated by the macOS kernel
- More accurate actual memory usage than RSS (Resident Set Size)
- Kernel calculation formula: `(internal - alternate_accounting) + (internal_compressed - alternate_accounting_compressed) + iokit_mapped + purgeable_nonvolatile + purgeable_nonvolatile_compressed + page_table`

**Implementation**:
- Uses `vmmap --summary <pid>` command to get Physical footprint
- When running as root process, executes `vmmap` directly
- For non-root, uses `sudo -n vmmap` (requires NOPASSWD setting in sudoers)
- Parses vmmap output with regex: `Physical footprint:\s+([0-9.]+)([KMGT])?`
- Automatically converts K/M/G/T units to MB

**Comparison with RSS**:
- RSS: Size of pages resident in physical memory (includes shared memory)
- phys_footprint: Physical memory actually used by the process (considers accurate allocation of shared memory)
- Example: For `karabiner_grabber`, RSS is about 13MB but phys_footprint is about 797MB

**Limitations**:
- `vmmap` command execution can take about 2 seconds
- For root processes, the executor also requires root privileges

##### Other OS (Linux, Windows)

On non-macOS platforms, uses gopsutil library to get RSS (Resident Set Size):
- RSS: Size of memory the process holds in physical memory
- Converts from bytes to megabytes

## License

MIT License

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss the proposed changes.

## Troubleshooting

### Process Not Found

Possible causes:

1. **karabiner_grabber is not running**: Monitoring starts automatically when the process launches
   ```bash
   ps aux | grep karabiner_grabber | grep -v grep
   ```

2. **LaunchDaemon not started correctly**: Check service status
   ```bash
   sudo launchctl list | grep karabiner.monitor
   ```

3. **Service failed to start**: Check error logs
   ```bash
   sudo tail -50 /var/log/karabiner-monitor.stderr
   ```

### Service Restarting Frequently

If `/var/log/karabiner-monitor.stderr` shows repeated errors:

1. Check logs to identify the error
2. Verify configuration file: `/Library/Application Support/karabiner-monitor/config.json`
3. Stop service, fix the problem, then restart:
   ```bash
   sudo launchctl unload /Library/LaunchDaemons/com.karabiner.monitor.plist
   # Fix the issue
   sudo launchctl load /Library/LaunchDaemons/com.karabiner.monitor.plist
   ```

### Process Not Killed Despite Exceeding Memory Threshold

1. Check memory usage in logs:
   ```bash
   sudo tail -f /var/log/karabiner-monitor/monitor.log
   ```

2. Process is killed after `idle_wait_seconds` wait time (default 10 seconds)

3. When kill succeeds, you'll see logs like:
   ```
   {"level":"INFO","msg":"killing process","pid":99849,"memory_mb":797.5}
   {"level":"INFO","msg":"process killed successfully","pid":99849}
   ```

### Notifications Not Appearing

**Recommended**: Install terminal-notifier to display notifications from LaunchDaemons.

```bash
brew install terminal-notifier
```

If terminal-notifier is not installed, it falls back to osascript, but notifications from root processes may be limited.

**How to verify notifications**:

1. **Check if terminal-notifier is installed**:
   ```bash
   which terminal-notifier
   ```

2. **Check for notification errors in logs**:
   ```bash
   sudo tail -50 /var/log/karabiner-monitor.stderr | grep notification
   ```

3. **Confirm restarts in log files** (alternative to notifications):
   ```bash
   sudo tail -f /var/log/karabiner-monitor/monitor.log | grep "killed"
   ```

4. **Check karabiner_grabber PID changes**:
   ```bash
   # Record PID before restart
   ps aux | grep karabiner_grabber | grep -v grep

   # Check again after a while
   # If PID changed, process was restarted
   ```

**Note**: After installing terminal-notifier, restart the service:
```bash
sudo launchctl unload /Library/LaunchDaemons/com.karabiner.monitor.plist
sudo launchctl load /Library/LaunchDaemons/com.karabiner.monitor.plist
```

## References

- [Karabiner-Elements](https://karabiner-elements.pqrs.org/)
- [gopsutil](https://github.com/shirou/gopsutil)
- [lumberjack](https://github.com/natefinch/lumberjack)
