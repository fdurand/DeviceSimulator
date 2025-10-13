# Build stage
FROM golang:1.24.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o device-simulator .

# Final stage
FROM scratch

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/device-simulator /usr/bin/device-simulator

# Copy configuration files
COPY --from=builder /app/config.ini /etc/device-simulator/config.ini
COPY --from=builder /app/config-xerox-printer.ini /etc/device-simulator/config-xerox-printer.ini
COPY --from=builder /app/xerox-dhcp-options.json /etc/device-simulator/xerox-dhcp-options.json

# Create directories for runtime data
VOLUME ["/var/lib/device-simulator", "/var/log/device-simulator"]

# Set user (note: container will still need --privileged for raw sockets)
USER 1000:1000

# Expose no ports by default (device simulator doesn't listen)

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["device-simulator", "-help"] || exit 1

# Default command
ENTRYPOINT ["/usr/bin/device-simulator"]
CMD ["-file", "/etc/device-simulator/config.ini"]

# Labels for metadata
LABEL maintainer="Fabrice Durand <fdurand@inverse.ca>"
LABEL description="DeviceSimulator - Network device simulation for testing and monitoring"
LABEL version="1.0.0"
LABEL org.opencontainers.image.source="https://github.com/fdurand/DeviceSimulator"
LABEL org.opencontainers.image.documentation="https://github.com/fdurand/DeviceSimulator/blob/main/README.md"
LABEL org.opencontainers.image.licenses="Apache-2.0"