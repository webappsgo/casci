#!/bin/bash
# CASCI systemd service installer for Linux
set -e

SERVICE_NAME="casci"
BINARY_NAME="casci"
SERVICE_USER="casci"
SERVICE_GROUP="casci"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/var/lib/casci"
LOG_DIR="/var/log/casci"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

echo "Installing CASCI systemd service..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
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

# Create service user and group
echo "Creating service user and group..."
if ! getent group "$SERVICE_GROUP" > /dev/null 2>&1; then
    groupadd --system "$SERVICE_GROUP"
    echo "Created group: $SERVICE_GROUP"
fi

if ! getent passwd "$SERVICE_USER" > /dev/null 2>&1; then
    useradd --system --gid "$SERVICE_GROUP" --shell /sbin/nologin \
        --comment "CASCI Service User" --home-dir "$DATA_DIR" "$SERVICE_USER"
    echo "Created user: $SERVICE_USER"
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
chown root:root "$INSTALL_DIR/$BINARY_NAME"

# Set directory permissions
echo "Setting directory permissions..."
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$DATA_DIR"
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$LOG_DIR"
chmod 755 "$DATA_DIR"
chmod 755 "$LOG_DIR"

# Install service file
echo "Installing systemd service file..."
if [ -f "./scripts/casci.service" ]; then
    cp "./scripts/casci.service" "$SERVICE_FILE"
else
    echo "Error: Service file not found at ./scripts/casci.service"
    exit 1
fi

chmod 644 "$SERVICE_FILE"

# Reload systemd
echo "Reloading systemd..."
systemctl daemon-reload

# Enable service
echo "Enabling service..."
systemctl enable "$SERVICE_NAME"

echo ""
echo "✓ Installation complete!"
echo ""
echo "Service commands:"
echo "  Start:   sudo systemctl start $SERVICE_NAME"
echo "  Stop:    sudo systemctl stop $SERVICE_NAME"
echo "  Restart: sudo systemctl restart $SERVICE_NAME"
echo "  Status:  sudo systemctl status $SERVICE_NAME"
echo "  Logs:    sudo journalctl -u $SERVICE_NAME -f"
echo ""
echo "Configuration:"
echo "  Data directory: $DATA_DIR"
echo "  Log directory:  $LOG_DIR"
echo "  Binary:         $INSTALL_DIR/$BINARY_NAME"
echo "  Service file:   $SERVICE_FILE"
echo ""
echo "To start the service now, run:"
echo "  sudo systemctl start $SERVICE_NAME"
