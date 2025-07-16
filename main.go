package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/krolaw/dhcp4"
	"gopkg.in/ini.v1"
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
	cfg, err := ini.Load(config.ConfigFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	Interface := cfg.Section("general").Key("interface").String()
	c.intNet, err = net.InterfaceByName(Interface)
	if err != nil {
		fmt.Printf("Fail to find network interface:%v, %v", Interface, err)
		os.Exit(1)
	}

	ClientMac := cfg.Section("general").Key("clientmac").String()
	c.ClientMAC, err = net.ParseMAC(ClientMac)

	if err != nil {
		c.ClientMAC, _ = net.ParseMAC("de:ad:be:ef:de:ad")
	}
}

func main() {

	ctx := context.Background()
	go func(ctx context.Context) {
		// Systemd
		daemon.SdNotify(false, "READY=1")
		interval, err := daemon.SdWatchdogEnabled(false)
		if err != nil || interval == 0 {
			return
		}
		for {
			daemon.SdNotify(false, "WATCHDOG=1")
			time.Sleep(interval / 3)
		}
	}(ctx)

	configuration := &Config{}
	configFile := flag.String("file", "/usr/local/etc/config.ini", "Configuration File Path")

	flag.Parse()

	configuration.ConfigFile = *configFile

	var c GlobalConfig

	c.ReadGlobalConfig(configuration)

	var d Interface
	d.intNet = c.intNet
	d.ClientMAC = c.ClientMAC
	d.readDhcpConfig(configuration)

	var u Upnp
	u.readUpnpConfig(configuration)

	u.intNet = c.intNet

	if u.Enabled {
		fmt.Println("UPnP Discovery is enabled")
		go func() {
			for {
				u.discover(time.Second * 10)

				time.Sleep(time.Second * 30)
			}
		}()
	}

	var acct Accounting
	acct.readRadiusAccountingConfig(configuration)
	acct.CallingStationId = d.ClientMAC.String()

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

	var auth Authentication
	auth.readRadiusAuthenticationConfig(configuration)
	auth.CallingStationId = d.ClientMAC.String()
	auth.UserName = d.ClientMAC.String()

	if auth.Enabled {
		fmt.Println("Radius Authentication is enabled")
		go func(ctx context.Context) {
			client := radius.Client{
				Retry: 3, // Number of retry attempts
			}
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

			response, err := client.Exchange(ctx, packet, fmt.Sprintf("%s:%s", auth.ServerIP, "1812"))

			if err != nil {
				fmt.Printf("Error during RADIUS authentication: %s\n", err)
				return
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

			time.Sleep(time.Second * 10)

		}(ctx)
	}

	var i IpFix
	i.readIpFixConfig(configuration)
	if i.Enabled {
		fmt.Println("IPFIX is enabled")
		go func() {
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
		}()
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
