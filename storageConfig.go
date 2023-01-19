package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"
	"fmt"

	"WebRTC4PVMS/sip"

	"github.com/hashicorp/go-version"
	"github.com/imdario/mergo"
	"github.com/liip/sheriff"
)

// Command line flag global variables

var configFile string

//NewStreamCore do load config file
func NewStreamCore() *StorageST {
	flag.BoolVar(&Verbose, "debug", true, "set debug mode")
	
	flag.StringVar(&configFile, "config", "config.json", "config patch (/etc/server/config.json or config.json)")
	flag.Parse()
	log.Println("ReadFile Configfile :",configFile)
	 
	var cnf StorageST
	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println("os.ReadFile(configFile):",configFile ," error :", err.Error())
		os.Exit(1)
	}
	err = json.Unmarshal(data, &cnf)
	if err != nil {
		fmt.Println("json.Unmarshal(data, &tmp) error :", err.Error())
		os.Exit(1)
	}

	Verbose 			= cnf.Server.Verbose
	WS_Url  			= cnf.Server.WS_URL
	Site_RefID    	= cnf.Server.Site_RefID
	Site_Desc    	= cnf.Server.Site_Desc
	Item_RefID   	= cnf.Server.Item_RefID
	Item_Desc    	= cnf.Server.Item_Desc
	Facil_RefID   	= cnf.Server.Facil_RefID
	Facil_Desc    	= cnf.Server.Facil_Desc
	
	sip.SIP_UDP_Port  		= cnf.Server.SIP_UDP_Port
	sip.SIP_TCP_Port  		= cnf.Server.SIP_TCP_Port
	sip.SIP_WSS_Port  		= cnf.Server.SIP_WSS_Port
	sip.SIP_User 			= cnf.Server.SIP_User
	sip.SIP_Password  		= cnf.Server.SIP_Password
	sip.SIP_Server  		= cnf.Server.SIP_Server
	sip.SIP_Transport  		= cnf.Server.SIP_Transport
	sip.SIP_Realm  			= cnf.Server.SIP_Realm


	//fmt.Println("cnf.Streams: ", len(cnf.Streams))
	for id, i2 := range cnf.Streams {		
			stream := cnf.StreamDefaults
			err = mergo.Merge(&stream, i2)
			if err != nil {
				fmt.Println("mergo.Merge(&channel, i2) error :", err.Error())
				os.Exit(1)
			}
			stream.clients = make(map[string]ClientST)
			stream.ack = time.Now().Add(-255 * time.Hour)
			stream.hlsSegmentBuffer = make(map[int]SegmentOld)
			stream.signals = make(chan int, 100)

		cnf.Streams[id] = stream
		//fmt.Println("i: ", i, " :i2", i2)
		//fmt.Println("id", id," i2.name: ", i2.Name, " :i2.url", i2.URL)
	}

	//for id, i2 := range cnf.Streams {
	//	fmt.Println("id", id, " i2.name: ", i2.Name, " :i2.url", i2.URL)
	//	if i2.URL == "rtmp://171.25.232.10/12d525bc9f014e209c1280bc0d46a87e" {
	//		fmt.Println("found id", id)
	//	}
	//}
 
	return &cnf

}

//ClientDelete Delete Client
func (obj *StorageST) SaveConfig() error {

	log.Println("Saving configuration to", configFile)
	v2, err := version.NewVersion("2.0.0")
	if err != nil {
		return err
	}
	data, err := sheriff.Marshal(&sheriff.Options{
		Groups:     []string{"config"},
		ApiVersion: v2,
	}, obj)
	if err != nil {
		return err
	}
	res, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(configFile, res, 0644)
	if err != nil {
		fmt.Println("os.WriteFile(configFile, res, 0644) error :", err.Error())
				
		return err
	}
	return nil
}
