package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"gopkg.in/ini.v1"
)

// ConfigManager provides centralized, cached configuration management
type ConfigManager struct {
	mu     sync.RWMutex
	cfg    *ini.File
	cache  map[string]interface{}
	loaded bool
}

var configManager = &ConfigManager{
	cache: make(map[string]interface{}),
}

// LoadConfig loads and caches the configuration file
func (cm *ConfigManager) LoadConfig(configFile string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := SafeConfigRead(configFile); err != nil {
		return err
	}

	cfg, err := ini.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config file: %v", err)
	}

	cm.cfg = cfg
	cm.loaded = true

	// Pre-load critical configuration into cache
	cm.preloadCache()

	logger.Info("Configuration loaded successfully from %s", configFile)
	return nil
}

// preloadCache loads frequently accessed config values into memory
func (cm *ConfigManager) preloadCache() {
	// Cache network interface
	if interfaceName := cm.cfg.Section("general").Key("interface").String(); interfaceName != "" {
		if intf, err := net.InterfaceByName(interfaceName); err == nil {
			cm.cache["interface"] = intf
		}
	}

	// Cache MAC addresses
	if clientMAC := cm.cfg.Section("general").Key("clientmac").String(); clientMAC != "" {
		if mac, err := net.ParseMAC(clientMAC); err == nil {
			cm.cache["clientmac"] = mac
		}
	}
}

// GetInterface returns the cached network interface
func (cm *ConfigManager) GetInterface() (*net.Interface, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.loaded {
		return nil, fmt.Errorf("configuration not loaded")
	}

	if intf, exists := cm.cache["interface"]; exists {
		return intf.(*net.Interface), nil
	}

	interfaceName := cm.cfg.Section("general").Key("interface").String()
	if interfaceName == "" {
		return nil, fmt.Errorf("no interface specified in configuration")
	}

	intf, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to find interface %s: %v", interfaceName, err)
	}

	cm.cache["interface"] = intf
	return intf, nil
}

// GetClientMAC returns the client MAC address with fallback
func (cm *ConfigManager) GetClientMAC() net.HardwareAddr {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if mac, exists := cm.cache["clientmac"]; exists {
		return mac.(net.HardwareAddr)
	}

	clientMAC := cm.cfg.Section("general").Key("clientmac").String()
	if mac, err := net.ParseMAC(clientMAC); err == nil {
		cm.cache["clientmac"] = mac
		return mac
	}

	// Fallback to default MAC
	defaultMAC, _ := net.ParseMAC("de:ad:be:ef:de:ad")
	logger.Warn("Using default MAC address: %s", defaultMAC.String())
	return defaultMAC
}

// GetBool safely gets a boolean value with default
func (cm *ConfigManager) GetBool(section, key string, defaultVal bool) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.loaded {
		return defaultVal
	}

	val := cm.cfg.Section(section).Key(key).String()
	switch val {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		if val != "" {
			logger.Warn("Invalid boolean value '%s' for %s.%s, using default %v", val, section, key, defaultVal)
		}
		return defaultVal
	}
}

// GetString safely gets a string value
func (cm *ConfigManager) GetString(section, key, defaultVal string) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.loaded {
		return defaultVal
	}

	val := cm.cfg.Section(section).Key(key).String()
	if val == "" {
		return defaultVal
	}
	return val
}

// GetInt safely gets an integer value with validation
func (cm *ConfigManager) GetInt(section, key string, defaultVal, min, max int) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.loaded {
		return defaultVal
	}

	val := cm.cfg.Section(section).Key(key).String()
	if val == "" {
		return defaultVal
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		logger.Warn("Invalid integer value '%s' for %s.%s, using default %d", val, section, key, defaultVal)
		return defaultVal
	}

	if intVal < min || intVal > max {
		logger.Warn("Integer value %d for %s.%s out of range [%d,%d], using default %d", intVal, section, key, min, max, defaultVal)
		return defaultVal
	}

	return intVal
}

// GetIP safely parses an IP address
func (cm *ConfigManager) GetIP(section, key string, defaultIP net.IP) net.IP {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.loaded {
		return defaultIP
	}

	val := cm.cfg.Section(section).Key(key).String()
	if val == "" {
		return defaultIP
	}

	ip := net.ParseIP(val)
	if ip == nil {
		logger.Warn("Invalid IP address '%s' for %s.%s, using default %v", val, section, key, defaultIP)
		return defaultIP
	}

	return ip
}

// GetMAC safely parses a MAC address
func (cm *ConfigManager) GetMAC(section, key string, defaultMAC net.HardwareAddr) net.HardwareAddr {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.loaded {
		return defaultMAC
	}

	val := cm.cfg.Section(section).Key(key).String()
	if val == "" {
		return defaultMAC
	}

	mac, err := net.ParseMAC(val)
	if err != nil {
		logger.Warn("Invalid MAC address '%s' for %s.%s, using default %v", val, section, key, defaultMAC)
		return defaultMAC
	}

	return mac
}

// GetDuration safely parses a duration
func (cm *ConfigManager) GetDuration(section, key string, defaultDuration time.Duration) time.Duration {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.loaded {
		return defaultDuration
	}

	val := cm.cfg.Section(section).Key(key).String()
	if val == "" {
		return defaultDuration
	}

	// Try parsing as seconds first (for backward compatibility)
	if seconds, err := strconv.Atoi(val); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as duration string
	if duration, err := time.ParseDuration(val); err == nil {
		return duration
	}

	logger.Warn("Invalid duration '%s' for %s.%s, using default %v", val, section, key, defaultDuration)
	return defaultDuration
}
