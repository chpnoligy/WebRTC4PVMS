package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	//
	//"time"
	//adapter "mymodule/mypackage"
	//"mymodule/mypackage"
	//"mymodule/storage"

	//"mymodule/websock"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	ErrorStreamURLNotAdd         = errors.New("WebRTC Error Stream URL Not Added")
	ErrorStreamIDNotFound          = errors.New("WebRTC StreamID Not Found")
	ErrorStreamURLNotStarted     = errors.New("WebRTC Stream URL Not Started")
	ErrorAnswer     	= errors.New("WebRTC Answer Error")
	ErrorClientAdd 		= errors.New("WebRTC Error Client Not Added")
)
 

//HTTPAPIServerStreamWebRTC stream video over WebRTC zonder websocket server
func HTTPAPIServerStreamWebRTC_NOWS(c *gin.Context) {
	//public.POST("/stream/:uuid/channel/:channel/webrtc"
	 
	log.Println("HTTPAPIServerStreamWebRTC_NOWS:",c.Param("uuid") , " Client IP:", c.ClientIP()) 
	
	
	if !Storage.StreamExist(c.Param("uuid")) {
		c.IndentedJSON(500, Message{Status: 0, Payload: ErrorStreamNotFound.Error()}) 
		fmt.Println("StreamChannelNotExist") 
		return
	}

	if !RemoteAuthorization("WebRTC", c.Param("uuid"), c.Query("token"), c.ClientIP()) {
		fmt.Println("RemoteAuthorization Error")
		return
	}
	
	Storage.StreamRun(c.Param("uuid"))
	codecs, err := Storage.StreamCodecs(c.Param("uuid"))
	if err != nil {
		c.IndentedJSON(500, Message{Status: 0, Payload: err.Error()})
		fmt.Println("StreamCodecs Error")
		return
	}
	
	fmt.Println("RemoteAuthorization Error") 

	muxerWebRTC := NewMuxer(Options{ICEServers: Storage.ServerICEServers(), ICEUsername: Storage.ServerICEUsername(), 
		ICECredential: Storage.ServerICECredential(), PortMin: Storage.ServerWebRTCPortMin(), PortMax: Storage.ServerWebRTCPortMax()})
	

	answer, err := muxerWebRTC.WriteHeader(codecs, c.PostForm("data"), c.Param("uuid"))
	if err != nil {
		c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
		fmt.Println("muxerWebRTC.WriteHeader Error")
		return
	}
	_, err = c.Writer.Write([]byte(answer))
	if err != nil {
		c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
		fmt.Println("Write Error")
		return
	}
	go func() {
		cid, ch, _, err := Storage.ClientAdd(c.Param("uuid"), WEBRTC)
		if err != nil {
			c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
			fmt.Println("ClientAdd Error")
			return
		}
		defer Storage.ClientDelete(c.Param("uuid"), cid)
		var videoStart bool
		noVideo := time.NewTimer(10 * time.Second)
		for {
			select {
			case <-noVideo.C:
				//				c.IndentedJSON(500, Message{Status: 0, Payload: ErrorStreamNoVideo.Error()})
				fmt.Println("Error StreamNoVideo")
				return
			case pck := <-ch:
				if pck.IsKeyFrame {
					noVideo.Reset(10 * time.Second)
					videoStart = true
				}
				if !videoStart {
					continue
				}
				err = muxerWebRTC.WritePacket(*pck)
				if err != nil {
					fmt.Println("Error WritePacket")
					return
				}
			}
		}
	}()
}



