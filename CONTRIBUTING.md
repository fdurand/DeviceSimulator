# Contributing to DeviceSimulator

Thank you for your interest in contributing to DeviceSimulator! This document provides guidelines and information for contributors.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Performance Guidelines](#performance-guidelines)
- [Documentation](#documentation)

## Code of Conduct

This project adheres to our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Git
- Linux system (required for raw socket operations)
- Basic understanding of networking protocols (DHCP, RADIUS, IPFIX, UPnP)

### Fork and Clone
1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/DeviceSimulator.git
   cd DeviceSimulator
   ```

## Development Setup

### Install Dependencies
```bash
# Install Go dependencies
go mod tidy

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

### Build the Project
```bash
# Build for current platform
go build -o devicesimulator .

# Build with optimizations
go build -ldflags="-s -w" -o devicesimulator .

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o devicesimulator-linux-amd64 .
```

### Run Tests
```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...

# Run with coverage
go test -cover ./...
```

## Making Changes

### Branch Naming
Use descriptive branch names:
- `feature/add-snmp-simulation`
- `bugfix/fix-dhcp-memory-leak`
- `docs/update-configuration-guide`
- `perf/optimize-packet-generation`

### Commit Messages
Follow conventional commit format:
```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Test changes
- `chore`: Build/tooling changes

Examples:
```
feat(dhcp): add support for option 82 relay agent information

fix(radius): resolve memory leak in client pool
- Initialize pool with proper cleanup
- Add finalizer for connection cleanup
- Update tests to verify memory usage

perf(config): implement configuration caching
- Add atomic cache for frequent config access
- Reduces config read time from 1µs to 41ns
- Add benchmarks to verify performance gains
```

## Testing

### Writing Tests
- Write tests for all new functionality
- Maintain or improve test coverage
- Use table-driven tests for multiple test cases
- Include benchmarks for performance-critical code

### Test Categories
1. **Unit Tests**: Test individual functions/methods
2. **Integration Tests**: Test component interactions
3. **Benchmark Tests**: Measure performance
4. **Race Tests**: Detect concurrency issues

### Example Test Structure
```go
func TestDHCPSimulation(t *testing.T) {
    tests := []struct {
        name     string
        config   Config
        expected Result
        wantErr  bool
    }{
        {
            name: "valid DHCP request",
            config: Config{
                MACAddress: "aa:bb:cc:dd:ee:ff",
                Interface:  "eth0",
            },
            expected: Result{Success: true},
            wantErr:  false,
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := SimulateDHCP(tt.config)
            if (err != nil) != tt.wantErr {
                t.Errorf("SimulateDHCP() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("SimulateDHCP() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Benchmarks
Include benchmarks for performance-critical code:
```go
func BenchmarkConfigAccess(b *testing.B) {
    config := NewConfig()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _ = config.GetValue("key")
    }
}
```

## Submitting Changes

### Before Submitting
1. **Run all tests**: `go test ./...`
2. **Check code formatting**: `gofmt -s -w .`
3. **Run static analysis**: `go vet ./...` and `staticcheck ./...`
4. **Security scan**: `gosec ./...`
5. **Update documentation** if needed
6. **Add/update tests** for your changes

### Pull Request Process
1. **Create a pull request** against the `main` branch
2. **Fill out the PR template** completely
3. **Ensure CI passes** (all checks must be green)
4. **Request review** from maintainers
5. **Address feedback** promptly and respectfully

### PR Checklist
- [ ] Tests pass locally and in CI
- [ ] Code follows project style guidelines
- [ ] Documentation updated (if applicable)
- [ ] Breaking changes documented
- [ ] Performance impact assessed
- [ ] Security implications considered

## Code Style

### Go Style Guidelines
Follow standard Go conventions:
- Use `gofmt` for formatting
- Follow effective Go practices
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Handle errors explicitly

### Project-Specific Guidelines
- Use structured logging with appropriate levels
- Implement graceful shutdown for long-running operations
- Use atomic operations for metrics and counters
- Prefer composition over inheritance
- Write self-documenting code

### Example Code Style
```go
// SimulateDHCPRequest generates a DHCP request for the specified device.
// It returns the server response and any error encountered.
func SimulateDHCPRequest(ctx context.Context, device *Device) (*DHCPResponse, error) {
    if device == nil {
        return nil, errors.New("device cannot be nil")
    }

    // Log the simulation attempt
    log.Info("Starting DHCP simulation",
        "mac", device.MACAddress,
        "interface", device.Interface,
    )

    // Create DHCP request packet
    request, err := buildDHCPRequest(device)
    if err != nil {
        return nil, fmt.Errorf("failed to build DHCP request: %w", err)
    }

    // Send request and wait for response
    response, err := sendDHCPRequest(ctx, request)
    if err != nil {
        return nil, fmt.Errorf("DHCP request failed: %w", err)
    }

    log.Info("DHCP simulation completed successfully",
        "mac", device.MACAddress,
        "assigned_ip", response.AssignedIP,
    )

    return response, nil
}
```

## Performance Guidelines

### Performance Best Practices
1. **Use connection pooling** for network resources
2. **Implement caching** for frequently accessed data
3. **Use atomic operations** for counters and metrics
4. **Avoid memory allocations** in hot paths
5. **Profile critical code paths** with benchmarks
6. **Consider concurrency** but avoid premature optimization

### Performance Testing
- Include benchmarks for new performance-critical code
- Compare before/after performance metrics
- Document performance improvements in PRs
- Set performance regression alerts where appropriate

### Memory Management
- Use object pooling for frequently allocated objects
- Implement proper cleanup in defer statements
- Monitor goroutine usage and prevent leaks
- Use build constraints for platform-specific optimizations

## Documentation

### Code Documentation
- Document all exported functions and types
- Include examples in documentation comments
- Explain complex algorithms or business logic
- Document performance characteristics where relevant

### User Documentation
- Update README.md for user-facing changes
- Add configuration examples
- Update troubleshooting guides
- Include migration guides for breaking changes

### Example Documentation
```go
// Config represents the device simulation configuration.
// It supports hot-reloading and provides thread-safe access
// to configuration values with sub-microsecond latency.
//
// Example usage:
//   config := NewConfig("device.ini")
//   macAddr := config.GetString("mac_address")
//   timeout := config.GetDuration("timeout")
//
// Performance: Configuration access is optimized with atomic
// caching, providing ~41ns access time for cached values.
type Config struct {
    // ... fields
}
```

## Getting Help

### Communication Channels
- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and community discussions
- **Pull Request Reviews**: Code-specific questions

### Development Questions
Before asking questions:
1. Check existing issues and discussions
2. Review the documentation
3. Search the codebase for similar implementations
4. Try to create a minimal reproduction case

When asking questions:
- Provide context and relevant details
- Include error messages and logs
- Describe what you've already tried
- Be specific about your environment

## Recognition

Contributors are recognized in several ways:
- Listed in release notes for significant contributions
- Mentioned in commit messages and PR descriptions
- Invited to become maintainers for sustained contributions
- Featured in project documentation for major features

Thank you for contributing to DeviceSimulator! 🎉