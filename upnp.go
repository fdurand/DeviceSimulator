package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

type Upnp struct {
	Enabled    bool // Enable/Disable UPnP Discovery
	intNet     *net.Interface
	UserAgent  string
	IPAddr     net.IP
	UDPPort    int
	deviceType string
}

func (u *Upnp) readUpnpConfig(config *Config) {

	cfg, err := ini.Load(config.ConfigFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	enabled := cfg.Section("upnp").Key("enabled")

	if enabled != nil {
		if enabled.String() == "true" || enabled.String() == "1" {
			u.Enabled = true
		} else if enabled.String() == "false" || enabled.String() == "0" {
			u.Enabled = false
		} else {
			fmt.Printf("Invalid value for enabled: %s, defaulting to false\n", enabled.String())
			u.Enabled = false
		}
	} else {
		fmt.Println("No 'enabled' key found in the configuration, defaulting to false")
		// Default to false if the key is not present
		u.Enabled = false
	}

	UserAgent := cfg.Section("upnp").Key("useragent").String()
	u.UserAgent = UserAgent

	ipaddr := cfg.Section("upnp").Key("ipaddr").String()
	u.IPAddr = net.ParseIP(ipaddr)

	udpport := cfg.Section("upnp").Key("udpport").String()
	UdpPort, err := strconv.Atoi(udpport)
	if err != nil {
		fmt.Printf("Fail to parse udp port:%v, %v", udpport, err)
		os.Exit(1)
	}
	u.UDPPort = UdpPort

	deviceType := cfg.Section("upnp").Key("devicetype").String()
	u.deviceType = deviceType
}

func (u *Upnp) discover(timeout time.Duration) {
	ssdp := &net.UDPAddr{IP: u.IPAddr, Port: u.UDPPort}

	tpl := `M-SEARCH * HTTP/1.1
HOST: %s
ST: %s
MAN: "ssdp:discover"
MX: %d
USER-AGENT: %s

`
	searchStr := fmt.Sprintf(tpl, ssdp, u.deviceType, timeout/time.Second, u.UserAgent)

	search := []byte(strings.Replace(searchStr, "\n", "\r\n", -1))

	fmt.Println("Starting discovery of device type", u.deviceType, "on", u.intNet.Name)

	socket, err := net.ListenMulticastUDP("udp4", u.intNet, &net.UDPAddr{IP: ssdp.IP})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer socket.Close() // Make sure our socket gets closed

	err = socket.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Sending search request for device type", u.deviceType, "on", u.intNet.Name)

	_, err = socket.WriteTo(search, ssdp)
	if err != nil {
		if e, ok := err.(net.Error); !ok || !e.Timeout() {
			fmt.Println(err)
		}
		return
	}

	fmt.Println("Listening for UPnP response for device type", u.deviceType, "on", u.intNet.Name)

	// Listen for responses until a timeout is reached
	for {
		resp := make([]byte, 65536)
		_, _, err := socket.ReadFrom(resp)
		if err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				fmt.Println("UPnP read:", err) //legitimate error, not a timeout.
			}
			break
		}
	}
	fmt.Println("Discovery for device type", u.deviceType, "on", u.intNet.Name, "finished.")
}
