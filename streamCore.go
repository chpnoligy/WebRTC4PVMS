package main

import (
	//"WebRTC4PVMS/sip"
	"fmt"
	"log"
	"math"

	"net/url"
	"strings"
	"time"

	"github.com/deepch/vdk/format/rtmp"


	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/rtspv2"
)

//StreamServerRunStreamDo stream run do mux
func StreamServerRunStreamDo(streamID string) {
	//fmt.Println("StreamServerRunStreamDo :", streamID)
	var status int
	defer func() {
		//TODO fix it no need unlock run if delete stream
		if status != 2 {
			Storage.StreamUnlock(streamID  )
		}
	}()
	
	for {
		 
		//fmt.Println("StreamServerRunStreamDo :", streamID , " :", channelID)
		//log.Println("Run stream")
		 
		opt, err := Storage.StreamControl(streamID )
		if err != nil {
			fmt.Println("Storage.StreamChannelControl(streamID, channelID) error:" ,  err.Error())
			return
		}
		 
		_ , err = url.ParseRequestURI(opt.URL)
		if err != nil {
			fmt.Println("ParseRequestURI:", opt.URL," error:" , err.Error())
			return
		}

		if opt.OnDemand && !Storage.ClientHas(streamID ) {
			fmt.Println("Stop stream no client")
			return
		}

		status, err = StreamServerRunStream(streamID , opt)  //RTSP stream opstarten
		if status > 0 {
			fmt.Println("Stream exit by signal or no client")
			return
		}

		if err != nil {
			fmt.Println("Stream error restart stream", err. Error())
		}
		time.Sleep(2 * time.Second)

	}
}

 
//hier word verbinding met de camera gemaakt

func StreamServerRunStream(streamID string,  opt *StreamST) (int, error) {
	if url, err := url.Parse(opt.URL); err == nil && strings.ToLower(url.Scheme) == "rtmp" {
		return StreamServerRunStreamRTMP(streamID, opt)
	}

	//log.Println("StreamServerRunStream:", opt.URL)
	url := opt.URL
    if(opt.Username != "" || opt.Password != ""){ 
		i := strings.Index(opt.URL, "@")
		if(i < 0) {
			i = strings.Index(opt.URL, "://")  // 'rtsp://192.168.178.205:554/ of rtsp://www.chp.nl:554 
			if(i > -1) {
				url1 := opt.URL[0:i]
				url2 := opt.URL[i + 3:len(opt.URL)]
				url = url1 + "://" + opt.Username + ":" + opt.Password + "@" + url2
			} 

		}
	}
	
	fmt.Println("StreamServerRunStream url:", url)

	keyTest := time.NewTimer(20 * time.Second)
	checkClients := time.NewTimer(20 * time.Second)
	var start bool
	var fps int
	var preKeyTS = time.Duration(0)
	var Seq []*av.Packet
	//RTSPClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{URL: opt.URL, 
	RTSPClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{URL: url, 
									InsecureSkipVerify: opt.InsecureSkipVerify, 
									DisableAudio: !opt.Audio, 
									DialTimeout: 3 * time.Second, 
									ReadWriteTimeout: 5 * time.Second, 
									Debug: opt.Debug, 
									OutgoingProxy: true})
	if err != nil {
		return 0, err
	}
	 
	//SIPClient, err := sip.SIP_Invite()
	fmt.Println("Dial Done wacht op codec")
	 
	Storage.StreamStatus(streamID, ONLINE)
	defer func() {
		RTSPClient.Close()
		Storage.StreamStatus(streamID, OFFLINE)
		Storage.StreamHLSFlush(streamID)
	}()
	var WaitCodec bool
	/*
		Example wait codec
	*/
	if RTSPClient.WaitCodec {
		WaitCodec = true
	} else {
		if len(RTSPClient.CodecData) > 0 {
			Storage.StreamCodecsUpdate(streamID, RTSPClient.CodecData, RTSPClient.SDPRaw)
		}
	}
	
	log.Println("Success connection RTSP stream:", streamID)

	var ProbeCount int
	var ProbeFrame int
	var ProbePTS time.Duration
	Storage.NewHLSMuxer(streamID)
	defer Storage.HLSMuxerClose(streamID)
	for {
		select {
		//Check stream have clients
		case <-checkClients.C:
			if opt.OnDemand && !Storage.ClientHas(streamID) {
				return 1, ErrorStreamNoClients
			}
			checkClients.Reset(20 * time.Second)
		//Check stream send key
		case <-keyTest.C:
			return 0,  ErrorStreamNoVideo
		//Read core signals
		case signals := <-opt.signals:
			switch signals {
			case SignalStreamStop:
				return 2,  ErrorStreamStopCoreSignal
			case SignalStreamRestart:
				return 0,  ErrorStreamRestart
			case SignalStreamClient:
				return 1,  ErrorStreamNoClients
			}
		//Read rtsp signals
		case signals := <-RTSPClient.Signals:
			fmt.Println("RTSPClient.Signals")
			switch signals {
			case rtspv2.SignalCodecUpdate:
				 Storage.StreamCodecsUpdate(streamID, RTSPClient.CodecData, RTSPClient.SDPRaw)
				WaitCodec = false
			case rtspv2.SignalStreamRTPStop:
				return 0,  ErrorStreamStopRTSPSignal
			}
		case packetRTP := <-RTSPClient.OutgoingProxyQueue:
			//OutgoingProxyQueue chan *[]byte
			//fmt.Println("RTSPClient.OutgoingProxyQueue")
			Storage.StreamCastProxy(streamID, packetRTP)

		case packetAV := <-RTSPClient.OutgoingPacketQueue:
			//OutgoingPacketQueue chan *av.Packet
			//fmt.Println("RTSPClient.OutgoingPacketQueue :", packetAV.Type)
			if WaitCodec {
				continue //nog geen Codec binnen
			}

			if packetAV.IsKeyFrame {
				keyTest.Reset(20 * time.Second)
				if preKeyTS > 0 {
					Storage.StreamHLSAdd(streamID, Seq, packetAV.Time-preKeyTS)
					Seq = []*av.Packet{}
				}
				preKeyTS = packetAV.Time
			}
			Seq = append(Seq, packetAV)
			
			Storage.StreamCast(streamID, packetAV)
			/*
			   HLS LL Test
			*/
			if packetAV.IsKeyFrame && !start {
				start = true
			}
			/*
				FPS mode probe
			*/
			if start {
				ProbePTS += packetAV.Duration
				ProbeFrame++
				if packetAV.IsKeyFrame && ProbePTS.Seconds() >= 1 {
					ProbeCount++
					if ProbeCount == 2 {
						fps = int(math.Round(float64(ProbeFrame) / ProbePTS.Seconds()))
					}
					ProbeFrame = 0
					ProbePTS = 0
				}
			}
			if start && fps != 0 {
				//TODO fix it
				packetAV.Duration = time.Duration((float32(1000)/float32(fps))*1000*1000) * time.Nanosecond
				Storage.HlsMuxerSetFPS(streamID, fps)
			    Storage.HlsMuxerWritePacket(streamID, packetAV)
			}
		}
	}
}

