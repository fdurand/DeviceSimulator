package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/krolaw/dhcp4"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

type GlobalConfig struct {
	intNet    *net.Interface
	ClientMAC net.HardwareAddr // Device Client MAC (In the DHCP Request)
}

type Config struct {
	ConfigFile string
}

func (c *GlobalConfig) ReadGlobalConfig(config *Config) {
	// This method is deprecated, use ConfigManager instead
	var err error
	c.intNet, err = configManager.GetInterface()
	if err != nil {
		logger.Fatal("Failed to get network interface: %v", err)
	}
	c.ClientMAC = configManager.GetClientMAC()
}

// startSystemdWatchdog starts the systemd watchdog if enabled
func startSystemdWatchdog(ctx context.Context) {
	daemon.SdNotify(false, "READY=1")
	interval, err := daemon.SdWatchdogEnabled(false)
	if err != nil || interval == 0 {
		logger.Debug("Systemd watchdog not enabled")
		return
	}

	logger.Info("Systemd watchdog enabled with interval: %v", interval)

	ticker := time.NewTicker(interval / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Systemd watchdog stopped")
			return
		case <-ticker.C:
			daemon.SdNotify(false, "WATCHDOG=1")
		}
	}
}

func main() {
	// Parse command line flags
	configFile := flag.String("file", "/usr/local/etc/config.ini", "Configuration File Path")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Initialize logger based on debug flag
	logLevel := INFO
	if *debug {
		logLevel = DEBUG
	}
	logger = NewLogger(logLevel)

	// Setup graceful shutdown
	shutdown := NewGracefulShutdown()

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	shutdown.Register(func() error {
		cancel()
		return nil
	})

	// Load configuration
	if err := configManager.LoadConfig(*configFile); err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}

	// Start systemd watchdog
	go startSystemdWatchdog(ctx)

	// Get network interface and client MAC with better error handling
	netInterface, err := configManager.GetInterface()
	if err != nil {
		logger.Fatal("Failed to get network interface: %v", err)
	}
	clientMAC := configManager.GetClientMAC()

	logger.Info("Using interface: %s, Client MAC: %s", netInterface.Name, clientMAC.String())

	// Initialize DHCP interface
	var d Interface
	d.intNet = netInterface
	d.ClientMAC = clientMAC
	d.readDhcpConfigOptimized()

	// Initialize UPnP
	var u Upnp
	u.readUpnpConfigOptimized()
	u.intNet = netInterface

	if u.Enabled {
		fmt.Println("UPnP Discovery is enabled")
		go func() {
			for {
				u.discover(time.Second * 10)

				time.Sleep(time.Second * 30)
			}
		}()
	}

	// Initialize RADIUS Accounting
	var acct Accounting
	acct.ReadRadiusAccountingConfigOptimized()
	acct.CallingStationId = clientMAC.String()

	if acct.Enabled {
		fmt.Println("Radius Accounting is enabled")
		go func(ctx context.Context) {
			// Event-Timestamp = "Sep 27 2018 10:27:04 EDT"
			// Acct-Input-Packets = 4622
			// Acct-Output-Packets = 3494
			// Acct-Input-Octets = 859981
			// Acct-Output-Octets = 1250913
			// Acct-Session-Time = 5400
			// Acct-Input-Gigawords = 0
			// Acct-Output-Gigawords = 0

		}(ctx)
	}

	// Initialize RADIUS Authentication
	var auth Authentication
	auth.ReadRadiusAuthenticationConfigOptimized()
	auth.CallingStationId = clientMAC.String()
	auth.UserName = clientMAC.String()

	if auth.Enabled {
		fmt.Println("Radius Authentication is enabled")
		go func(ctx context.Context) {

			client := radius.DefaultClient
			client.MaxPacketErrors = 2
			packet := radius.New(radius.CodeAccessRequest, []byte(auth.Secret))

			rfc2865.UserName_SetString(packet, auth.UserName)
			rfc2865.UserPassword_SetString(packet, auth.UserName)
			rfc2865.NASIPAddress_Set(packet, net.ParseIP(auth.NASIPAddress))

			rfc2865.CallingStationID_AddString(packet, auth.CallingStationId)
			rfc2865.CalledStationID_AddString(packet, auth.CalledStationId)

			rfc2865.NASPortType_Add(packet, rfc2865.NASPortType_Value_Ethernet)

			nasPort, err := strconv.Atoi(auth.NASPort)

			if err != nil {
				fmt.Printf("Error converting NASPort to integer: %s\n", err)
				nasPort = 0 // Default to 0 if conversion fails
			}

			rfc2865.NASPort_Add(packet, rfc2865.NASPort(nasPort))

			rfc2865.FramedIPAddress_Set(packet, net.ParseIP(auth.FramedIPAddress))
			rfc2865.NASIdentifier_AddString(packet, auth.NASIdentifier)

			rfc2869.NASPortID_AddString(packet, auth.NASPortId)

			rfc2865.NASIPAddress_Set(packet, net.ParseIP(auth.NASIPAddress))
			for {
				response, err := client.Exchange(ctx, packet, fmt.Sprintf("%s:%s", auth.ServerIP, "1812"))
				if err != nil {
					fmt.Printf("Error during RADIUS authentication: %s\n", err)
					time.Sleep(time.Second * 30)
					continue
				}
				switch response.Code {
				case radius.CodeAccessAccept:
					fmt.Println("Authentication successful")
				case radius.CodeAccessReject:
					fmt.Println("Authentication rejected")
				default:
					fmt.Printf("Received unexpected response code: %s\n", response.Code)
				}

				if len(response.Attributes) > 0 {
					fmt.Println("\nResponse attributes:")
					for _, attr := range response.Attributes {
						fmt.Printf("  Type: %d, Value: %x\n", attr.Type, attr.Attribute)
					}
				}
				time.Sleep(time.Second * 30) // Adjust the interval as needed
			}

		}(ctx)
	}

	// Initialize IPFIX
	var i IpFix
	i.readIpFixConfigOptimized()
	if i.Enabled {
		fmt.Println("IPFIX is enabled")
		go func(ctx context.Context) {
			traffic, err := i.readIpFixTraffic(i.Traffic)
			if err != nil {
				fmt.Printf("Error reading IPFIX traffic: %s", err)
				return
			}
			for {
				for _, t := range traffic {
					packet := i.generateIPFIXPacket(t)
					i.sendIPFIX(packet)
				}
				time.Sleep(time.Second * 10) // Adjust the interval as needed
			}
		}(ctx)
	}

	if d.Enabled {
		fmt.Println("DHCP Discovery is enabled")
		// Random xid
		xid := make([]byte, 4)
		rand.Read(xid)

		// Add options
		var options = Options{}

		// Read options from json file
		dhcpOptions := options.ReadOptions(d.Options)

		broadcast := false
		if d.DstMac.String() == "FF:FF:FF:FF:FF:FF" {
			broadcast = true
		}

		// Request IP address

		packet := RequestPacket(dhcp4.Request, d.ClientMAC, d.GiAddr, d.CiAddr, xid, broadcast, dhcpOptions)

		Client, err := NewRawClient(d.intNet)
		if err != nil {
			fmt.Printf("Error : %s", err)
			panic(err)
		}

		for {
			err = Client.sendDHCP(d.DstMac, d.SrcMac, packet, d.ServerIP, d.GiAddr)
			if err != nil {
				fmt.Printf("Error : %s", err)
				panic(err)
			}
			time.Sleep(d.Renew)
		}
	}
	select {}
}
