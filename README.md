# DeviceSimulator

A Go-based network device simulator that handles DHCP, RADIUS authentication, IPFIX, and UPnP protocols.

## Features

- DHCP client simulation
- RADIUS authentication
- IPFIX data export
- UPnP device discovery
- Raw socket communication
- Configurable network interface binding

## Prerequisites

- Go 1.23.1 or later
- Linux operating system (requires raw socket capabilities)
- sudo privileges (for raw socket operations)

## Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the project:
   ```bash
   go build -o bin/device-simulator .
   ```

## Configuration

Edit the `config.ini` file to configure the simulator:

```ini
[general]
# MAC address of the device to simulate
clientmac=90:6c:ac:64:95:c1
# Network interface to use
interface=eth0

[dhcp]
enabled=true
server=10.10.1.1
renew=30
giaddr=10.10.20.1
ciaddr=10.10.1.22
srcmac=90:6c:ac:64:95:c1
dstmac=fe:ff:ff:ff:ff:ff
```

## Usage

Run the simulator with appropriate privileges:

```bash
# Default simulation
sudo ./bin/device-simulator -file config.ini

# Xerox printer simulation
sudo ./bin/device-simulator -file config-xerox-printer.ini

# Enable debug logging
sudo ./bin/device-simulator -file config.ini -debug
```

### Device Simulations

- **`config.ini`**: Generic device simulation
- **`config-xerox-printer.ini`**: Xerox VersaLink C405 printer simulation (see [XEROX-SIMULATION.md](XEROX-SIMULATION.md))

## Development

### Quick Start with Make

```bash
# Install development tools
make tools

# Run tests and build optimized binary
make all

# Run the simulator
make run

# Build debug version with race detection
make debug

# Run benchmarks
make bench

# Generate coverage report
make coverage
```

### VS Code Tasks

- **Build**: `Ctrl+Shift+P` → "Tasks: Run Task" → "Build DeviceSimulator"
- **Run**: `Ctrl+Shift+P` → "Tasks: Run Task" → "Run DeviceSimulator"
- **Test**: `Ctrl+Shift+P` → "Tasks: Run Task" → "Go Test"

### Debugging

Use the provided launch configurations:
- **Launch DeviceSimulator**: Standard execution  
- **Debug DeviceSimulator**: Debug mode with breakpoints

### Performance Optimization Features

- **Centralized Configuration Management**: Cached config loading with thread-safety
- **Connection Pooling**: Efficient RADIUS client reuse
- **Rate Limiting**: Built-in request rate limiting
- **Metrics & Monitoring**: Real-time performance metrics
- **Memory Optimization**: Reduced allocations and garbage collection pressure
- **Network Interface Caching**: Cached network interface lookups
- **Structured Logging**: Efficient logging with multiple levels
- **Graceful Shutdown**: Proper resource cleanup

## Dependencies

- `github.com/coreos/go-systemd`: systemd integration
- `github.com/krolaw/dhcp4`: DHCP protocol implementation
- `github.com/mdlayher/ethernet`: Ethernet frame handling
- `github.com/mdlayher/raw`: Raw socket operations
- `gopkg.in/ini.v1`: Configuration file parsing
- `layeh.com/radius`: RADIUS protocol implementation

## License

See LICENSE file for details.