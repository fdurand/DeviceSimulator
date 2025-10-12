package main

import (
	"context"
	"sync/atomic"
	"time"
)

// Metrics provides runtime performance monitoring
type Metrics struct {
	DHCPRequests    int64
	RADIUSRequests  int64
	IPFIXPackets    int64
	UPnPDiscoveries int64
	Errors          int64
	StartTime       time.Time
}

var metrics = &Metrics{
	StartTime: time.Now(),
}

// IncrementDHCP atomically increments DHCP request counter
func (m *Metrics) IncrementDHCP() {
	atomic.AddInt64(&m.DHCPRequests, 1)
}

// IncrementRADIUS atomically increments RADIUS request counter
func (m *Metrics) IncrementRADIUS() {
	atomic.AddInt64(&m.RADIUSRequests, 1)
}

// IncrementIPFIX atomically increments IPFIX packet counter
func (m *Metrics) IncrementIPFIX() {
	atomic.AddInt64(&m.IPFIXPackets, 1)
}

// IncrementUPnP atomically increments UPnP discovery counter
func (m *Metrics) IncrementUPnP() {
	atomic.AddInt64(&m.UPnPDiscoveries, 1)
}

// IncrementErrors atomically increments error counter
func (m *Metrics) IncrementErrors() {
	atomic.AddInt64(&m.Errors, 1)
}

// GetUptime returns the application uptime
func (m *Metrics) GetUptime() time.Duration {
	return time.Since(m.StartTime)
}

// PrintStats prints current statistics
func (m *Metrics) PrintStats() {
	logger.Info("=== DeviceSimulator Statistics ===")
	logger.Info("Uptime: %v", m.GetUptime())
	logger.Info("DHCP Requests: %d", atomic.LoadInt64(&m.DHCPRequests))
	logger.Info("RADIUS Requests: %d", atomic.LoadInt64(&m.RADIUSRequests))
	logger.Info("IPFIX Packets: %d", atomic.LoadInt64(&m.IPFIXPackets))
	logger.Info("UPnP Discoveries: %d", atomic.LoadInt64(&m.UPnPDiscoveries))
	logger.Info("Errors: %d", atomic.LoadInt64(&m.Errors))
	logger.Info("================================")
}

// StartMetricsReporter starts a goroutine that periodically reports metrics
func StartMetricsReporter(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Debug("Metrics reporter stopped")
				return
			case <-ticker.C:
				metrics.PrintStats()
			}
		}
	}()
}
