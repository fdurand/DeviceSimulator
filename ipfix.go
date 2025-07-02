package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"gopkg.in/ini.v1"
)

type IpFix struct {
	Enabled         bool // Enable/Disable IPFIX
	Traffic         string
	DestinationIP   net.IP // Destination IP for IPFIX packets
	DestinationPort int    // Destination port for IPFIX packets
}

type Traffic struct {
	SourceIP        net.IP `json:"SourceIP"`
	DestinationIP   net.IP `json:"DestinationIP"`
	SourcePort      uint16 `json:"SourcePort"`
	DestinationPort uint16 `json:"DestinationPort"`
	Packets         uint32 `json:"Packets"`
	Octets          uint32 `json:"Octets"`
	Protocol        string `json:"Protocol"`
}

func (i *IpFix) readIpFixConfig(config *Config) {

	cfg, err := ini.Load(config.ConfigFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	enabled := cfg.Section("ipfix").Key("enabled")

	if enabled != nil {
		if enabled.String() == "true" || enabled.String() == "1" {
			i.Enabled = true
		} else if enabled.String() == "false" || enabled.String() == "0" {
			i.Enabled = false
		} else {
			fmt.Printf("Invalid value for enabled: %s, defaulting to false\n", enabled.String())
			i.Enabled = false
		}
	} else {
		fmt.Println("No 'enabled' key found in the configuration, defaulting to false")
		// Default to false if the key is not present
		i.Enabled = false
	}

	DestinationIP := cfg.Section("ipfix").Key("destination_ip").String()

	i.DestinationIP = net.ParseIP(DestinationIP)
	if i.DestinationIP == nil {
		fmt.Printf("Invalid IP address for destination_ip: %s\n", DestinationIP)
	}

	DestinationPort := cfg.Section("ipfix").Key("destination_port").String()
	port, err := strconv.Atoi(DestinationPort)
	if err != nil {
		fmt.Printf("Invalid port number for destination_port: %s, defaulting to 4739\n", DestinationPort)
		i.DestinationPort = 4739 // Default IPFIX port
	} else {
		i.DestinationPort = port
	}
	if i.DestinationPort < 1 || i.DestinationPort > 65535 {
		fmt.Printf("Port number out of range for destination_port: %d, defaulting to 4739\n", i.DestinationPort)
		i.DestinationPort = 4739 // Default IPFIX port
	}

	Traffic := cfg.Section("ipfix").Key("traffic").String()
	i.Traffic = Traffic
}

func (i *IpFix) readIpFixTraffic(traffic string) ([]Traffic, error) {

	IpFixTraffic := []Traffic{}

	err := json.Unmarshal([]byte(traffic), &IpFixTraffic)

	if err != nil {
		fmt.Printf("Error : %s", err)
		return nil, err
	}
	if len(IpFixTraffic) == 0 {
		fmt.Println("No traffic data found in the configuration")
		return nil, fmt.Errorf("no traffic data found")
	}
	return IpFixTraffic, nil
}

// Generates a minimal IPFIX packet (header + one data set with fake values)
func (i *IpFix) generateIPFIXPacket(traffic Traffic) []byte {
	// IPFIX header is 16 bytes
	packet := make([]byte, 16+traffic.Octets) // header + one data record (example size)

	// Header
	binary.BigEndian.PutUint16(packet[0:2], 10)                        // Version (IPFIX)
	binary.BigEndian.PutUint16(packet[2:4], uint16(len(packet)))       // Length
	binary.BigEndian.PutUint32(packet[4:8], uint32(time.Now().Unix())) // Export Time
	binary.BigEndian.PutUint32(packet[8:12], 1)                        // Sequence Number
	binary.BigEndian.PutUint32(packet[12:16], 256)                     // Observation Domain ID

	// Data Set (Template ID 256, just as an example)
	offset := 16
	binary.BigEndian.PutUint16(packet[offset:offset+2], 256)                      // Set ID (Template ID)
	binary.BigEndian.PutUint16(packet[offset+2:offset+4], uint16(traffic.Octets)) // Length of this set

	// Fake data record (24 bytes, just for demonstration)
	copy(packet[offset+4:offset+8], traffic.SourceIP.To4()) // src IP
	copy(packet[offset+8:offset+12], traffic.DestinationIP.To4())
	binary.BigEndian.PutUint16(packet[offset+12:offset+14], traffic.SourcePort)
	binary.BigEndian.PutUint16(packet[offset+14:offset+16], traffic.DestinationPort)
	binary.BigEndian.PutUint32(packet[offset+16:offset+20], traffic.Packets)
	binary.BigEndian.PutUint32(packet[offset+20:offset+24], traffic.Octets)

	// Padding (if needed)
	for i := offset + 24; i < offset+28; i++ {
		packet[i] = 0
	}

	return packet
}

func (i *IpFix) sendIPFIX(packet []byte) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", i.DestinationIP, i.DestinationPort))
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Sending IPFIX packets to", addr.String())
	_, err = conn.Write(packet)
	if err != nil {
		fmt.Printf("Error sending IPFIX packet: %s\n", err)
		return
	}
	fmt.Println("IPFIX packet sent successfully")
}
