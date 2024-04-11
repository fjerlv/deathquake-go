#!/bin/bash

# Quake 3 Arena Dedicated Server Launcher
# ========================================
#
# This script starts a Quake 3 dedicated server and logs all game events
# with timestamps to a file for analysis with deathquake-go.
#
# Usage:
#   ./start_server.sh <ioquake_directory>
#
# Arguments:
#   ioquake_directory - Path to ioquake3 installation (required)
#
# Examples:
#   ./start_server.sh /home/fjerlv/ioquake5
#   ./start_server.sh /opt/ioquake3
#   ./start_server.sh ~/games/ioquake3
#
# What it does:
#   1. Copies the server configuration to the ioquake3 baseq3 directory
#   2. Sets up the RCON password for remote server administration
#   3. Starts the dedicated server (ioq3ded.x86_64)
#   4. Adds timestamps to each log line using gawk
#   5. Saves the timestamped output to a dated log file in the current directory
#
# Output:
#   - Log file: game_YYYYMMDD_HHMMSS.log
#   - The log is also displayed in the terminal (via tee)
#
# Requirements:
#   - ioquake3 server binary (ioq3ded.x86_64) in the specified directory
#   - server.cfg in the current directory
#   - gawk for timestamp formatting

# Check if ioquake3 directory argument is provided
if [ -z "$1" ]; then
    echo "Error: ioquake3 directory is required"
    echo "Usage: $0 <ioquake_directory>"
    echo "Example: $0 /home/fjerlv/ioquake5"
    exit 1
fi

IOQUAKE_DIR="$1"

# Validate that the directory exists
if [ ! -d "$IOQUAKE_DIR" ]; then
    echo "Error: ioquake3 directory not found: $IOQUAKE_DIR"
    exit 1
fi

# Validate that the server binary exists
if [ ! -f "$IOQUAKE_DIR/ioq3ded.x86_64" ]; then
    echo "Error: Server binary not found: $IOQUAKE_DIR/ioq3ded.x86_64"
    exit 1
fi

# Validate that server.cfg exists in current directory
if [ ! -f "server.cfg" ]; then
    echo "Error: server.cfg not found in current directory"
    exit 1
fi

echo "Using ioquake3 directory: $IOQUAKE_DIR"

# Copy server configuration to ioquake3 directory
cp server.cfg "$IOQUAKE_DIR/baseq3/server.cfg"

# Create RCON configuration file with password
echo seta rconpassword \"Hunter2\" > "$IOQUAKE_DIR/baseq3/rcon.cfg"

# Start the dedicated server and pipe output through:
# - gawk: adds timestamps to each line
# - tee: saves to log file while displaying in terminal
"$IOQUAKE_DIR/ioq3ded.x86_64" +exec server.cfg 2>&1 | gawk '{ print strftime("%Y-%m-%d %H:%M:%S", systime()), $0; fflush() }' | tee -a /home/fjerlv/GolandProjects/deathquake-go/game_$(date +%Y%m%d_%H%M%S).log

