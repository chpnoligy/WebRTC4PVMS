package main

import (
	//"WebRTC4PVMS/sip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	 

	"fmt"
	"log"
	"time"

	"github.com/deepch/vdk/av"

	"github.com/deepch/vdk/codec/h264parser"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	 
)

var (
	ErrorNotFound          = errors.New("WebRTC Stream Not Found")
	ErrorCodecNotSupported = errors.New("WebRTC Codec Not Supported")
	ErrorClientOffline     = errors.New("WebRTC Client Offline")
	ErrorNotTrackAvailable = errors.New("WebRTC Not Track Available")
	ErrorIgnoreAudioTrack  = errors.New("WebRTC Ignore Audio Track codec not supported WebRTC support only PCM_ALAW, PCM_MULAW Or Opus")
)

 

type Muxer struct {
	streams   map[int8]*Stream
	status    webrtc.ICEConnectionState
	stop      bool
	pc        *webrtc.PeerConnection
	ClientACK *time.Timer
	StreamACK *time.Timer
	Options   Options
	pc_guid   string
}
type Stream struct {
	codec av.CodecData
	track *webrtc.TrackLocalStaticSample
}
type Options struct {
	// ICEServers is a required array of ICE server URLs to connect to (e.g., STUN or TURN server URLs)
	ICEServers []string
	// ICEUsername is an optional username for authenticating with the given ICEServers
	ICEUsername string
	// ICECredential is an optional credential (i.e., password) for authenticating with the given ICEServers
	ICECredential string
	// ICECandidates sets a list of external IP addresses of 1:1
	ICECandidates []string
	// PortMin is an optional minimum (inclusive) ephemeral UDP port range for the ICEServers connections
	PortMin uint16
	// PortMin is an optional maximum (inclusive) ephemeral UDP port range for the ICEServers connections
	PortMax uint16
}

type TXWS_Message_Frame struct {
	Frame_Type string 	
	Frame_To_WSID string 		 		 
	Frame_Payload string 		 
}

type TXWS_MessageType_Candidate struct {
	MessageType string
	PeerID string
	Candidate string
}

type TXWS_MessageType_Answer struct {
	MessageType string
	PeerID string
	PeerDescription string
}



func NewMuxer(options Options) *Muxer {
	tmp := Muxer{Options: options, ClientACK: time.NewTimer(time.Second * 20), StreamACK: time.NewTimer(time.Second * 20), streams: make(map[int8]*Stream)}
	//go tmp.WaitCloser()
	//options NewMuxer {[stun:stun.l.google.com:19302]   [] 60000 61000}
	return &tmp
}

func (element *Muxer) NewPeerConnection(configuration webrtc.Configuration) (*webrtc.PeerConnection, error) {
	 
	if len(element.Options.ICEServers) > 0 {
		log.Println("Set ICEServers", element.Options.ICEServers)
		configuration.ICEServers = append(configuration.ICEServers, webrtc.ICEServer{
			URLs:           element.Options.ICEServers,
			Username:       element.Options.ICEUsername,
			Credential:     element.Options.ICECredential,
			CredentialType: webrtc.ICECredentialTypePassword,
		})
	} else {
		configuration.ICEServers = append(configuration.ICEServers, webrtc.ICEServer{
			URLs: []string{"stun:stun.l.google.com:19302"},
		})
	}

//	vcodec := flag.String("vcodec", "H264", "video codec type (H264/VP8/VP9)")
//	acodec := flag.String("acodec", "OPUS", "audio codec type (OPUS)")
//	log.Println("codecs video:", *vcodec, " audio:" , *acodec)
//	flag.Parse()

	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}
