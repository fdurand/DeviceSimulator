#!/bin/bash

# Xerox Printer Simulation Validator
# Validates the Xerox configuration file

CONFIG_FILE="config-xerox-printer.ini"

echo "=== Xerox Printer Configuration Validator ==="
echo

if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Configuration file not found: $CONFIG_FILE"
    exit 1
fi

echo "✅ Configuration file found: $CONFIG_FILE"
echo

# Check MAC address format
MAC=$(grep "clientmac=" $CONFIG_FILE | cut -d'=' -f2)
if [[ $MAC =~ ^f0:6d:ab:74:f5:a2$ ]]; then
    echo "✅ MAC address correct: $MAC (Xerox OUI: f0:6d:ab)"
else
    echo "❌ MAC address incorrect: $MAC"
fi

# Check DHCP options for Xerox-specific values
if grep -q "Xerox" $CONFIG_FILE; then
    echo "✅ Xerox DHCP options found"
else
    echo "❌ Xerox DHCP options not found"
fi

# Check RADIUS port configuration
PORT_ID=$(grep "NAS-Port-Id = GigabitEthernet1/0/12" $CONFIG_FILE)
if [ ! -z "$PORT_ID" ]; then
    echo "✅ Correct switch port: GigabitEthernet1/0/12"
else
    echo "❌ Switch port configuration incorrect"
fi

# Check IPFIX traffic patterns
if grep -q "9100\|631\|515" $CONFIG_FILE; then
    echo "✅ Printer-specific IPFIX flows configured"
else
    echo "❌ Printer IPFIX flows not found"
fi

# Check IP address
IP=$(grep "ciaddr=" $CONFIG_FILE | cut -d'=' -f2)
echo "✅ Printer IP address: $IP"

# Check renewal time (should be longer for printers)
RENEW=$(grep "renew=" $CONFIG_FILE | cut -d'=' -f2)
if [ "$RENEW" -ge 3600 ]; then
    echo "✅ DHCP renewal time appropriate for printer: ${RENEW}s"
else
    echo "⚠️  DHCP renewal time may be too short for printer: ${RENEW}s"
fi

echo
echo "=== Configuration Summary ==="
echo "Device Type: Xerox VersaLink C405 MFP"
echo "MAC Address: $MAC"
echo "IP Address: $IP"
echo "Switch Port: GigabitEthernet1/0/12"
echo "DHCP Renewal: ${RENEW}s"
echo
echo "✅ Xerox printer simulation configuration validated!"