#!/bin/bash
# CASCI launchd service installer for macOS
set -e

SERVICE_NAME="com.casapps.casci"
BINARY_NAME="casci"
SERVICE_USER="_casci"
SERVICE_GROUP="_casci"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/usr/local/var/casci"
LOG_DIR="/usr/local/var/log/casci"
PLIST_FILE="/Library/LaunchDaemons/${SERVICE_NAME}.plist"

echo "Installing CASCI launchd service for macOS..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root (use sudo)"
    exit 1
fi

# Detect binary location
if [ -f "./casci" ]; then
    BINARY_PATH="./casci"
elif [ -f "./bin/casci" ]; then
    BINARY_PATH="./bin/casci"
elif [ -f "./binaries/casci" ]; then
    BINARY_PATH="./binaries/casci"
else
    echo "Error: CASCI binary not found"
    exit 1
fi

echo "Found binary: $BINARY_PATH"

# Find available UID/GID in system range (100-499, working down from 499)
echo "Creating service user and group..."
AVAILABLE_ID=""
for id in $(seq 499 -1 100); do
    if ! dscl . -read /Groups "_casci_$id" &>/dev/null && \
       ! dscl . -read /Users "_casci_$id" &>/dev/null; then
        AVAILABLE_ID=$id
        break
    fi
done

if [ -z "$AVAILABLE_ID" ]; then
    echo "Error: No available UID/GID in range 100-499"
    exit 1
fi

# Create group
if ! dscl . -read /Groups "$SERVICE_GROUP" &>/dev/null; then
    dscl . -create /Groups "$SERVICE_GROUP"
    dscl . -create /Groups "$SERVICE_GROUP" PrimaryGroupID "$AVAILABLE_ID"
    dscl . -create /Groups "$SERVICE_GROUP" RealName "CASCI Service Group"
    dscl . -create /Groups "$SERVICE_GROUP" Password "*"
    echo "Created group: $SERVICE_GROUP (GID: $AVAILABLE_ID)"
fi

# Create user
if ! dscl . -read /Users "$SERVICE_USER" &>/dev/null; then
    dscl . -create /Users "$SERVICE_USER"
    dscl . -create /Users "$SERVICE_USER" UniqueID "$AVAILABLE_ID"
    dscl . -create /Users "$SERVICE_USER" PrimaryGroupID "$AVAILABLE_ID"
    dscl . -create /Users "$SERVICE_USER" UserShell /usr/bin/false
    dscl . -create /Users "$SERVICE_USER" NFSHomeDirectory /var/empty
    dscl . -create /Users "$SERVICE_USER" RealName "CASCI Service User"
    dscl . -create /Users "$SERVICE_USER" Password "*"
    # Hide from login window
    dscl . -create /Users "$SERVICE_USER" IsHidden 1
    defaults write /Library/Preferences/com.apple.loginwindow HiddenUsersList -array-add "$SERVICE_USER"
    echo "Created user: $SERVICE_USER (UID: $AVAILABLE_ID)"
fi

# Create directories
echo "Creating directories..."
mkdir -p "$DATA_DIR"
mkdir -p "$LOG_DIR"
mkdir -p "$DATA_DIR/artifacts"
mkdir -p "$DATA_DIR/cache"

# Install binary
echo "Installing binary to $INSTALL_DIR..."
cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
chmod 755 "$INSTALL_DIR/$BINARY_NAME"
chown root:wheel "$INSTALL_DIR/$BINARY_NAME"

# Set directory permissions
echo "Setting directory permissions..."
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$DATA_DIR"
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$LOG_DIR"
chmod 755 "$DATA_DIR"
chmod 755 "$LOG_DIR"

# Install plist file
echo "Installing launchd plist file..."
if [ -f "./scripts/com.casapps.casci.plist" ]; then
    cp "./scripts/com.casapps.casci.plist" "$PLIST_FILE"
else
    echo "Error: Plist file not found at ./scripts/com.casapps.casci.plist"
    exit 1
fi

chmod 644 "$PLIST_FILE"
chown root:wheel "$PLIST_FILE"

# Load service
echo "Loading service..."
launchctl load "$PLIST_FILE"

echo ""
echo "✓ Installation complete!"
echo ""
echo "Service commands:"
echo "  Start:   sudo launchctl start $SERVICE_NAME"
echo "  Stop:    sudo launchctl stop $SERVICE_NAME"
echo "  Restart: sudo launchctl stop $SERVICE_NAME && sudo launchctl start $SERVICE_NAME"
echo "  Status:  sudo launchctl list | grep casci"
echo "  Logs:    tail -f $LOG_DIR/stdout.log"
echo ""
echo "Configuration:"
echo "  Data directory: $DATA_DIR"
echo "  Log directory:  $LOG_DIR"
echo "  Binary:         $INSTALL_DIR/$BINARY_NAME"
echo "  Plist file:     $PLIST_FILE"
echo ""
echo "The service has been loaded and will start automatically on boot."
