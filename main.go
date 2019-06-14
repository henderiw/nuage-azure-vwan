package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/nuagenetworks/go-bambou/bambou"
	"github.com/nuagenetworks/vspk-go/vspk"
)

// VSD crednetials
var vsdURL = "https://138.203.39.87:8443"
var vsdUser = "csproot"
var vsdPass = "csproot"
var vsdEnterprise = "csp"

// Usr is a user
var Usr *vspk.Me

func main() {
	rcvdAzureVWanData, readErr := ioutil.ReadFile("azure-vwan.json")
	if readErr != nil {
		log.Fatal(readErr)
	}

	fmt.Printf("File contents: %s", rcvdAzureVWanData)

	// init the empty structure
	var cfg azureVWanCfg
	// unmarshal (deserialize) the json and save the result in the struct &cfg
	err := json.Unmarshal([]byte(rcvdAzureVWanData), &cfg)
	if err != nil {
		log.Fatal(err)
	}

	vwanHubIP1 := cfg.VpnSiteConnections[0].GatewayConfiguration.IPAddresses.Instance0
	vwanHubIP2 := cfg.VpnSiteConnections[0].GatewayConfiguration.IPAddresses.Instance1
	vwanSiteName := cfg.VpnSiteConfiguration.Name
	//vwanSiteIP := cfg.VpnSiteConfiguration.IPAddress
	//vwanSiteBGPEnabled := cfg.VpnSiteConnections[0].ConnectionConfiguration.IsBgpEnabled
	vwanSitePSK := cfg.VpnSiteConnections[0].ConnectionConfiguration.PSK
	//vwanSiteSAsize := cfg.VpnSiteConnections[0].ConnectionConfiguration.IPsecParameters.SADataSizeInKilobytes
	//vwanSiteSALifeTime := cfg.VpnSiteConnections[0].ConnectionConfiguration.IPsecParameters.SALifeTimeInSeconds

	//Start session to VSD
	var s *bambou.Session
	s, Usr = vspk.NewSession(vsdUser, vsdPass, vsdEnterprise, vsdURL)
	if err := s.Start(); err != nil {
		fmt.Println("Unable to connect to Nuage VSD: " + err.Description)
		os.Exit(1)
	}

	//find enterprise
	enterpriseCfg := map[string]interface{}{
		"Name": "vspkPublicNonDpdk",
	}

	enterprise := nuagewim.nuageEnterprise(enterpriseCfg, Usr)

	//create or get PSK

	ikePSKCfg := map[string]interface{}{
		"Name":           "vspkAzure",
		"Description":    "vspkAzure",
		"UnencryptedPSK": vwanSitePSK,
	}

	ikePSK := nuagewim.nuageCreateIKEPSK(ikePSKCfg, enterprise)

	//create IKE Gateway(s)

	ikeGatewayCfg1 := map[string]interface{}{
		"Name":        "vspkAzureIKEGatewayName1",
		"Description": "vspkAzureIKEGatewayName1",
		"IKEVersion":  "V2",
		"IPAddress":   vwanHubIP1,
	}

	ikeGateway1 := nuagewim.nuageCreateIKEGateway(ikeGatewayCfg1, enterprise)

	ikeGatewayCfg2 := map[string]interface{}{
		"Name":        "vspkAzureIKEGatewayName2",
		"Description": "vspkAzureIKEGatewayName2",
		"IKEVersion":  "V2",
		"IPAddress":   vwanHubIP2,
	}

	ikeGateway2 := nuagewim.nuageCreateIKEGateway(ikeGatewayCfg2, enterprise)

	//Create IKE Encryption Profile

	ikeEncryptionProfileCfg := map[string]interface{}{
		"Name":                              "vspkAzureIKEEncryptionProfile",
		"Description":                       "vspkAzureIKEEncryptionProfile",
		"DPDMode":                           "REPLY_ONLY",
		"ISAKMPAuthenticationMode":          "PRE_SHARED_KEY",
		"ISAKMPDiffieHelmanGroupIdentifier": "GROUP_2_1024_BIT_DH",
		"ISAKMPEncryptionAlgorithm":         "AES256",
		"ISAKMPEncryptionKeyLifetime":       28800,
		"ISAKMPHashAlgorithm":               "SHA256",
		"IPsecEnablePFS":                    true,
		"IPsecEncryptionAlgorithm":          "AES256",
		"IPsecPreFragment":                  true,
		"IPsecSALifetime":                   3600,
		"IPsecAuthenticationAlgorithm":      "HMAC_SHA256",
		"IPsecSAReplayWindowSize":           "WINDOW_SIZE_64",
	}

	ikeEncryptionProfile := nuagewim.nuageCreateIKEEncryptionProfile(ikeEncryptionProfileCfg, enterprise)

	//Create IKE Gateway Profile

	ikeGatewayProfileCfg1 := map[string]interface{}{
		"Name":                             "vspkAzureIKEGatewayProfile1",
		"Description":                      "vspkAzureIKEGatewayProfile1",
		"AssociatedIKEAuthenticationID":    ikePSK.ID,
		"IKEGatewayIdentifier":             vwanHubIP1,
		"IKEGatewayIdentifierType":         "ID_IPV4_ADDR",
		"AssociatedIKEGatewayID":           ikeGateway1.ID,
		"AssociatedIKEEncryptionProfileID": ikeEncryptionProfile.ID,
	}

	ikeGatewayProfile1 := nuagewim.nuageCreateIKEGatewayProfile(ikeGatewayProfileCfg1, enterprise)

	ikeGatewayProfileCfg2 := map[string]interface{}{
		"Name":                             "vspkAzureIKEGatewayProfile2",
		"Description":                      "vspkAzureIKEGatewayProfile2",
		"AssociatedIKEAuthenticationID":    ikePSK.ID,
		"IKEGatewayIdentifier":             vwanHubIP2,
		"IKEGatewayIdentifierType":         "ID_IPV4_ADDR",
		"AssociatedIKEGatewayID":           ikeGateway2.ID,
		"AssociatedIKEEncryptionProfileID": ikeEncryptionProfile.ID,
	}

	ikeGatewayProfile2 := nuagewim.nuageCreateIKEGatewayProfile(ikeGatewayProfileCfg2, enterprise)

	nsGatewayCfg := map[string]interface{}{
		"Name":                  "vspkNsgE200Wifi1",
		"TCPMSSEnabled":         true,
		"TCPMaximumSegmentSize": 1330,
		"NetworkAcceleration":   "NONE",
	}

	nsGateway := nuagewim.nuageNSG(nsGatewayCfg, enterprise)

	nsPortCfg := map[string]interface{}{
		"Name": "port1",
	}

	nsPort := nuagewim.nuageNSGPort(nsPortCfg, nsGateway)

	nsVlanCfg := map[string]interface{}{
		"Value": "0",
	}

	nsVlan := nuagewim.nuageVlan(nsVlanCfg, nsPort)

	ikeGatewayConnCfg1 := map[string]interface{}{
		"Name":                          "vspkAzureIKEGatewayConnection1",
		"Description":                   "vspkAzureIKEGatewayConnection1",
		"NSGIdentifier":                 "Home-Wim",
		"NSGIdentifierType":             "ID_KEY_ID",
		"NSGRole":                       "INITIATOR",
		"AllowAnySubnet":                true,
		"AssociatedIKEGatewayProfileID": ikeGatewayProfile1.ID,
		"AssociatedIKEAuthenticationID": ikePSK.ID,
	}

	ikeGatewayConn1 := nuagewim.nuageIKEGatewayConnection(ikeGatewayConnCfg1, nsVlan)
	fmt.Println(ikeGatewayConn1)

	ikeGatewayConnCfg2 := map[string]interface{}{
		"Name":                          "vspkAzureIKEGatewayConnection2",
		"Description":                   "vspkAzureIKEGatewayConnection2",
		"NSGIdentifier":                 vwanSiteName,
		"NSGIdentifierType":             "ID_KEY_ID",
		"NSGRole":                       "INITIATOR",
		"AllowAnySubnet":                true,
		"AssociatedIKEGatewayProfileID": ikeGatewayProfile2.ID,
		"AssociatedIKEAuthenticationID": ikePSK.ID,
	}

	ikeGatewayConn2 := nuagewim.nuageIKEGatewayConnection(ikeGatewayConnCfg2, nsVlan)
	fmt.Println(ikeGatewayConn2)

}
