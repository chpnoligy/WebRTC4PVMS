package main

import (
	"fmt"
	"github.com/pion/webrtc/v3"
)

func PrintHello() {
	fmt.Println("Hello, Modules! This is mypackage speaking!")
}

type PeerConnection_Guid struct {
	Pc *webrtc.PeerConnection
}
type Peerconnection_guid_list map[string]PeerConnection_Guid 
var Pc_guid_list Peerconnection_guid_list    //lijst met GUIDs van de peerconnections, geef de GUID en de peerconnection komt



var Verbose        bool
var WS_Url         string
var Site_RefID     string
var Site_Desc      string
var Item_RefID     string
var Item_Desc      string
var Facil_RefID    string
var Facil_Desc     string 

 
 