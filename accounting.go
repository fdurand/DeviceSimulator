package main

import (
	"fmt"
	"net"
	"os"

	"gopkg.in/ini.v1"
)

type Accounting struct {
	Enabled          bool   // Enable/Disable Radius Accounting
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

func (a *Accounting) ReadRadiusAccountingConfig(config *Config) {

	cfg, err := ini.Load(config.ConfigFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	enabled := cfg.Section("accounting").Key("enabled")

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
