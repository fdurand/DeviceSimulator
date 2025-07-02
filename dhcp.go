package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/krolaw/dhcp4"
	"gopkg.in/ini.v1"
)

type Interface struct {
	Enabled   bool // Enable/Disable DHCP Discovery
	Name      string
	intNet    *net.Interface
	ServerIP  net.IP           // DHCP Server IP (Destination IP)
	DstMac    net.HardwareAddr // Server Destination MAC (Ethernet Header)
	Renew     time.Duration    // Renewal time
	ClientMAC net.HardwareAddr // Device Client MAC (In the DHCP Request)
	GiAddr    net.IP           // Source Gateway IP
	SrcMac    net.HardwareAddr // Source MAC (Ethernet Header)
	CiAddr    net.IP           // Client IP (Requesting IP)
	Options   string
}

// Options Struct
type Options struct {
	Option dhcp4.OptionCode `json:"option"`
	Value  string           `json:"value"`
	Type   string           `json:"type"`
}

// Creates a request packet that a Client would send to a server.
func RequestPacket(mt dhcp4.MessageType, chAddr net.HardwareAddr, giAddr net.IP, cIAddr net.IP, xId []byte, broadcast bool, options []dhcp4.Option) dhcp4.Packet {
	p := dhcp4.NewPacket(dhcp4.BootRequest)
	p.SetCHAddr(chAddr)
	p.SetXId(xId)
	if cIAddr != nil {
		p.SetCIAddr(cIAddr)
	}
	p.SetGIAddr(giAddr)
	p.SetBroadcast(broadcast)
	p.AddOption(dhcp4.OptionDHCPMessageType, []byte{byte(mt)})
	for _, o := range options {
		p.AddOption(o.Code, o.Value)
	}
	p.PadToMinSize()
	return p
}

func (a *Options) ReadOptions(body string) []dhcp4.Option {

	DHCPOptions := []Options{}
	var dhcpOptions = []dhcp4.Option{}

	err := json.Unmarshal([]byte(body), &DHCPOptions)
	if err != nil {
		fmt.Printf("Error : %s", err)
		panic(err)
	}
	for _, option := range DHCPOptions {
		var dhcpOption = dhcp4.Option{}
		var Value interface{}
		switch option.Type {
		case "ipaddr":
			Value = net.ParseIP(option.Value)
			dhcpOption.Code = option.Option
			dhcpOption.Value = Value.(net.IP).To4()
		case "string":
			Value = option.Value
			dhcpOption.Code = option.Option
			dhcpOption.Value = []byte(Value.(string))
		case "int":
			Value = option.Value
			val, _ := strconv.Atoi(Value.(string))
			bs := make([]byte, 4)
			binary.BigEndian.PutUint32(bs, uint32(val))
			dhcpOption.Code = option.Option
			dhcpOption.Value = bs
		case "bytes":
			dhcpOption.Code = option.Option
			for _, value := range strings.Split(option.Value, ",") {
				val, _ := strconv.Atoi(value)
				dhcpOption.Value = append(dhcpOption.Value, byte(val))
			}
		}
		dhcpOptions = append(dhcpOptions, dhcpOption)
	}
	return dhcpOptions
}

func (d *Interface) readDhcpConfig(config *Config) {

	cfg, err := ini.Load(config.ConfigFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	enabled := cfg.Section("dhcp").Key("enabled")

	if enabled != nil {
		if enabled.String() == "true" || enabled.String() == "1" {
			d.Enabled = true
		} else if enabled.String() == "false" || enabled.String() == "0" {
			d.Enabled = false
		} else {
			fmt.Printf("Invalid value for enabled: %s, defaulting to false\n", enabled.String())
			d.Enabled = false
		}
	} else {
		fmt.Println("No 'enabled' key found in the configuration, defaulting to false")
		// Default to false if the key is not present
		d.Enabled = false

	}

	Server := cfg.Section("dhcp").Key("server").String()
	d.ServerIP = net.ParseIP(Server)

	GiAddr := cfg.Section("dhcp").Key("giaddr").String()
	d.GiAddr = net.ParseIP(GiAddr)

	if d.GiAddr == nil {
		d.GiAddr = net.IPv4zero
	}

	CiAddr := cfg.Section("dhcp").Key("ciaddr").String()
	d.CiAddr = net.ParseIP(CiAddr)

	if d.CiAddr == nil {
		d.CiAddr = net.IPv4zero
	}

	SrcMac := cfg.Section("dhcp").Key("srcmac").String()
	d.SrcMac, err = net.ParseMAC(SrcMac)

	if err != nil {
		d.SrcMac = d.intNet.HardwareAddr
	}

	DstMac := cfg.Section("dhcp").Key("dstmac").String()
	d.DstMac, err = net.ParseMAC(DstMac)

	if err != nil {
		d.DstMac, _ = net.ParseMAC("FF:FF:FF:FF:FF:FF")
	}

	Renew := cfg.Section("dhcp").Key("renew").String()
	timeout, err := strconv.Atoi(Renew)
	if err != nil {
		fmt.Printf("Fail to parse renew timeout:%v, %v", Renew, err)
		os.Exit(1)
	}
	d.Renew = time.Duration(timeout) * time.Second

	d.Options = cfg.Section("dhcp").Key("options").String()

}