/*
	switch *vcodec {
	case "H264":
		m.RegisterCodec(webrtc.NewRTPH264Codec(webrtc.DefaultPayloadTypeH264, 90000))
	case "VP8":
		m.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))
	case "VP9":
		m.RegisterCodec(webrtc.NewRTPVP9Codec(webrtc.DefaultPayloadTypeVP9, 90000))
	default:
		log.Println("Not support video codec", *vcodec)
		return nil, errors.New("Not support video codec")
	}

	switch *acodec {
	case "OPUS":
		m.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
	default:
		log.Println("Not support audio codec", *acodec)
		return nil, errors.New("Not support audio codec")
	}

  
	m.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
	*/

	i := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		return nil, err
	}
	 
	s := webrtc.SettingEngine{}
	if element.Options.PortMin > 0 && element.Options.PortMax > 0 && element.Options.PortMax > element.Options.PortMin {
		s.SetEphemeralUDPPortRange(element.Options.PortMin, element.Options.PortMax)
		log.Println("Set UDP ports to", element.Options.PortMin, "-", element.Options.PortMax)
	}
	 
	if len(element.Options.ICECandidates) > 0 {
		s.SetNAT1To1IPs(element.Options.ICECandidates, webrtc.ICECandidateTypeHost)
		log.Println("Set ICECandidates", element.Options.ICECandidates)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i), webrtc.WithSettingEngine(s))
	
	return api.NewPeerConnection(configuration)
}
 




