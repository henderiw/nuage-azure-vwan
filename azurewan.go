package main

import "time"

type azureVWanCfg struct {
	ConfigurationVersion struct {
		LastUpdatedTime time.Time `json:"LastUpdatedTime"`
		Version         string    `json:"Version"`
	} `json:"configurationVersion"`
	VpnSiteConfiguration struct {
		Name      string `json:"Name"`
		IPAddress string `json:"IPAddress"`
	} `json:"vpnSiteConfiguration"`
	VpnSiteConnections []struct {
		HubConfiguration struct {
			AddressSpace     string   `json:"AddressSpace"`
			Region           string   `json:"Region"`
			ConnectedSubnets []string `json:"ConnectedSubnets"`
		} `json:"hubConfiguration"`
		GatewayConfiguration struct {
			IPAddresses struct {
				Instance0 string `json:"Instance0"`
				Instance1 string `json:"Instance1"`
			} `json:"IpAddresses"`
		} `json:"gatewayConfiguration"`
		ConnectionConfiguration struct {
			IsBgpEnabled    bool   `json:"IsBgpEnabled"`
			PSK             string `json:"PSK"`
			IPsecParameters struct {
				SADataSizeInKilobytes int `json:"SADataSizeInKilobytes"`
				SALifeTimeInSeconds   int `json:"SALifeTimeInSeconds"`
			} `json:"IPsecParameters"`
		} `json:"connectionConfiguration"`
	} `json:"vpnSiteConnections"`
}
