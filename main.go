package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/krolaw/dhcp4"
	"gopkg.in/ini.v1"
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

	go func() {
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
	}()

	configuration := &Config{}

	configuration.ConfigFile = *flag.String("file", "/usr/local/etc/config.ini", "Configuration File Path")

	var c GlobalConfig

	c.ReadGlobalConfig(configuration)

	var d Interface
	d.readDhcpConfig(configuration)

	d.intNet = c.intNet
	d.ClientMAC = c.ClientMAC

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

	var a Accounting
	a.readAccountingConfig(configuration)
	a.CallingStationId = d.ClientMAC.String()

	if a.Enabled {
		fmt.Println("Radius Accounting is enabled")
		go func() {
			// Event-Timestamp = "Sep 27 2018 10:27:04 EDT"
			// Acct-Input-Packets = 4622
			// Acct-Output-Packets = 3494
			// Acct-Input-Octets = 859981
			// Acct-Output-Octets = 1250913
			// Acct-Session-Time = 5400
			// Acct-Input-Gigawords = 0
			// Acct-Output-Gigawords = 0

		}()
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