func (element *Muxer) Handle_Offer(streams []av.CodecData, offer_sdp64 string, 
	//   offer_peerid string, ws_conn websocket.Conn, resp_WSID string) (string, error) {
		  offer_peerid string, resp_WSID string, Stream_ID string) (string, error) {
  
  //resp_WSID socketnr van VPS waar het antwoord heen moet
  var WriteHeaderSuccess bool

  //fmt.Println("adapter.go Handle_Offer streams:",len(streams))
  if(len(streams) == 0){
	   
	  fmt.Println("adapter.go Handle_Offer streams:",len(streams))
	  return "", ErrorNotFound
  }
  sdpB, err := base64.StdEncoding.DecodeString(offer_sdp64)   // binnenkomend offer
  if err != nil {
	  return "", err
  }
  offer := webrtc.SessionDescription{
	  Type: webrtc.SDPTypeOffer,
	  SDP:  string(sdpB),
  }
  
  peerConnection, err := element.NewPeerConnection(webrtc.Configuration{
	  SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
  })
  if err != nil {
	  return "", err
  }
		  
  Pc_guid_list[offer_peerid] = PeerConnection_Guid{peerConnection}  //id opslaan
  



 // const rtcpPLIInterval = time.Second * 3
  fmt.Println("Handle Offer")
  //var localAudioTrack *webrtc.TrackRemote = nil

  //localAudioTrack, err = peerConnection.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "audio", "pion")
	  //if err != nil {
	  //	log.Printf("peerConnection.NewTrack(OPUS) failed. %v\n", err)
	  //	return
	  //}

  //	_, err = peerConnection.AddTrack(localAudioTrack)
	  //if err != nil {
	  //	log.Printf("peerConnection.AddTrack(Audio) failed. %v\n", err)
	  //	return
	  //}
  //*webrtc.RTPSender , err = peerConnection.AddTrack(localAudioTrack)
/******************
  peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	  fmt.Printf("peerConnection.OnTrack(%v)\n", remoteTrack)
	  for {
		  rtpPacket, _ , err := remoteTrack.ReadRTP() //komt binnen van webclient
		  if err != nil {
			  log.Println("remoteTrack.ReadRTP:", err)
			  return
		  }
		  //fmt.Println("ReadRTP ", rtpPacket.String())
		  //fmt.Printf("RXD %02x %02x %02x %02x\n", rtpPacket.Payload[0], rtpPacket.Payload[1], rtpPacket.Payload[2], rtpPacket.Payload[3])
		  sip.Send_To_UPD(rtpPacket.Payload)
	  }
	  
	  
	   
  })
   
  *******************/ 

  

   
   
  //fmt.Println("adapter.go Handle_Offer stream:", streams)
  // streams = uitgaande streams, 
  for ii, ii2 := range streams {
	  var track *webrtc.TrackLocalStaticSample
	  if ii2.Type().IsVideo() {
		  //fmt.Println("adapter.go Handle_Offer IsVideo")
		  if ii2.Type() == av.H264 {
			  fmt.Println("adapter.go Handle_Offer IsVideo av.H264")
		  }
	  }
	  if ii2.Type().IsAudio() {
		  //fmt.Println("adapter.go Handle_Offer IsAudio")
		  switch ii2.Type() {
		  case av.PCM_ALAW:
			  fmt.Println("adapter.go Handle_Offer IsAudio:", webrtc.MimeTypePCMA)
		  case av.PCM_MULAW:
			  fmt.Println("adapter.go Handle_Offer IsAudio:", webrtc.MimeTypePCMU)
		  case av.OPUS:
			  fmt.Println("adapter.go Handle_Offer IsAudio:", webrtc.MimeTypeOpus)
		  default:
			  fmt.Println("adapter.go Handle_Offer IsAudio onbekend" )
			  continue
		  }
	  }
	  element.streams[int8(ii)] = &Stream{track: track, codec: ii2}
  }

  defer func() {
	  if !WriteHeaderSuccess {
		  err = element.Close()
		  if err != nil {
			  log.Println(err)
		  }
	  }
  }()
  for i, i2 := range streams {
	  var track *webrtc.TrackLocalStaticSample
	  if i2.Type().IsVideo() {
		  if i2.Type() == av.H264 {
			  track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
				  MimeType: webrtc.MimeTypeH264,
			  }, "pion-rtsp-video", "pion-video")
			  if err != nil {
				  return "", err
			  }
			  if rtpSender, err := peerConnection.AddTrack(track); err != nil {
				  return "", err
			  } else {
				  go func() {
					  rtcpBuf := make([]byte, 1500)
					  for {
						  if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
							  return
						  }
					  }
				  }()
			  }
		  }
	  } else if i2.Type().IsAudio() {
		  AudioCodecString := webrtc.MimeTypePCMA
		  switch i2.Type() {
		  case av.PCM_ALAW:
			  AudioCodecString = webrtc.MimeTypePCMA
		  case av.PCM_MULAW:
			  AudioCodecString = webrtc.MimeTypePCMU
		  case av.OPUS:
			  AudioCodecString = webrtc.MimeTypeOpus
		  default:
			  log.Println(ErrorIgnoreAudioTrack)
			  continue
		  }
		  track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
			  MimeType:  AudioCodecString,
			  Channels:  uint16(i2.(av.AudioCodecData).ChannelLayout().Count()),
			  ClockRate: uint32(i2.(av.AudioCodecData).SampleRate()),
		  }, "pion-rtsp-audio", "pion-rtsp-audio")
		  if err != nil {
			  return "", err
		  }
		  if rtpSender, err := peerConnection.AddTrack(track); err != nil {
			  return "", err
		  } else {
			  go func() {
				  rtcpBuf := make([]byte, 1500)
				  for {
					  if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
						  return
					  }
				  }
			  }()
		  }
	  }
	  element.streams[int8(i)] = &Stream{track: track, codec: i2}
  }
  if len(element.streams) == 0 {
	  return "", ErrorNotTrackAvailable
  }





  peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {	 
	  element.status = connectionState
	  //log.Println("OnICEConnectionStateChange:",connectionState )
	  if connectionState == webrtc.ICEConnectionStateDisconnected {
		  element.Close()
	  }
  })
  
   peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
	   
	  //fmt.Println("adapter.go Handle_Offer OnDataChannel:", Stream_ID)
	  strm_struct, err := Storage.StreamControl(Stream_ID )
	  if err != nil {
		  fmt.Println("adapter.go Handle_Offer Storage.StreamControl(streamID) error:" ,  err.Error())
		  return
	  }

	  d.OnMessage(func(msg webrtc.DataChannelMessage) {
		  element.ClientACK.Reset(5 * time.Second)
		  //per client
		  if msg.IsString {
			  //log.Println("Handle_Offer DataChannel OnMessage :", string(msg.Data))
			  log.Println("Handle_Offer DataChannel OnMessage")
			  Handle_Incomming_Data(  string(msg.Data), strm_struct)
		  } else {
			  log.Println("Handle_Offer DataChannel OnMessage Binair Not Implemented")
		  }

	  })
	  d.OnOpen(func() {
		  //log.Println("Handle_Offer DataChannel OnOpen ")  
		  //per client
	  })
	  d.OnClose(func() {
		  //log.Println("Handle_Offer DataChannel OnClose ")  
	  })
	  d.OnError(func(err error)  {
		  log.Println("Handle_Offer DataChannel OnError :", err.Error())  
	  })
  })

  peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
	  if c == nil {
		  return
	  }
	  //if Verbose {
	  //	fmt.Println("Outgoing CANDIDATE !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	  //} 
	  outbound, _ := json.Marshal(c.ToJSON())
	  msg := TXWS_MessageType_Candidate{"icecandidate", element.pc_guid, string(outbound)}
	  payload, err := json.Marshal(msg) 
	  if(err != nil){
		  return
	  }   
	  
	  txf:= TXWS_Message_Frame{"WEBRTC_RESP", resp_WSID, string(payload) }
	  txd, err := json.Marshal(txf) 
	  if(err != nil){
		  return
	  }  
	  fmt.Println("Outgoing message MessageType :", msg.MessageType);
	  err = Write_Array(txd)
	  // err = ws_conn.WriteMessage(websocket.TextMessage, txd)
	  if(err != nil){
		  fmt.Println("Handle Offer icecandidate Write_Array error :", err.Error()  )
	  }		 

  })  

  err = peerConnection.SetRemoteDescription(offer); 
  if(err != nil) {
	  fmt.Println("adapter.go peerConnection.SetRemoteDescription(offer) err:", err.Error())
	  return "", err
  }

  gatherCompletePromise := webrtc.GatheringCompletePromise(peerConnection)
  
  answer, err := peerConnection.CreateAnswer(nil)
  if err != nil {
	  return "", err
  }
  if err = peerConnection.SetLocalDescription(answer); err != nil {
	  return "", err
  }

  element.pc = peerConnection
  
   
  element.pc_guid = offer_peerid 
  //fmt.Println("element.pc_guid:",element.pc_guid)
  //if Verbose {
  //	fmt.Println("Outgoing ANSWER !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
  //}
  resp := peerConnection.LocalDescription()
  msg := TXWS_MessageType_Answer{"answer", element.pc_guid , base64.StdEncoding.EncodeToString([]byte(resp.SDP))}
  payload, err := json.Marshal(msg) 
  if(err != nil){
	  return "", err
  }  
  
  txf:= TXWS_Message_Frame{"WEBRTC_RESP", resp_WSID, string(payload) }
  txd, err := json.Marshal(txf) 
  if(err != nil){
	  return "", err
  }  
  
  fmt.Println("Outgoing message MessageType :", msg.MessageType); 
  err = Write_Array(txd)
  if(err != nil){
	  fmt.Println("Handle Offer answer Write_Array error :", err.Error()  )
  }
   
  waitT := time.NewTimer(time.Second * 10)
  select {
  case <-waitT.C:
	  return "", errors.New("gatherCompletePromise wait")
  case <-gatherCompletePromise:
	  //Connected
  }
  
  //resp := peerConnection.LocalDescription()
  WriteHeaderSuccess = true
  return base64.StdEncoding.EncodeToString([]byte(resp.SDP)), nil

}

 
func (element *Muxer) WriteHeader(streams []av.CodecData, sdp64 string, stream_ID string) (string, error) {
	var WriteHeaderSuccess bool

	if len(streams) == 0 {
		return "", ErrorNotFound
	}
	sdpB, err := base64.StdEncoding.DecodeString(sdp64)   // offer
	if err != nil {
		return "", err
	}
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  string(sdpB),
	}
	
	peerConnection, err := element.NewPeerConnection(webrtc.Configuration{
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	})
	if err != nil {
		return "", err
	}
	defer func() {
		if !WriteHeaderSuccess {
			err = element.Close()
			if err != nil {
				log.Println(err)
			}
		}
	}()
	for i, i2 := range streams {
		var track *webrtc.TrackLocalStaticSample
		if i2.Type().IsVideo() {
			if i2.Type() == av.H264 {
				track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
					MimeType: webrtc.MimeTypeH264,
				}, "pion-rtsp-video", "pion-video")
				if err != nil {
					return "", err
				}
				if rtpSender, err := peerConnection.AddTrack(track); err != nil {
					return "", err
				} else {
					go func() {
						rtcpBuf := make([]byte, 1500)
						for {
							if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
								return
							}
						}
					}()
				}
			}
		} else if i2.Type().IsAudio() {
			AudioCodecString := webrtc.MimeTypePCMA
			switch i2.Type() {
			case av.PCM_ALAW:
				AudioCodecString = webrtc.MimeTypePCMA
			case av.PCM_MULAW:
				AudioCodecString = webrtc.MimeTypePCMU
			case av.OPUS:
				AudioCodecString = webrtc.MimeTypeOpus
			default:
				log.Println(ErrorIgnoreAudioTrack)
				continue
			}
			track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
				MimeType:  AudioCodecString,
				Channels:  uint16(i2.(av.AudioCodecData).ChannelLayout().Count()),
				ClockRate: uint32(i2.(av.AudioCodecData).SampleRate()),
			}, "pion-rtsp-audio", "pion-rtsp-audio")
			if err != nil {
				return "", err
			}
			if rtpSender, err := peerConnection.AddTrack(track); err != nil {
				return "", err
			} else {
				go func() {
					rtcpBuf := make([]byte, 1500)
					for {
						if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
							return
						}
					}
				}()
			}
		}
		element.streams[int8(i)] = &Stream{track: track, codec: i2}
	}
	if len(element.streams) == 0 {
		return "", ErrorNotTrackAvailable
	}

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {		 
		element.status = connectionState
		log.Println("adapter.go OnICEConnectionStateChange:",connectionState )
		if connectionState == webrtc.ICEConnectionStateDisconnected {
			element.Close()
		}
	})

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		//log.Println("adapter.go WriteHeader OnDataChannel:", stream_ID  )
		strm_struct, err := Storage.StreamControl(stream_ID )
		if err != nil {
			fmt.Println("adapter.go WriteHeader Storage.StreamControl(streamID) error:" ,  err.Error())
			return
		}

		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			element.ClientACK.Reset(5 * time.Second)
			//per client
			if msg.IsString {
				//log.Println("WriteHeader DataChannel OnMessage :", string(msg.Data) )
				log.Println("WriteHeader DataChannel OnMessage ")  
				
				Handle_Incomming_Data(string(msg.Data), strm_struct)

			} else {
				log.Println("WriteHeader DataChannel OnMessage Binair Not Implemented")
			}

		})
		d.OnOpen(func() {
			//log.Println("WriteHeader DataChannel OnOpen ")  
			//per client
		})
		d.OnClose(func() {
			//log.Println("WriteHeader DataChannel OnClose ")  
		})
		d.OnError(func(err error)  {
			log.Println("WriteHeader DataChannel OnError :", err.Error())  
		})
		 
	})
	 

	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		return "", err
	}
	gatherCompletePromise := webrtc.GatheringCompletePromise(peerConnection)
	
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "", err
	}
	element.pc = peerConnection

	waitT := time.NewTimer(time.Second * 10)
	select {
	case <-waitT.C:
		return "", errors.New("gatherCompletePromise wait")
	case <-gatherCompletePromise:
		//Connected
	}
	resp := peerConnection.LocalDescription()
	WriteHeaderSuccess = true
	return base64.StdEncoding.EncodeToString([]byte(resp.SDP)), nil

}


