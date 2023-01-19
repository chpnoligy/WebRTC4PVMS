package main

import (
	"encoding/json"
	"fmt"
	"log"
	"WebRTC4PVMS/onvif"
	"regexp"
	//"github.com/pion/webrtc/v3"
)
type RxData_Type struct {
    Item string           		`json:"item,omitempty"`
    Command string          	`json:"command,omitempty"`
	Params string               `json:"params,omitempty"`
}

func Extract_IP_Afdres(ipstr string) string {
	
	re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	return re.FindString(ipstr)

}


func Handle_Incomming_Data(rxdata string, strm *StreamST) {

	var rxd RxData_Type
	err := json.Unmarshal([]byte(rxdata) , &rxd)
	if err != nil {
		fmt.Println("json.Unmarshal Handle_Incomming_Data error:", err.Error())
		return  
	}
	ipa := Extract_IP_Afdres(strm.URL)
	log.Println("Handle_Incomming_Data:", rxd , "to IP:", ipa)
	d := onvif.Device{
		XAddr:    "http://" + ipa + "/onvif/device_service",
		User:     strm.Username,
		Password: strm.Password,
	}
	profiles, err := d.GetProfiles()
	if err != nil {
		fmt.Println("error d.GetProfiles():", err.Error())
		return  
	}
	 

	switch( rxd.Command) {
	case "gotopreset":
		//fmt.Println("gotopreset:", rxd.Params)
		err = d.GotoPreset(profiles[0].Token,rxd.Params)
		if err != nil {
			fmt.Println("error d.GotoPreset:", err.Error())
			return  
		}
	case "stop":
		err = d.Stop(profiles[0].Token)
		if err != nil {
			fmt.Println("error d.Stop:", err.Error())
			return  
		}
	case "relativemove":
		 
		var ptzv  onvif.PTZVector 
		ptzv.PanTilt.X = 1  	//rechts
		ptzv.PanTilt.Y = 1 		//up
		ptzv.Zoom.X =  1 		//zoom in

		err = d.RelativeMove(profiles[0].Token,ptzv)
		if err != nil {
			fmt.Println("error d.Stop:", err.Error())
			return  
		}
	}

}
