package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"
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

// Helper function to parse MAC address from string format (e.g., "fa:cb:aa:c5:68:fa")
func parseMACAddress(macStr string) ([]byte, error) {
	if macStr == "" {
		return nil, fmt.Errorf("empty MAC address")
	}

	mac, err := net.ParseMAC(macStr)
	if err != nil {
		return nil, fmt.Errorf("invalid MAC address format: %s", macStr)
	}

	return mac, nil
}

// Helper function to get device information from configManager
func getDeviceInfo() (net.IP, []byte, error) {
	// Get device IP address (ciaddr from dhcp section)
	deviceIP := configManager.GetIP("dhcp", "ciaddr", nil)
	if deviceIP == nil {
		return nil, nil, fmt.Errorf("device IP address (ciaddr) not found in config")
	}

	// Get device MAC address (clientmac from general section)
	clientmacStr := configManager.GetString("general", "clientmac", "")
	if clientmacStr == "" {
		return nil, nil, fmt.Errorf("device MAC address (clientmac) not found in config")
	}

	deviceMAC, err := parseMACAddress(clientmacStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse device MAC: %v", err)
	}

	return deviceIP, deviceMAC, nil
}

// Helper function to determine MAC addresses based on IP matching
func determineMACAddresses(traffic Traffic, deviceIP net.IP, deviceMAC []byte) (srcMAC, dstMAC []byte) {
	// Default MAC addresses for non-device endpoints
	defaultSrcMAC := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	defaultDstMAC := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}

	// Check if source IP matches device IP
	if traffic.SourceIP.Equal(deviceIP) {
		srcMAC = deviceMAC
	} else {
		srcMAC = defaultSrcMAC
	}

	// Check if destination IP matches device IP
	if traffic.DestinationIP.Equal(deviceIP) {
		dstMAC = deviceMAC
	} else {
		dstMAC = defaultDstMAC
	}

	return srcMAC, dstMAC
}