//packetje naar webclient
func (element *Muxer) WritePacket(pkt av.Packet) (err error) {
	
	log.Println("WritePacket", pkt.Time, element.stop, 
	        webrtc.ICEConnectionStateConnected, 
	        pkt.Idx, pkt.Type) //, 
	       // element.streams[pkt.Idx])

	var WritePacketSuccess bool
	defer func() {
		if !WritePacketSuccess {
			element.Close()
		}
	}()
	if element.stop {
		return ErrorClientOffline
	}
	if element.status == webrtc.ICEConnectionStateChecking {
		WritePacketSuccess = true
		return nil
	}
	if element.status != webrtc.ICEConnectionStateConnected {
		return nil
	}
	if tmp, ok := element.streams[pkt.Idx]; ok {
		element.StreamACK.Reset(10 * time.Second)
		if len(pkt.Data) < 5 {
			return nil
		}
		switch tmp.codec.Type() {
		case av.H264:
			nalus, _ := h264parser.SplitNALUs(pkt.Data)
			for _, nalu := range nalus {
				naltype := nalu[0] & 0x1f
				if naltype == 5 {
					codec := tmp.codec.(h264parser.CodecData)
					err = tmp.track.WriteSample(media.Sample{Data: append([]byte{0, 0, 0, 1}, bytes.Join([][]byte{codec.SPS(), codec.PPS(), nalu}, []byte{0, 0, 0, 1})...), Duration: pkt.Duration})

				} else if naltype == 1 {
					err = tmp.track.WriteSample(media.Sample{Data: append([]byte{0, 0, 0, 1}, nalu...), Duration: pkt.Duration})
				}
				if err != nil {
					return err
				}
			}
			WritePacketSuccess = true
			return
			/*

				if pkt.IsKeyFrame {
					pkt.Data = append([]byte{0, 0, 0, 1}, bytes.Join([][]byte{codec.SPS(), codec.PPS(), pkt.Data[4:]}, []byte{0, 0, 0, 1})...)
				} else {
					pkt.Data = pkt.Data[4:]
				}

			*/
		case av.PCM_ALAW:
		case av.OPUS:
		case av.PCM_MULAW:
		case av.AAC:
			//TODO: NEED ADD DECODER AND ENCODER
			return ErrorCodecNotSupported
		case av.PCM:
			//TODO: NEED ADD ENCODER
			return ErrorCodecNotSupported
		default:
			return ErrorCodecNotSupported
		}
		err = tmp.track.WriteSample(media.Sample{Data: pkt.Data, Duration: pkt.Duration})
		if err == nil {
			WritePacketSuccess = true
		}
		return err
	} else {
		WritePacketSuccess = true
		return nil
	}
}

func (element *Muxer) Close() error {
	element.stop = true
	if element.pc != nil {
		err := element.pc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (element *Muxer) Find_PC(Peer_Guid string) error {

	//for i, i2 := range element {
	
	
	
	//}
	return nil

}

func Response_Error(PeerID string, resp_WSID string, Error_msg string){
	if Verbose {
		fmt.Println("Outgoing ERROR !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
	msg := TXWS_MessageType_Answer{"error", PeerID , Error_msg}
	//msg := TXWS_MessageType_Answer{"answer", element.pc_guid , "base64.StdEncoding.EncodeToString"}
	payload, err := json.Marshal(msg) 
	if(err != nil){
		return 
	}  
	
	txf:= TXWS_Message_Frame{"WEBRTC_RESP", resp_WSID, string(payload) }
	txd, err := json.Marshal(txf) 
	if(err != nil){
		return  
	}  
	err = Write_Array(txd)
	if(err != nil){
		fmt.Println("Response_Error.Write_Array error :", err.Error()  )
	}
}
