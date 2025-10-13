# Xerox Printer Device Simulation

## Overview

This configuration simulates a **Xerox VersaLink C405** multifunction printer with MAC address `f0:6d:ab:74:f5:a2` connected to switch port **GigabitEthernet1/0/12**.

## Configuration Details

### Network Settings
- **MAC Address**: `f0:6d:ab:74:f5:a2` (Xerox OUI: f0:6d:ab)
- **IP Address**: `10.10.1.45`
- **Switch Port**: `GigabitEthernet1/0/12`
- **Network Interface**: `eth0`

### DHCP Options (Xerox-Specific)
- **Option 12**: `XRX-VersaLink-C405` (Hostname)
- **Option 55**: Parameter Request List including printer-specific options
- **Option 60**: Vendor Class ID with Xerox manufacturer details
- **Option 43**: Vendor-specific information
- **Option 125**: Vendor-identifying vendor class
- **Option 61**: Client identifier (MAC address)

### RADIUS Authentication
- **Port**: GigabitEthernet1/0/12 (NAS-Port-Id)
- **NAS-Port**: 12
- **Device Type**: Ethernet connected printer
- **Session ID**: XRX-C405-f06dab74f5a2-000001

### Simulated IPFIX Traffic Flows

The configuration generates realistic printer traffic patterns:

1. **Print Jobs**
   - LPR (Line Printer Remote) - Port 515 → 9100
   - IPP (Internet Printing Protocol) - Port 631 → 80
   - Raw printing data - Large byte counts (2MB+ typical)

2. **Scanning Operations**
   - SMB scanning to network folders - Port 9100 → 445
   - Email scanning via SMTP - Port 25 → 587
   - FTP scan-to-folder - Port 21 → 990

3. **Management Traffic**
   - Web interface access - Port 80 ↔ 443
   - SNMP monitoring - Port 161 ↔ 162
   - Firmware updates via DNS - Port 53

4. **Network Discovery**
   - UPnP device discovery - Port 1900 (multicast)
   - DHCP renewals - Port 68 → 67

### Usage

Run the Xerox printer simulation:

```bash
# Build the simulator
make build

# Run with Xerox configuration
sudo ./bin/device-simulator -file config-xerox-printer.ini

# Enable debug logging
sudo ./bin/device-simulator -file config-xerox-printer.ini -debug
```

### Realistic Behavior Patterns

The simulation includes:
- **Periodic DHCP renewals** (every 3600 seconds - typical for printers)
- **UPnP device announcements** for network discovery
- **Variable traffic patterns** representing different printer operations
- **Appropriate packet sizes** for different protocols
- **Realistic port combinations** used by enterprise printers

### Monitoring

The simulator will generate:
- DHCP requests with Xerox vendor identification
- RADIUS authentication for network access control
- IPFIX flows showing typical printer communication patterns
- UPnP discovery packets for device visibility

This provides a comprehensive simulation of an enterprise Xerox printer for network testing, monitoring validation, and security analysis.

## Network Security Considerations

This simulation helps test:
- **802.1X authentication** for printer devices
- **VLAN assignment** based on device type
- **Network access control** policies
- **Traffic monitoring** and anomaly detection
- **DHCP option filtering** and validation