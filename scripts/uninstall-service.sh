#!/bin/bash
# CASCI Service Uninstaller
# Works on Linux (systemd), macOS (launchd), and calls Windows PowerShell script

set -e

SERVICE_NAME="casci"

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        echo "windows"
    else
        echo "unknown"
    fi
}

OS=$(detect_os)

case "$OS" in
    linux)
        echo "Uninstalling CASCI systemd service..."
        
        if [ "$EUID" -ne 0 ]; then
            echo "Error: This script must be run as root"
            exit 1
        fi
        
        # Stop and disable service
        systemctl stop "$SERVICE_NAME" 2>/dev/null || true
        systemctl disable "$SERVICE_NAME" 2>/dev/null || true
        
        # Remove service file
        rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
        
        # Reload systemd
        systemctl daemon-reload
        
        # Optionally remove binary and data
        read -p "Remove binary from /usr/local/bin? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -f "/usr/local/bin/casci"
        fi
        
        read -p "Remove data directory /var/lib/casci? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "/var/lib/casci"
        fi
        
        echo "✓ Service uninstalled"
        ;;
        
    macos)
        echo "Uninstalling CASCI launchd service..."
        
        if [ "$EUID" -ne 0 ]; then
            echo "Error: This script must be run as root (use sudo)"
            exit 1
        fi
        
        # Unload service
        launchctl unload "/Library/LaunchDaemons/com.casapps.casci.plist" 2>/dev/null || true
        
        # Remove plist
        rm -f "/Library/LaunchDaemons/com.casapps.casci.plist"
        
        # Optionally remove binary and data
        read -p "Remove binary from /usr/local/bin? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -f "/usr/local/bin/casci"
        fi
        
        read -p "Remove data directory /usr/local/var/casci? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "/usr/local/var/casci"
        fi
        
        echo "✓ Service uninstalled"
        ;;
        
    windows)
        echo "For Windows, run uninstall-windows.ps1 with PowerShell as Administrator"
        ;;
        
    *)
        echo "Error: Unsupported operating system"
        exit 1
        ;;
esac
