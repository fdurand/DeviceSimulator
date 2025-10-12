package main

import (
	"net"
	"testing"
	"time"
)

// BenchmarkConfigManager tests configuration loading performance
func BenchmarkConfigManager(b *testing.B) {
	// Create a temporary config for testing
	configManager := &ConfigManager{
		cache: make(map[string]interface{}),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		configManager.GetBool("dhcp", "enabled", false)
		configManager.GetString("dhcp", "server", "10.0.0.1")
		configManager.GetInt("dhcp", "renew", 30, 1, 3600)
	}
}

// BenchmarkMACParsing tests MAC address parsing performance
func BenchmarkMACParsing(b *testing.B) {
	macStr := "90:6c:ac:64:95:c1"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := net.ParseMAC(macStr)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNetworkInterfaceCache tests network interface caching
func BenchmarkNetworkInterfaceCache(b *testing.B) {
	ni := &NetworkInterface{
		interfaces: make(map[string]*net.Interface),
	}

	// Pre-populate cache with a fake interface for testing
	fakeInterface := &net.Interface{
		Name: "eth0",
		MTU:  1500,
	}
	ni.interfaces["eth0"] = fakeInterface

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ni.GetInterface("eth0")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRateLimiter tests rate limiter performance
func BenchmarkRateLimiter(b *testing.B) {
	rl := NewRateLimiter(1000) // 1000 requests per second
	defer rl.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if !rl.TryWait() {
				// Rate limited
			}
		}
	})
}

// BenchmarkMetricsIncrement tests metrics increment performance
func BenchmarkMetricsIncrement(b *testing.B) {
	m := &Metrics{
		StartTime: time.Now(),
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.IncrementDHCP()
		}
	})
}

// BenchmarkIPParsing tests IP address parsing performance
func BenchmarkIPParsing(b *testing.B) {
	ipStr := "192.168.1.100"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			b.Fatal("Failed to parse IP")
		}
	}
}

// TestConfigManagerConcurrency tests thread safety of ConfigManager
func TestConfigManagerConcurrency(t *testing.T) {
	cm := &ConfigManager{
		cache: make(map[string]interface{}),
	}

	// Simulate concurrent access
	done := make(chan bool)

	// Start multiple goroutines reading configuration
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				cm.GetBool("test", "enabled", false)
				cm.GetString("test", "value", "default")
				cm.GetInt("test", "number", 42, 0, 100)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestMetricsConcurrency tests thread safety of metrics
func TestMetricsConcurrency(t *testing.T) {
	m := &Metrics{
		StartTime: time.Now(),
	}

	done := make(chan bool)

	// Start multiple goroutines incrementing counters
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				m.IncrementDHCP()
				m.IncrementRADIUS()
				m.IncrementIPFIX()
				m.IncrementUPnP()
				m.IncrementErrors()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify counters have expected values
	if m.DHCPRequests != 10000 {
		t.Errorf("Expected 10000 DHCP requests, got %d", m.DHCPRequests)
	}
}
