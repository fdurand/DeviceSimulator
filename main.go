package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/krolaw/dhcp4"
	"gopkg.in/ini.v1"
)

type Interface struct {
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

type Upnp struct {
	intNet     *net.Interface
	UserAgent  string
	IPAddr     net.IP
	UDPPort    int
	deviceType string
}

type Accounting struct {
	ServerIP         net.IP // Radius Accounting server IP
	Secret           string // Radius Accounting secret
	UserName         string
	AcctSessionId    string
	CallingStationId string
	CalledStationId  string
	NASPort          string
	NASPortType      string
	FramedIPAddress  string
	NASIdentifier    string
	NASPortId        string
	NASIPAddress     string
}

type GlobalConfig struct {
	intNet    *net.Interface
	ClientMAC net.HardwareAddr // Device Client MAC (In the DHCP Request)
}

// Options Struct
type Options struct {
	Option dhcp4.OptionCode `json:"option"`
	Value  string           `json:"value"`
	Type   string           `json:"type"`
}

type Config struct {
	ConfigFile *string
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
		c.ClientMAC, err = net.ParseMAC("de:ad:be:ef:de:ad")
	}
}

func (d *Interface) readDhcpConfig(config *Config) {

	cfg, err := ini.Load(config.ConfigFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
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
		d.DstMac, err = net.ParseMAC("FF:FF:FF:FF:FF:FF")
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

func (u *Upnp) readUpnpConfig(config *Config) {

	cfg, err := ini.Load(config.ConfigFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
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

func (a *Accounting) readAccountingConfig(config *Config) {

	cfg, err := ini.Load(config.ConfigFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	server := cfg.Section("accounting").Key("server").String()
	a.ServerIP = net.ParseIP(server)

	secret := cfg.Section("accounting").Key("secret").String()
	a.Secret = secret

	username := cfg.Section("accounting").Key("User-Name").String()
	a.UserName = username

	acctsessionid := cfg.Section("accounting").Key("Acct-Session-Id").String()
	a.AcctSessionId = acctsessionid

	callingstationid := cfg.Section("accounting").Key("Calling-Station-Id").String()
	a.CallingStationId = callingstationid

	calledstationid := cfg.Section("accounting").Key("Called-Station-Id").String()
	a.CalledStationId = calledstationid

	nasport := cfg.Section("accounting").Key("NAS-Port").String()
	a.NASPort = nasport

	nasporttype := cfg.Section("accounting").Key("NAS-Port-Type").String()
	a.NASPortType = nasporttype

	framedipaddress := cfg.Section("accounting").Key("Framed-IP-Address").String()
	a.FramedIPAddress = framedipaddress

	nasidentifier := cfg.Section("accounting").Key("NAS-Identifier").String()
	a.NASIdentifier = nasidentifier

	nasportid := cfg.Section("accounting").Key("NAS-Port-Id").String()
	a.NASPortId = nasportid

	nasipaddress := cfg.Section("accounting").Key("NAS-IP-Address").String()
	a.NASIPAddress = nasipaddress

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

	configuration.ConfigFile = flag.String("file", "/usr/local/etc/config.ini", "Configuration File Path")

	var c GlobalConfig

	c.ReadGlobalConfig(configuration)

	var d Interface
	d.readDhcpConfig(configuration)

	d.intNet = c.intNet
	d.ClientMAC = c.ClientMAC

	var u Upnp
	u.readUpnpConfig(configuration)

	u.intNet = c.intNet

	go func() {
		for {
			u.discover(time.Second * 10)

			time.Sleep(time.Second * 30)
		}
	}()

	var a Accounting
	a.readAccountingConfig(configuration)
	a.CallingStationId = d.ClientMAC.String()

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

	// body, err := os.ReadFile("/usr/local/etc/Options.json")
	// if err != nil {
	// 	fmt.Printf("Error : %s", err)
	// 	panic(err)
	// }
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