func HTTPAPIServerStreamWebRTC(clientIP string, stream_URL string, stream_ID string, 
									resp_WSID string, 
									PeerID string, OfferSDP string, 
									wssocket_conn websocket.Conn) ( error ) {
			  
	 //resp_WSID is het wd_id van de websocket connection in de VPS, waar de offer vandaan kwam. Hier moet het answer heen
	 
	log.Println("HTTPAPIServerStreamWebRTC PID:", PeerID, " URL:", stream_URL, " streamID:", stream_ID, " respID:", resp_WSID )
	//fmt.Println("Conn", wssocket_conn.LocalAddr().String(), " url:", cam_url)
	//client IP = de VPS server
	//rtsp://koos:kooskoos1@192.168.178.111:554/cam/realmonitor?channel=1&subtype=0
	 
	if stream_ID == "" && stream_URL != "" {
		 
		//var streamTmp storage.StreamST
		fnd := Storage.URLExist(stream_URL)

		if (!fnd) {
			//URL bestaat niet, dus bijmaken
			fmt.Println("URL", stream_URL, " Not Found") 
			var streamTmp StreamST
			streamTmp.Name="stream name"
			streamTmp.URL= stream_URL 
			streamTmp.OnDemand = true
			streamTmp.Debug = false
			streamTmp.Audio = false

			err :=  Storage.SStreamAdd(PeerID, streamTmp)
			if err != nil {
				fmt.Println("storage.Storage.SStreamAdd(PeerGuid,streamTmp) err:", err.Error()) 
				return ErrorStreamURLNotAdd
			}
			stream_ID = PeerID //gebruik het peer id voor streamid, die is immers leeg
		} 
	}  
	 //stream_URL leeg
	if (!Storage.StreamExist(stream_ID)) {
	 	fmt.Println("HTTPAPIServerStreamWebRTC StreamID:",stream_ID," bestaan niet"  )
		Response_Error(PeerID, resp_WSID, "Stream Not Found:" ) 
		return ErrorStreamIDNotFound
	}
	 
	 
//stukje opstarten RTSP Stream
	//if !RemoteAuthorization("WebRTC", streamID, channelID, c.Query("token"), c.ClientIP()) {
	//if !RemoteAuthorization("WebRTC", streamID, channelID, "", clientIP ) {
	//	requestLogger.WithFields(logrus.Fields{
	//		"call": "RemoteAuthorization",
	//	}).Errorln(ErrorStreamNotFound.Error())
	//	return
	//}
	 
	Storage.StreamRun(stream_ID )
	
	fmt.Println("RTSP Stream opgestart, loopt in aparte thread, wacht op codecs")
	
	codecs, err := Storage.StreamCodecs(stream_ID)
	if err != nil {
	 	fmt.Println("Storage.StreamChannelCodecs(streamID) error :", err.Error()  )
		 Response_Error(PeerID, resp_WSID,"RTSPStream Not Started") 
		return ErrorStreamURLNotStarted
	}
	fmt.Println("RTSP StreamCodecs binnen")
 


 





	//NewMuxer in adapter.go
	muxerWebRTC := NewMuxer(Options{ICEServers: Storage.ServerICEServers(), 
		ICEUsername: Storage.ServerICEUsername(), 
		ICECredential: Storage.ServerICECredential(), 
		PortMin: Storage.ServerWebRTCPortMin(), 
		PortMax: Storage.ServerWebRTCPortMax()}) 
	


	// maak een answer 
	answer, err := muxerWebRTC.Handle_Offer(codecs, OfferSDP, PeerID, resp_WSID, stream_ID)  //hier zit een new peerconnection in
	//answer = encoded 64, heeft geen functie
	
	if answer == "" {
	    return ErrorAnswer
	}
	if err != nil {

		fmt.Println("muxerWebRTC.Handle_Offer error :", err.Error()  )
		return ErrorAnswer
	}
     
	go func() {
		
		cid, ch, _, err :=  Storage.ClientAdd(stream_ID,  WEBRTC)
		if err != nil {
			fmt.Println("HTTPAPIServerStreamWebRTC_CHP Storage.ClientAdd error"  )

			//c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
			//requestLogger.WithFields(logrus.Fields{
			//	"call": "ClientAdd",
			//}).Errorln(err.Error())
			return
		}
		defer  Storage.ClientDelete(stream_ID, cid)
		var videoStart bool
		noVideo := time.NewTimer(5 * time.Second)
		for {	 
			select {
			case <-noVideo.C:
				//c.IndentedJSON(500, Message{Status: 0, Payload: ErrorStreamNoVideo.Error()})
				//requestLogger.WithFields(logrus.Fields{
				//	"call": "ErrorStreamNoVideo",
				//}).Errorln(ErrorStreamNoVideo.Error())
				fmt.Println("ErrorStreamNoVideo")
				return
			case pck := <-ch:
				if pck.IsKeyFrame {
					noVideo.Reset(5 * time.Second)
					videoStart = true
				}
				if !videoStart {
					continue
				}
				 
				err = muxerWebRTC.WritePacket(*pck)
				if err != nil {
					fmt.Println("muxerWebRTC.WritePacket(*pck) error:", err.Error() )
					return
				}
			}
		}
	}()
    return nil
}
