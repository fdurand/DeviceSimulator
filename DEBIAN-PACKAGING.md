# Debian Packaging for DeviceSimulator

This document describes how to build and distribute DeviceSimulator as a Debian package.

## Prerequisites

Install the necessary packaging tools:

```bash
sudo apt-get update
sudo apt-get install devscripts build-essential debhelper-compat golang-go dh-golang dh-systemd
```

## Package Structure

The Debian package includes:

### Main Package: `device-simulator`
- **Binary**: `/usr/bin/device-simulator`
- **Configurations**: `/etc/device-simulator/`
  - `config.ini` (default configuration)
  - `config-xerox-printer.ini` (Xerox printer profile)
  - `xerox-dhcp-options.json` (DHCP options reference)
- **Systemd Service**: `/lib/systemd/system/device-simulator@.service`
- **Documentation**: `/usr/share/doc/device-simulator/`
- **Man Page**: `/usr/share/man/man1/device-simulator.1`
- **Validation Script**: `/usr/share/device-simulator/validate-xerox-config.sh`

### Service User
- **User**: `device-simulator` (system user)
- **Data Directory**: `/var/lib/device-simulator/`
- **Log Directory**: `/var/log/device-simulator/`

## Building the Package

### Quick Build
```bash
# Build binary package
make deb-build

# Build and install locally
make deb-install
```

### Manual Build
```bash
# Build binary package
debuild -us -uc -b

# Build source package
debuild -S -us -uc

# Clean build artifacts
debuild clean
```

### Build Artifacts
After building, you'll find these files in the parent directory:
- `device-simulator_1.0.0_amd64.deb` - Main package
- `device-simulator_1.0.0_amd64.changes` - Build changes
- `device-simulator_1.0.0_amd64.buildinfo` - Build information

## Installation and Usage

### Install Package
```bash
sudo dpkg -i device-simulator_1.0.0_amd64.deb
sudo apt-get install -f  # Fix any dependency issues
```

### Systemd Service Usage
```bash
# Start default simulation
sudo systemctl start device-simulator@default

# Start Xerox printer simulation  
sudo systemctl start device-simulator@xerox-printer

# Enable automatic startup
sudo systemctl enable device-simulator@default

# View logs
journalctl -u device-simulator@default -f
```

### Manual Execution
```bash
# Run default configuration
sudo device-simulator -file /etc/device-simulator/config.ini

# Run with debug output
sudo device-simulator -file /etc/device-simulator/config.ini -debug
```

## Package Management

### Remove Package
```bash
make deb-remove
# or
sudo dpkg -r device-simulator
```

### Completely Remove (including config)
```bash
make deb-purge  
# or
sudo dpkg -P device-simulator
```

## Configuration Management

### Default Configurations
The package installs example configurations in `/etc/device-simulator/`:
- Modify these files for your specific testing scenarios
- Create additional `config-*.ini` files for different device profiles
- Use with systemd: `systemctl start device-simulator@profile-name`

### Creating Custom Profiles
1. Create `/etc/device-simulator/config-mydevice.ini`
2. Start with: `sudo systemctl start device-simulator@mydevice`

## Security Considerations

### Capabilities
The systemd service uses Linux capabilities instead of full root privileges:
- `CAP_NET_RAW` - For raw socket operations
- `CAP_NET_ADMIN` - For network interface management

### Service Hardening
The systemd service includes security hardening:
- `NoNewPrivileges=true`
- `PrivateTmp=true`  
- `ProtectHome=true`
- `ProtectSystem=strict`
- Restricted filesystem access

## Package Maintenance

### Version Updates
1. Update `debian/changelog` with new version and changes
2. Rebuild package: `make deb-build`
3. Test installation: `make deb-install`

### Changelog Format
```
device-simulator (1.1.0) unstable; urgency=medium

  * New features and improvements
  * Bug fixes and optimizations
  
 -- Maintainer Name <email@example.com>  Date
```

## Troubleshooting

### Common Issues

**Permission Errors**
```bash
# Ensure proper capabilities
sudo setcap cap_net_raw,cap_net_admin+ep /usr/bin/device-simulator
```

**Network Interface Issues**
```bash
# Check available interfaces
ip link show

# Verify configuration
sudo device-simulator -file /etc/device-simulator/config.ini -debug
```

**Service Failures**
```bash
# Check service status
systemctl status device-simulator@default

# View detailed logs
journalctl -u device-simulator@default -n 50
```

### Log Files
- **System Logs**: `journalctl -u device-simulator@*`
- **Application Logs**: `/var/log/device-simulator/` (if file logging enabled)

## Distribution

### Local Repository
Set up a local APT repository:
```bash
# Install repository tools
sudo apt-get install dpkg-scanpackages gzip

# Create repository structure  
mkdir -p /path/to/repo/binary
cp device-simulator_*.deb /path/to/repo/binary/

# Generate package list
cd /path/to/repo
dpkg-scanpackages binary /dev/null | gzip -9c > binary/Packages.gz
```

### PPA Upload
For Ubuntu PPA distribution:
```bash
# Build source package
debuild -S

# Upload to PPA
dput ppa:username/ppa-name device-simulator_1.0.0_source.changes
```

This comprehensive packaging setup provides professional-grade Debian package management for DeviceSimulator with proper systemd integration, security hardening, and maintainable configuration management.