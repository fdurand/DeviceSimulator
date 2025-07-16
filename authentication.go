package main

import (
	"fmt"
	"net"
	"os"

	"gopkg.in/ini.v1"
)

type Authentication struct {
	Enabled          bool   // Enable/Disable Radius Accounting
	ServerIP         net.IP // Radius Accounting server IP
	Secret           string // Radius Accounting secret
	UserName         string
	CallingStationId string
	CalledStationId  string
	NASPort          string
	NASPortType      string
	FramedIPAddress  string
	NASIdentifier    string
	NASPortId        string
	NASIPAddress     string
}

func (a *Authentication) ReadRadiusAuthenticationConfig(config *Config) {

	cfg, err := ini.Load(config.ConfigFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	enabled := cfg.Section("authentication").Key("enabled")

	if enabled != nil {
		if enabled.String() == "true" || enabled.String() == "1" {
			a.Enabled = true
		} else if enabled.String() == "false" || enabled.String() == "0" {
			a.Enabled = false
		} else {
			fmt.Printf("Invalid value for enabled: %s, defaulting to false\n", enabled.String())
			a.Enabled = false
		}
	} else {
		fmt.Println("No 'enabled' key found in the configuration, defaulting to false")
		// Default to false if the key is not present
		a.Enabled = false
	}

	server := cfg.Section("authentication").Key("server").String()
	a.ServerIP = net.ParseIP(server)

	secret := cfg.Section("authentication").Key("secret").String()
	a.Secret = secret

	username := cfg.Section("authentication").Key("User-Name").String()
	a.UserName = username

	callingstationid := cfg.Section("authentication").Key("Calling-Station-Id").String()
	a.CallingStationId = callingstationid

	calledstationid := cfg.Section("authentication").Key("Called-Station-Id").String()
	a.CalledStationId = calledstationid

	nasport := cfg.Section("authentication").Key("NAS-Port").String()
	a.NASPort = nasport

	nasporttype := cfg.Section("authentication").Key("NAS-Port-Type").String()
	a.NASPortType = nasporttype

	framedipaddress := cfg.Section("authentication").Key("Framed-IP-Address").String()
	a.FramedIPAddress = framedipaddress

	nasidentifier := cfg.Section("authentication").Key("NAS-Identifier").String()
	a.NASIdentifier = nasidentifier

	nasportid := cfg.Section("authentication").Key("NAS-Port-Id").String()
	a.NASPortId = nasportid

	nasipaddress := cfg.Section("authentication").Key("NAS-IP-Address").String()
	a.NASIPAddress = nasipaddress

}
