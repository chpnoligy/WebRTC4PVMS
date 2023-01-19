package main

import (
	//"encoding/base64"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	//"net/url"
	"sync"
	"time"

	//"net"

	//http "mymodule/http"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

//var message string
//var addr = flag.String("5.157.80.190", ":8080", "http service address")
// WebSocketClient return websocket client connection

type WebSocketClient struct {
	configStr string
	wsconn    *websocket.Conn
	 
}

var running bool
 
var ws_clnt WebSocketClient 

var ws_client *WebsocketClient_strct

func NewWebSocketClient(url string) error {
      
    ws_clnt = WebSocketClient{}

	ws_clnt.configStr = url
	 
    running = true
	go Listen()   //om wat te ontvangen, doet ook de connect
	return nil
    //return  nil
}

func Connect() *websocket.Conn {

	if ws_clnt.wsconn != nil {

		return ws_clnt.wsconn
	}

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for ; ; <-ticker.C {

		//u := url.URL{Scheme: "wss", Host: "wss://5.157.80.190:8083"}
		dialer := *websocket.DefaultDialer
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		ws, _, err := dialer.Dial(ws_clnt.configStr, nil) 
		//ws, _, err := websocket.DefaultDialer.Dial(ws_clnt.configStr, nil) 
		//dialer := *websocket.DefaultDialer
		//dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		if err != nil {
			log.Println("Websocket Client Cannot Connect to", ws_clnt.configStr) 
		    continue
		}
		log.Println("Websocket Client Connected to", ws_clnt.configStr) 
		ws_clnt.wsconn = ws
		 
		ws_client = &WebsocketClient_strct{
			websocket: ws,
		}
        type Message struct {
            Frame_Type string
            Site_RefID string
            Site_Desc string
            Faciliy_RefID string
            Faciliy_Desc string
            Item string
            Item_RefID string
            Item_Desc string
            Item_RefID_2 string
            Item_Desc_2 string
            Event_Code string
            Event_Desc string
            Info string
            TijdDatum string
        }
        if (Verbose) {
			//log.Println("verbose")
		}
        m := Message{"SYSTEEM",  Site_RefID, Site_Desc, Facil_RefID, Facil_Desc,"VIDEOBRIDGE", Item_RefID, Item_Desc,"","","","CONNECTED","0","2022-12-14 12:34:56"}
        Write_String(m)

		return ws_clnt.wsconn
	}
}


 
type RXWS_Message_Frame struct {
	Frame_Type string 			`json:"Frame_Type,omitempty"`
	Frame_From_WSID string 		`json:"Frame_From_WSID,omitempty"` 
	Frame_Payload string 		`json:"Frame_Payload,omitempty"`  
}

type RXWS_Message_Type struct {
	MessageType string 		`json:"messageType,omitempty"` 
}
type RXWS_Message_Type_Offer struct {
	MessageType string 		`json:"messageType,omitempty"`
	PeerDescription string 	`json:"peerDescription,omitempty"`
	IPClient string 		`json:"ipClient,omitempty"`
	StreamID string 		`json:"streamID,omitempty"`
	RespWSID string 		`json:"respWSID,omitempty"`    //ws.ws_id waar answer en icecandite heen moet
	PeerID string 		    `json:"peerID,omitempty"`
	URL string 		        `json:"url,omitempty"`
}

type RXWS_Message_Type_Candidate struct {
	MessageType string 		`json:"messageType,omitempty"`
	PeerID string 		    `json:"peerID,omitempty"`
	Candidate string 		`json:"iceCandidate,omitempty"`
}
type RXWS_Message_Type_Bye struct {
	MessageType string 		`json:"messageType,omitempty"`
	PeerID string 		    `json:"peerID,omitempty"`
}

 