// Generates a comprehensive IPFIX packet with 23 fields matching the specification
func (i *IpFix) generateIPFIXPacket(traffic Traffic) []byte {
	// --- Template Set ---
	// IPFIX header: 16 bytes
	// Template Set header: 4 bytes (Set ID + Length)
	// Template Record: 4 bytes (Template ID + Field Count)
	// Field Specifiers: 23 fields (19 standard + 4 enterprise-specific)
	// Standard fields: 19 * 4 bytes = 76 bytes
	// Enterprise fields: 4 * 8 bytes (4 bytes type + 4 bytes PEN) = 32 bytes
	// Total field specifiers: 76 + 32 = 108 bytes
	// Total template set: 4 + 4 + 108 = 116 bytes
	templateSet := make([]byte, 116)

	// Template Set header
	binary.BigEndian.PutUint16(templateSet[0:2], 2)   // Set ID for Template Set is 2
	binary.BigEndian.PutUint16(templateSet[2:4], 116) // Length

	// Template Record
	binary.BigEndian.PutUint16(templateSet[4:6], 257) // Template ID = 257
	binary.BigEndian.PutUint16(templateSet[6:8], 23)  // Field Count = 23

	offset := 8

	// Field 1: SRC_MAC (56, 6)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 56)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 6)
	offset += 4

	// Field 2: SOURCE_MAC (81, 6)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 81)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 6)
	offset += 4

	// Field 3: DESTINATION_MAC (80, 6)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 80)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 6)
	offset += 4

	// Field 4: DST_MAC (57, 6)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 57)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 6)
	offset += 4

	// Field 5: IP_SRC_ADDR (8, 4)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 8)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 4)
	offset += 4

	// Field 6: IP_DST_ADDR (12, 4)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 12)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 4)
	offset += 4

	// Field 7: L4_SRC_PORT (7, 2)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 7)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 2)
	offset += 4

	// Field 8: L4_DST_PORT (11, 2)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 11)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 2)
	offset += 4

	// Field 9: TCP_FLAGS (6, 1)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 6)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 1)
	offset += 4

	// Field 10: DIRECTION (61, 1)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 61)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 1)
	offset += 4

	// Field 11: PKTS (2, 8)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 2)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 8)
	offset += 4

	// Field 12: flowStartMilliseconds (152, 8)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 152)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 8)
	offset += 4

	// Field 13: flowEndMilliseconds (153, 8)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 153)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 8)
	offset += 4

	// Field 14: biflowDirection (239, 1)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 239)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 1)
	offset += 4

	// Field 15: newConnectionDeltaCount (278, 4)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 278)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 4)
	offset += 4

	// Field 16: Connection client IPv4 address (Enterprise field: type 12236, PEN 9, length 4)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 0x8000|12236) // Set enterprise bit
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 4)
	binary.BigEndian.PutUint32(templateSet[offset+4:offset+8], 9) // PEN: ciscoSystems
	offset += 8

	// Field 17: Connection client transport port (Enterprise field: type 12240, PEN 9, length 2)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 0x8000|12240) // Set enterprise bit
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 2)
	binary.BigEndian.PutUint32(templateSet[offset+4:offset+8], 9) // PEN: ciscoSystems
	offset += 8

	// Field 18: Connection server IPv4 address (Enterprise field: type 12237, PEN 9, length 4)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 0x8000|12237) // Set enterprise bit
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 4)
	binary.BigEndian.PutUint32(templateSet[offset+4:offset+8], 9) // PEN: ciscoSystems
	offset += 8

	// Field 19: Connection server transport port (Enterprise field: type 12241, PEN 9, length 2)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 0x8000|12241) // Set enterprise bit
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 2)
	binary.BigEndian.PutUint32(templateSet[offset+4:offset+8], 9) // PEN: ciscoSystems
	offset += 8

	// Field 20: observationPointId (138, 8)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 138)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 8)
	offset += 4

	// Field 21: IP_PROTOCOL_VERSION (60, 1)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 60)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 1)
	offset += 4

	// Field 22: PROTOCOL (4, 1)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 4)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 1)
	offset += 4

	// Field 23: APPLICATION_ID (95, 4)
	binary.BigEndian.PutUint16(templateSet[offset:offset+2], 95)
	binary.BigEndian.PutUint16(templateSet[offset+2:offset+4], 4)

	// --- Data Set ---
	// Calculate data record size based on all 23 fields:
	// MAC addresses: 6+6+6+6 = 24 bytes
	// IP addresses: 4+4 = 8 bytes
	// Ports: 2+2 = 4 bytes
	// TCP flags: 1 byte
	// Direction: 1 byte
	// Packets: 8 bytes
	// Flow times: 8+8 = 16 bytes
	// Biflow direction: 1 byte
	// New connection count: 4 bytes
	// Connection client IP: 4 bytes
	// Connection client port: 2 bytes
	// Connection server IP: 4 bytes
	// Connection server port: 2 bytes
	// Observation point: 8 bytes
	// IP protocol version: 1 byte
	// Protocol: 1 byte
	// Application ID: 4 bytes
	// Total: 24+8+4+1+1+8+16+1+4+4+2+4+2+8+1+1+4 = 93 bytes
	dataRecordSize := 93
	dataSet := make([]byte, 4+dataRecordSize)
	binary.BigEndian.PutUint16(dataSet[0:2], 257)                  // Set ID matches Template ID (257)
	binary.BigEndian.PutUint16(dataSet[2:4], uint16(len(dataSet))) // Length

	dataOffset := 4

	// Get device information from configManager
	deviceIP, deviceMAC, err := getDeviceInfo()
	if err != nil {
		fmt.Printf("Warning: Failed to get device info from config: %v. Using default MAC addresses.\n", err)
		// Fallback to default MAC addresses
		deviceIP = net.ParseIP("0.0.0.0") // This will never match, so defaults will be used
		deviceMAC = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	}

	// Determine MAC addresses based on IP matching
	srcMAC, dstMAC := determineMACAddresses(traffic, deviceIP, deviceMAC)

	// Field 1: SRC_MAC (6 bytes) - Use determined source MAC
	copy(dataSet[dataOffset:dataOffset+6], srcMAC)
	dataOffset += 6

	// Field 2: SOURCE_MAC (6 bytes) - Same as SRC_MAC
	copy(dataSet[dataOffset:dataOffset+6], srcMAC)
	dataOffset += 6

	// Field 3: DESTINATION_MAC (6 bytes) - Use determined destination MAC
	copy(dataSet[dataOffset:dataOffset+6], dstMAC)
	dataOffset += 6

	// Field 4: DST_MAC (6 bytes) - Same as DESTINATION_MAC
	copy(dataSet[dataOffset:dataOffset+6], dstMAC)
	dataOffset += 6 // Field 5: IP_SRC_ADDR (4 bytes)
	copy(dataSet[dataOffset:dataOffset+4], traffic.SourceIP.To4())
	dataOffset += 4

	// Field 6: IP_DST_ADDR (4 bytes)
	copy(dataSet[dataOffset:dataOffset+4], traffic.DestinationIP.To4())
	dataOffset += 4

	// Field 7: L4_SRC_PORT (2 bytes)
	binary.BigEndian.PutUint16(dataSet[dataOffset:dataOffset+2], traffic.SourcePort)
	dataOffset += 2

	// Field 8: L4_DST_PORT (2 bytes)
	binary.BigEndian.PutUint16(dataSet[dataOffset:dataOffset+2], traffic.DestinationPort)
	dataOffset += 2

	// Field 9: TCP_FLAGS (1 byte) - SYN+ACK flags
	dataSet[dataOffset] = 0x18
	dataOffset += 1

	// Field 10: DIRECTION (1 byte) - 0=ingress, 1=egress
	dataSet[dataOffset] = 0x01
	dataOffset += 1

	// Field 11: PKTS (8 bytes)
	binary.BigEndian.PutUint64(dataSet[dataOffset:dataOffset+8], uint64(traffic.Packets))
	dataOffset += 8

	// Field 12: flowStartMilliseconds (8 bytes)
	flowStart := uint64(time.Now().UnixMilli())
	binary.BigEndian.PutUint64(dataSet[dataOffset:dataOffset+8], flowStart)
	dataOffset += 8

	// Field 13: flowEndMilliseconds (8 bytes)
	flowEnd := uint64(time.Now().UnixMilli() + 1000) // 1 second later
	binary.BigEndian.PutUint64(dataSet[dataOffset:dataOffset+8], flowEnd)
	dataOffset += 8

	// Field 14: biflowDirection (1 byte)
	dataSet[dataOffset] = 0x01
	dataOffset += 1

	// Field 15: newConnectionDeltaCount (4 bytes)
	binary.BigEndian.PutUint32(dataSet[dataOffset:dataOffset+4], 1)
	dataOffset += 4

	// Field 16: Connection client IPv4 address (4 bytes)
	copy(dataSet[dataOffset:dataOffset+4], traffic.SourceIP.To4())
	dataOffset += 4

	// Field 17: Connection client transport port (2 bytes)
	binary.BigEndian.PutUint16(dataSet[dataOffset:dataOffset+2], traffic.SourcePort)
	dataOffset += 2

	// Field 18: Connection server IPv4 address (4 bytes)
	copy(dataSet[dataOffset:dataOffset+4], traffic.DestinationIP.To4())
	dataOffset += 4

	// Field 19: Connection server transport port (2 bytes)
	binary.BigEndian.PutUint16(dataSet[dataOffset:dataOffset+2], traffic.DestinationPort)
	dataOffset += 2

	// Field 20: observationPointId (8 bytes)
	binary.BigEndian.PutUint64(dataSet[dataOffset:dataOffset+8], 1)
	dataOffset += 8

	// Field 21: IP_PROTOCOL_VERSION (1 byte) - IPv4 = 4
	dataSet[dataOffset] = 4
	dataOffset += 1

	// Field 22: PROTOCOL (1 byte) - TCP = 6, UDP = 17
	var protocolNum uint8 = 6 // Default to TCP
	if traffic.Protocol == "UDP" {
		protocolNum = 17
	}
	dataSet[dataOffset] = protocolNum
	dataOffset += 1

	// Field 23: APPLICATION_ID (4 bytes) - Generic application
	binary.BigEndian.PutUint32(dataSet[dataOffset:dataOffset+4], 80) // HTTP application

	// --- IPFIX Message Header ---
	totalLen := 16 + len(templateSet) + len(dataSet)
	packet := make([]byte, totalLen)
	binary.BigEndian.PutUint16(packet[0:2], 10) // Version
	binary.BigEndian.PutUint16(packet[2:4], uint16(totalLen))
	binary.BigEndian.PutUint32(packet[4:8], uint32(time.Now().Unix()))
	binary.BigEndian.PutUint32(packet[8:12], 1)    // Sequence Number
	binary.BigEndian.PutUint32(packet[12:16], 257) // Observation Domain ID matches Template ID

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