func StreamServerRunStreamRTMP(streamID string, opt *StreamST) (int, error) {
	keyTest := time.NewTimer(20 * time.Second)
	checkClients := time.NewTimer(20 * time.Second)
	OutgoingPacketQueue := make(chan *av.Packet, 1000)
	Signals := make(chan int, 100)
	var start bool
	var fps int
	var preKeyTS = time.Duration(0)
	var Seq []*av.Packet

	conn, err := rtmp.DialTimeout(opt.URL, 3*time.Second)
	if err != nil {
		return 0, err
	}

	Storage.StreamStatus(streamID,ONLINE)
	defer func() {
		conn.Close()
		Storage.StreamStatus(streamID, OFFLINE)
		Storage.StreamHLSFlush(streamID)
	}()
	var WaitCodec bool

	codecs, err := conn.Streams()
	if err != nil {
		return 0, err
	}
	preDur := make([]time.Duration, len(codecs))
	Storage.StreamCodecsUpdate(streamID, codecs, []byte{})

	log.Println("Success connection RTMP stream:", streamID )

	var ProbeCount int
	var ProbeFrame int
	var ProbePTS time.Duration
	Storage.NewHLSMuxer(streamID)
	defer Storage.HLSMuxerClose(streamID)

	go func() {
		for {
			ptk, err := conn.ReadPacket()
			if err != nil {
				break
			}
			OutgoingPacketQueue <- &ptk
		}
		Signals <- 1
	}()

	for {
		select {
		//Check stream have clients
		case <-checkClients.C:
			if opt.OnDemand && !Storage.ClientHas(streamID) {
				return 1, ErrorStreamNoClients
			}
			checkClients.Reset(20 * time.Second)
		//Check stream send key
		case <-keyTest.C:
			return 0,  ErrorStreamNoVideo
		//Read core signals
		case signals := <-opt.signals:
			switch signals {
			case SignalStreamStop:
				return 2, ErrorStreamStopCoreSignal
			case SignalStreamRestart:
				return 0, ErrorStreamRestart
			case SignalStreamClient:
				return 1, ErrorStreamNoClients
			}
		//Read rtsp signals
		case <-Signals:
			return 0, ErrorStreamStopRTSPSignal
		case packetAV := <-OutgoingPacketQueue:
			if preDur[packetAV.Idx] != 0 {
				packetAV.Duration = packetAV.Time - preDur[packetAV.Idx]
			}

			preDur[packetAV.Idx] = packetAV.Time

			if WaitCodec {
				continue
			}

			if packetAV.IsKeyFrame {
				keyTest.Reset(20 * time.Second)
				if preKeyTS > 0 {
					Storage.StreamHLSAdd(streamID, Seq, packetAV.Time-preKeyTS)
					Seq = []*av.Packet{}
				}
				preKeyTS = packetAV.Time
			}
			Seq = append(Seq, packetAV)
			Storage.StreamCast(streamID, packetAV)
			/*
			   HLS LL Test
			*/
			if packetAV.IsKeyFrame && !start {
				start = true
			}
			/*
				FPS mode probe
			*/
			if start {
				ProbePTS += packetAV.Duration
				ProbeFrame++
				if packetAV.IsKeyFrame && ProbePTS.Seconds() >= 1 {
					ProbeCount++
					if ProbeCount == 2 {
						fps = int(math.Round(float64(ProbeFrame) / ProbePTS.Seconds()))
					}
					ProbeFrame = 0
					ProbePTS = 0
				}
			}
			if start && fps != 0 {
				//TODO fix it
				packetAV.Duration = time.Duration((float32(1000)/float32(fps))*1000*1000) * time.Nanosecond
				Storage.HlsMuxerSetFPS(streamID, fps)
				Storage.HlsMuxerWritePacket(streamID,  packetAV)
			}
		}
	}
}