func Listen() {

    ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		for {
			wsc := Connect()   //wsc = ws_conn.wsconn
			if wsc == nil {
				return  
			}

			//_, bytMsg, err := wsc.ReadMessage()
			_, bytMsg, err := wsc.ReadMessage()
			if err != nil {
                if !running {
                    return 
                }
				wslog("listen", err, "Cannot read websocket message")
				Stop()
				break
			}
			
			//l := len(bytMsg)
			//fmt.Println("receive l:", l)		   
			
			var rx_frame RXWS_Message_Frame   

			err = json.Unmarshal(bytMsg, &rx_frame)
			if err != nil {
				fmt.Println("json.Unmarshal Frame_Type error:", err.Error())
				break  
			}
			//fmt.Println("receive rx_frame.Frame_Type:", rx_frame.Frame_Type ) 
			switch(rx_frame.Frame_Type) {
				//case "SYSTEM":
				//	fmt.Println("receive rx_frame.Frame_Payload:", rx_frame.Frame_Payload ) 
				
				case "WEBRTC_REQ":
					var message RXWS_Message_Type
					err = json.Unmarshal([]byte(rx_frame.Frame_Payload), &message)
					if err != nil {
						fmt.Println("json.Unmarshal MsgType error:", err.Error())
						break  
					}
					//fmt.Println("Incomming message MessageType :", message.MessageType) 
					switch (message.MessageType) { 
					case "offer":
						 
						//	fmt.Println("Incomming OFFER !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
						 
						var msg_offer RXWS_Message_Type_Offer
						err = json.Unmarshal([]byte(rx_frame.Frame_Payload), &msg_offer)
						if err != nil {
							fmt.Println("json.Unmarshal RXWS_Message_Type_Offer error:", err.Error())
							break  
						}
						//fmt.Println("IPClient:", msg_offer.IPClient) //niet helemaal duidelijk waarvoor
						//fmt.Println("ChannelID:", msg_offer.ChannelID)
						//fmt.Println("StreamID:", msg_offer.StreamID)			 
						//fmt.Println("PeerGUID:", msg_offer.PeerID)			 
					    //fmt.Println("Frame_From_WSID:", rx_frame.Frame_From_WSID)	
						err := HTTPAPIServerStreamWebRTC(msg_offer.IPClient, msg_offer.URL, msg_offer.StreamID, 
								rx_frame.Frame_From_WSID, 
								msg_offer.PeerID,
								msg_offer.PeerDescription, *wsc) 
						if err != nil {
							fmt.Println("http.HTTPAPIServerStreamWebRTC error:", err.Error())

							break  
						}
					case "icecandidate":
						 
						//	fmt.Println("Incomming CANDIDATE !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
						 
						var msg_candidate RXWS_Message_Type_Candidate
						err = json.Unmarshal([]byte(rx_frame.Frame_Payload), &msg_candidate)
						if err != nil {
							fmt.Println("json.Unmarshal RXWS_Message_Type_Candidate error:", err.Error())
							break  
						}
						//fmt.Println("PeerGUID:", msg_candidate.PeerID)
						
						var candidate webrtc.ICECandidateInit
						err := json.Unmarshal([]byte(msg_candidate.Candidate), &candidate)
						if err != nil {
							fmt.Println("json.Unmarshal([]byte(msg_candidate.Candidate), &candidate) err:", err.Error()) 
							break
						}
						
						pc := Pc_guid_list[msg_candidate.PeerID].Pc
						if(pc == nil){
							fmt.Println("PeerID:", msg_candidate.PeerID , " not Found !!!")
							break
						}
						err = pc.AddICECandidate(candidate)
						if(err != nil){
							fmt.Println("peerConnection.AddICECandidate error:", err.Error()) 
						}
						
					case "bye":
						 
						//	fmt.Println("Incomming BYE!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
						 
						var msg_bye RXWS_Message_Type_Bye
						err = json.Unmarshal([]byte(rx_frame.Frame_Payload), &msg_bye)
						if err != nil {
							fmt.Println("json.Unmarshal RXWS_Message_Type_Bye error:", err.Error())
							break  
						}
						
						pc :=  Pc_guid_list[msg_bye.PeerID].Pc
						if(pc == nil){
							fmt.Println("PeerID:", msg_bye.PeerID , " not Found !!!")
							break 
						}
						pc.Close()
						delete( Pc_guid_list,msg_bye.PeerID )
		
					default:
						fmt.Println("Incomming Onbekend Type !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!", message.MessageType)	
					}
				
				default:
					fmt.Println("Frame Type !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!", rx_frame.Frame_Type)	
				  
		 
			} // swich frametype

		}
	}
}
 

type WebsocketClient_strct struct {
	websocket   *websocket.Conn
	mutex            sync.Mutex
}

func (websocketClient *WebsocketClient_strct) sendMessage(msg []byte) error {
	websocketClient.mutex.Lock()
	defer websocketClient.mutex.Unlock()

	return websocketClient.websocket.WriteMessage(websocket.TextMessage, msg)
}

// Write data to the websocket server or drop it after 50ms
func  Write_String(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
		 
	err = ws_client.sendMessage(data)
	if err != nil {
		 	wslog("Write", nil, "WebSocket Write Error")
		 	return err
		  }
	return nil
}

func  Write_Array(data []byte) error {
	
	// data := string(payload )
	//adata, err := json.Marshal(payload)
	//if err != nil {
	//	return err
	//}
		 
	err := ws_client.sendMessage(data)
	if err != nil {
		 	wslog("Write", nil, "WebSocket Write Error")
		 	return err
		  }
	return nil
}

 


// Close will send close message and shutdown websocket connection
func  Stop() {

    running = false
	if ws_clnt.wsconn != nil {
		ws_clnt.wsconn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		ws_clnt.wsconn.Close()
		ws_clnt.wsconn = nil
	}
}

func wslog(f string, err error, msg string) {
	if err != nil {
		fmt.Printf("Error in func: %s, err: %v, msg: %s\n", f, err, msg)
	} else {
		fmt.Printf("Func: %s, %s\n", f, msg)
	}
}
 
func WSClient_Connect(url string) {

	log.Println("WSClient_Connect :", url)
    Pc_guid_list = make( Peerconnection_guid_list)
    
	err := NewWebSocketClient(url)
    if err != nil {
		panic(err)
	}

 
}

func WSCLient_Disconnect() {


    log.Println("WSCLient_Disconnect")
   
	Stop()
	 
}

