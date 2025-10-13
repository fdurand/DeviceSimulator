#!/bin/bash

# Debian Package Validator for DeviceSimulator

PACKAGE_NAME="device-simulator_1.0.0_amd64.deb"

if [ ! -f "$PACKAGE_NAME" ]; then
    echo "❌ Package file not found: $PACKAGE_NAME"
    exit 1
fi

echo "=== DeviceSimulator Debian Package Validation ==="
echo

# Check package info
echo "📦 Package Information:"
dpkg-deb --info "$PACKAGE_NAME"
echo

# Check package contents
echo "📁 Package Contents:"
dpkg-deb --contents "$PACKAGE_NAME" | head -20
echo "... (showing first 20 files)"
echo

# Validate package structure
echo "✅ Package Structure Validation:"

# Check if key files exist in package
REQUIRED_FILES=(
    "./usr/bin/device-simulator"
    "./etc/device-simulator/config.ini"
    "./etc/device-simulator/config-xerox-printer.ini"
    "./lib/systemd/system/device-simulator@.service"
    "./usr/share/doc/device-simulator/README.md"
    "./usr/share/man/man1/device-simulator.1.gz"
    "./DEBIAN/control"
    "./DEBIAN/postinst"
)

for file in "${REQUIRED_FILES[@]}"; do
    if dpkg-deb --contents "$PACKAGE_NAME" | grep -q "$file"; then
        echo "✅ $file"
    else
        echo "❌ Missing: $file"
    fi
done

echo
echo "📊 Package Statistics:"
echo "Size: $(ls -lh $PACKAGE_NAME | awk '{print $5}')"
echo "Files: $(dpkg-deb --contents $PACKAGE_NAME | wc -l)"

echo
echo "🔍 Control Scripts:"
dpkg-deb --control "$PACKAGE_NAME" temp-control
ls -la temp-control/
rm -rf temp-control/

echo
echo "✅ Package validation completed!"
echo "Install with: sudo dpkg -i $PACKAGE_NAME"