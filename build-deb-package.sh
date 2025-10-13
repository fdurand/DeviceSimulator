#!/bin/bash

# Simple Debian Package Builder for DeviceSimulator
# This script creates a basic .deb package without complex build dependencies

set -e

PACKAGE_NAME="device-simulator"
VERSION="1.0.0"
ARCH="amd64"
MAINTAINER="Fabrice Durand <fdurand@inverse.ca>"

# Create package directory structure
PKG_DIR="debian-package/${PACKAGE_NAME}_${VERSION}_${ARCH}"
mkdir -p "$PKG_DIR"/{DEBIAN,usr/bin,etc/device-simulator,lib/systemd/system,usr/share/doc/device-simulator,usr/share/man/man1,usr/share/device-simulator}

echo "Building DeviceSimulator binary..."
make build

echo "Creating package structure..."

# Copy binary
cp bin/device-simulator "$PKG_DIR/usr/bin/"
chmod 755 "$PKG_DIR/usr/bin/device-simulator"

# Copy configuration files
cp config.ini "$PKG_DIR/etc/device-simulator/"
cp config-xerox-printer.ini "$PKG_DIR/etc/device-simulator/"
cp xerox-dhcp-options.json "$PKG_DIR/etc/device-simulator/"

# Copy systemd service
cp debian/device-simulator@.service "$PKG_DIR/lib/systemd/system/"

# Copy documentation
cp README.md "$PKG_DIR/usr/share/doc/device-simulator/"
cp XEROX-SIMULATION.md "$PKG_DIR/usr/share/doc/device-simulator/"
cp DEBIAN-PACKAGING.md "$PKG_DIR/usr/share/doc/device-simulator/"

# Copy man page
cp debian/device-simulator.1 "$PKG_DIR/usr/share/man/man1/"
gzip -9 "$PKG_DIR/usr/share/man/man1/device-simulator.1"

# Copy validation script
cp validate-xerox-config.sh "$PKG_DIR/usr/share/device-simulator/"
chmod 755 "$PKG_DIR/usr/share/device-simulator/validate-xerox-config.sh"

# Create DEBIAN control files
cat > "$PKG_DIR/DEBIAN/control" << EOF
Package: $PACKAGE_NAME
Version: $VERSION
Architecture: $ARCH
Maintainer: $MAINTAINER
Depends: libc6
Section: net
Priority: optional
Homepage: https://github.com/fdurand/DeviceSimulator
Description: Network device simulator for testing and monitoring
 DeviceSimulator is a comprehensive network device simulation tool that can
 emulate various network devices and their protocols for testing, monitoring,
 and validation purposes.
 .
 Features include DHCP client simulation, RADIUS authentication and accounting,
 IPFIX flow generation, and UPnP device discovery.
EOF

# Create postinst script
cat > "$PKG_DIR/DEBIAN/postinst" << 'EOF'
#!/bin/bash
set -e

case "$1" in
    configure)
        # Create device-simulator user if it doesn't exist
        if ! getent passwd device-simulator > /dev/null 2>&1; then
            adduser --system --group --home /var/lib/device-simulator \
                    --no-create-home --disabled-login \
                    --gecos "DeviceSimulator Service User" device-simulator
        fi
        
        # Create necessary directories
        mkdir -p /var/lib/device-simulator
        mkdir -p /var/log/device-simulator
        
        # Set proper ownership
        chown device-simulator:device-simulator /var/lib/device-simulator
        chown device-simulator:device-simulator /var/log/device-simulator
        
        # Reload systemd
        systemctl daemon-reload || true
        
        echo "DeviceSimulator has been installed successfully."
        echo "Configuration files are located in /etc/device-simulator/"
        echo "To start: sudo systemctl start device-simulator@default"
        ;;
esac

exit 0
EOF

# Create prerm script
cat > "$PKG_DIR/DEBIAN/prerm" << 'EOF'
#!/bin/bash
set -e

case "$1" in
    remove|upgrade|deconfigure)
        # Stop running instances
        systemctl stop device-simulator@* 2>/dev/null || true
        ;;
esac

exit 0
EOF

# Create postrm script
cat > "$PKG_DIR/DEBIAN/postrm" << 'EOF'
#!/bin/bash
set -e

case "$1" in
    purge)
        # Remove user and directories on purge
        if getent passwd device-simulator > /dev/null 2>&1; then
            deluser device-simulator || true
        fi
        
        rmdir /var/lib/device-simulator 2>/dev/null || true
        rmdir /var/log/device-simulator 2>/dev/null || true
        ;;
esac

exit 0
EOF

# Make scripts executable
chmod 755 "$PKG_DIR/DEBIAN/postinst"
chmod 755 "$PKG_DIR/DEBIAN/prerm"
chmod 755 "$PKG_DIR/DEBIAN/postrm"

# Build the package
echo "Building .deb package..."
dpkg-deb --build "$PKG_DIR"

# Move to current directory
mv "debian-package/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb" .

echo "✅ Package built successfully: ${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
echo ""
echo "Install with: sudo dpkg -i ${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
echo "Remove with:  sudo dpkg -r ${PACKAGE_NAME}"
echo "Purge with:   sudo dpkg -P ${PACKAGE_NAME}"