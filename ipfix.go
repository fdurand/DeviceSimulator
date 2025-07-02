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
	// --- Template Set ---
	// IPFIX header: 16 bytes
	// Template Set header: 4 bytes (Set ID + Length)
	// Template Record: 4 bytes (Template ID + Field Count)
	// Field Specifiers: 5 fields * 4 bytes each = 20 bytes
	// Total template set: 4 + 4 + 20 = 28 bytes
	templateSet := make([]byte, 4+4+20)
	// Template Set header
	binary.BigEndian.PutUint16(templateSet[0:2], 2)                        // Set ID for Template Set is 2
	binary.BigEndian.PutUint16(templateSet[2:4], uint16(len(templateSet))) // Length
	// Template Record
	binary.BigEndian.PutUint16(templateSet[4:6], 256) // Template ID
	binary.BigEndian.PutUint16(templateSet[6:8], 5)   // Field Count

	// Field Specifiers (Information Element ID, Field Length)
	// SourceIPv4Address (8, 4)
	binary.BigEndian.PutUint16(templateSet[8:10], 8)
	binary.BigEndian.PutUint16(templateSet[10:12], 4)
	// DestinationIPv4Address (12, 4)
	binary.BigEndian.PutUint16(templateSet[12:14], 12)
	binary.BigEndian.PutUint16(templateSet[14:16], 4)
	// SourcePort (7, 2)
	binary.BigEndian.PutUint16(templateSet[16:18], 7)
	binary.BigEndian.PutUint16(templateSet[18:20], 2)
	// DestinationPort (11, 2)
	binary.BigEndian.PutUint16(templateSet[20:22], 11)
	binary.BigEndian.PutUint16(templateSet[22:24], 2)
	// PacketDeltaCount (2, 4)
	binary.BigEndian.PutUint16(templateSet[24:26], 2)
	binary.BigEndian.PutUint16(templateSet[26:28], 4)

	// --- Data Set ---
	// Data Set header: 4 bytes (Set ID + Length)
	// Data Record: 4+4+2+2+4 = 16 bytes
	dataSet := make([]byte, 4+16)
	binary.BigEndian.PutUint16(dataSet[0:2], 256)                  // Set ID matches Template ID
	binary.BigEndian.PutUint16(dataSet[2:4], uint16(len(dataSet))) // Length

	offset := 4
	copy(dataSet[offset:offset+4], traffic.SourceIP.To4())
	copy(dataSet[offset+4:offset+8], traffic.DestinationIP.To4())
	binary.BigEndian.PutUint16(dataSet[offset+8:offset+10], traffic.SourcePort)
	binary.BigEndian.PutUint16(dataSet[offset+10:offset+12], traffic.DestinationPort)
	binary.BigEndian.PutUint32(dataSet[offset+12:offset+16], traffic.Packets)

	// --- IPFIX Message Header ---
	totalLen := 16 + len(templateSet) + len(dataSet)
	packet := make([]byte, totalLen)
	binary.BigEndian.PutUint16(packet[0:2], 10) // Version
	binary.BigEndian.PutUint16(packet[2:4], uint16(totalLen))
	binary.BigEndian.PutUint32(packet[4:8], uint32(time.Now().Unix()))
	binary.BigEndian.PutUint32(packet[8:12], 1)    // Sequence Number
	binary.BigEndian.PutUint32(packet[12:16], 256) // Observation Domain ID

	// Copy template set and data set into packet
	copy(packet[16:], templateSet)
	copy(packet[16+len(templateSet):], dataSet)

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
