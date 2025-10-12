package main

import (
	"net"
	"time"
)

// Optimized configuration methods using the ConfigManager

// readDhcpConfigOptimized uses the ConfigManager for better performance
func (d *Interface) readDhcpConfigOptimized() {
	d.Enabled = configManager.GetBool("dhcp", "enabled", false)
	d.ServerIP = configManager.GetIP("dhcp", "server", net.IPv4zero)
	d.GiAddr = configManager.GetIP("dhcp", "giaddr", net.IPv4zero)
	d.CiAddr = configManager.GetIP("dhcp", "ciaddr", net.IPv4zero)

	// Default to interface MAC if not specified
	d.SrcMac = configManager.GetMAC("dhcp", "srcmac", d.intNet.HardwareAddr)

	// Default to broadcast MAC
	broadcastMAC, _ := net.ParseMAC("FF:FF:FF:FF:FF:FF")
	d.DstMac = configManager.GetMAC("dhcp", "dstmac", broadcastMAC)

	// Renew interval with reasonable default and limits
	d.Renew = configManager.GetDuration("dhcp", "renew", 30*time.Second)

	// DHCP options
	d.Options = configManager.GetString("dhcp", "options", "[]")

	logger.Info("DHCP configured - Enabled: %v, Server: %v, Renew: %v",
		d.Enabled, d.ServerIP, d.Renew)
}

// readUpnpConfigOptimized uses the ConfigManager for better performance
func (u *Upnp) readUpnpConfigOptimized() {
	u.Enabled = configManager.GetBool("upnp", "enabled", false)
	u.UserAgent = configManager.GetString("upnp", "useragent", "siemens ag simatic s7")
	u.deviceType = configManager.GetString("upnp", "devicetype", "urn:schemas-upnp-org:device:InternetGatewayDevice:1")
	u.IPAddr = configManager.GetIP("upnp", "ipaddr", net.ParseIP("239.255.255.250"))
	u.UDPPort = configManager.GetInt("upnp", "udpport", 1900, 1, 65535)

	logger.Info("UPnP configured - Enabled: %v, IP: %v, Port: %d",
		u.Enabled, u.IPAddr, u.UDPPort)
}

// ReadRadiusAccountingConfigOptimized uses the ConfigManager for better performance
func (a *Accounting) ReadRadiusAccountingConfigOptimized() {
	a.Enabled = configManager.GetBool("accounting", "enabled", false)

	serverIP := configManager.GetString("accounting", "server", "")
	if serverIP != "" {
		a.ServerIP = net.ParseIP(serverIP)
		if a.ServerIP == nil {
			logger.Warn("Invalid accounting server IP: %s", serverIP)
			a.Enabled = false
		}
	}

	a.Secret = configManager.GetString("accounting", "secret", "secret")
	a.UserName = configManager.GetString("accounting", "User-Name", "")
	a.AcctSessionId = configManager.GetString("accounting", "Acct-Session-Id", "")
	a.CallingStationId = configManager.GetString("accounting", "Calling-Station-Id", "")
	a.CalledStationId = configManager.GetString("accounting", "Called-Station-Id", "")
	a.NASPort = configManager.GetString("accounting", "NAS-Port", "")
	a.NASPortType = configManager.GetString("accounting", "NAS-Port-Type", "")
	a.FramedIPAddress = configManager.GetString("accounting", "Framed-IP-Address", "")
	a.NASIdentifier = configManager.GetString("accounting", "NAS-Identifier", "")
	a.NASPortId = configManager.GetString("accounting", "NAS-Port-Id", "")
	a.NASIPAddress = configManager.GetString("accounting", "NAS-IP-Address", "")

	logger.Info("RADIUS Accounting configured - Enabled: %v, Server: %v",
		a.Enabled, a.ServerIP)
}

// ReadRadiusAuthenticationConfigOptimized uses the ConfigManager for better performance
func (a *Authentication) ReadRadiusAuthenticationConfigOptimized() {
	a.Enabled = configManager.GetBool("authentication", "enabled", false)

	serverIP := configManager.GetString("authentication", "server", "")
	if serverIP != "" {
		a.ServerIP = net.ParseIP(serverIP)
		if a.ServerIP == nil {
			logger.Warn("Invalid authentication server IP: %s", serverIP)
			a.Enabled = false
		}
	}

	a.Secret = configManager.GetString("authentication", "secret", "secret")
	a.UserName = configManager.GetString("authentication", "User-Name", "")
	a.CallingStationId = configManager.GetString("authentication", "Calling-Station-Id", "")
	a.CalledStationId = configManager.GetString("authentication", "Called-Station-Id", "")
	a.NASPort = configManager.GetString("authentication", "NAS-Port", "")
	a.NASPortType = configManager.GetString("authentication", "NAS-Port-Type", "")
	a.FramedIPAddress = configManager.GetString("authentication", "Framed-IP-Address", "")
	a.NASIdentifier = configManager.GetString("authentication", "NAS-Identifier", "")
	a.NASPortId = configManager.GetString("authentication", "NAS-Port-Id", "")
	a.NASIPAddress = configManager.GetString("authentication", "NAS-IP-Address", "")

	logger.Info("RADIUS Authentication configured - Enabled: %v, Server: %v",
		a.Enabled, a.ServerIP)
}

// readIpFixConfigOptimized uses the ConfigManager for better performance
func (i *IpFix) readIpFixConfigOptimized() {
	i.Enabled = configManager.GetBool("ipfix", "enabled", false)
	i.DestinationIP = configManager.GetIP("ipfix", "destination_ip", net.ParseIP("127.0.0.1"))
	i.DestinationPort = configManager.GetInt("ipfix", "destination_port", 4739, 1, 65535)
	i.Traffic = configManager.GetString("ipfix", "traffic", "[]")

	logger.Info("IPFIX configured - Enabled: %v, Destination: %v:%d",
		i.Enabled, i.DestinationIP, i.DestinationPort)
}
