package main

import (
	 
	"fmt"

	"time"

	"github.com/deepch/vdk/av"
	"github.com/imdario/mergo"
)

// StreamMake check stream exist
func (obj *StorageST) StreamMake(strm StreamST) StreamST {
	fmt.Println("CStreamMake url:" , strm.URL)

	stream := obj.StreamDefaults
	if err := mergo.Merge(&stream, strm); err != nil {
		// Just ignore the default values and continue
		stream = strm
		fmt.Println("mergo.Merge(&channel, val) error:", err.Error())
	}
	//make client's
	stream.clients = make(map[string]ClientST)
	//make last ack
	stream.ack = time.Now().Add(-255 * time.Hour)
	//make hls buffer
	stream.hlsSegmentBuffer = make(map[int]SegmentOld)
	//make signals buffer chain
	stream.signals = make(chan int, 100)
	return stream
}

// StreamRunAll run all stream go
func (obj *StorageST) StreamRunAll() {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	for k, v := range obj.Streams {
		if !v.OnDemand {
			v.runLock = true
			go StreamServerRunStreamDo(k )
			obj.Streams[k] = v
		}
	}
}

// StreamChannelRun one stream and lock
func (obj *StorageST) StreamRun(streamID string ) {
	 
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[streamID]; ok {
		if !streamTmp.runLock {
			streamTmp.runLock = true   //runlock false = streamopstarten
			obj.Streams[streamID] = streamTmp
			go StreamServerRunStreamDo(streamID )
		}
	}
}

// StreamChannelUnlock unlock status to no lock
func (obj *StorageST) StreamUnlock(streamID string ) {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[streamID]; ok {
		streamTmp.runLock = false 
		obj.Streams[streamID] = streamTmp
	}
}

// StreamChannelControl get stream aan de hand van streamid  
func (obj *StorageST) StreamControl(key string) (*StreamST, error) {
	 
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[key]; ok {
		return &streamTmp, nil
	}
	return nil, ErrorStreamNotFound
}
 
// StreamChannelExist check stream exist
func (obj *StorageST) StreamExist(streamID string) bool {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	 
	if streamTmp, ok := obj.Streams[streamID]; ok {
		streamTmp.ack = time.Now()
		obj.Streams[streamID] = streamTmp 
		return ok 
	}
	return false


}

// StreamChannelExist check stream exist
func (obj *StorageST) URLExist(url string) bool  {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	for _ , streamTmp := range obj.Streams {
		if streamTmp.URL == url { 
			return true
		}
	}
	return false
}
 

// StreamChannelReload reload stream
func (obj *StorageST) CStreamReload(uuid string ) error {
	obj.mutex.RLock()
	defer obj.mutex.RUnlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
		streamTmp.signals <- SignalStreamRestart
		return nil
	}
	return ErrorStreamNotFound
}

// StreamInfo return stream info
func (obj *StorageST) CStreamInfo(uuid string ) (*StreamST, error) {
	obj.mutex.RLock()
	defer obj.mutex.RUnlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
		return &streamTmp, nil
	}
	return nil, ErrorStreamNotFound
}

// StreamChannelCodecs get stream codec storage or wait
func (obj *StorageST) StreamCodecs(streamID string ) ([]av.CodecData, error) {
	for i := 0; i < 100; i++ {  //wacht max 5 sec
		obj.mutex.RLock()
		streamTmp, ok := obj.Streams[streamID]
		obj.mutex.RUnlock()
		if !ok {
			return nil, ErrorStreamNotFound
		}
		

		if streamTmp.codecs != nil {
			return streamTmp.codecs, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil, ErrorStreamChannelCodecNotFound
}

// StreamChannelStatus change stream status
func (obj *StorageST) StreamStatus(key string,  val int) {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[key]; ok {
		obj.Streams[key] = streamTmp
	}
}

// StreamChannelCast broadcast stream
func (obj *StorageST) StreamCast(key string, avp *av.Packet) {
	obj.mutex.Lock()
	fmt.Println("StreamCast:", avp.Type)
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[key]; ok {
		if len(streamTmp.clients) > 0 {
			for _, i2 := range streamTmp.clients {
				if i2.mode == RTSP {
					//RTSP client krijgt het van de proxy
					continue
				}
				if len(i2.outgoingAVPacket) < 1000 {
					//fmt.Println("i2.outgoingAVPacket <- avp")
					i2.outgoingAVPacket <- avp   //stuur naar channel
				} else if len(i2.signals) < 10 {
					i2.signals <- SignalStreamStop
				}
			}
		}
		streamTmp.ack = time.Now()			 
		obj.Streams[key] = streamTmp
	}
}

// StreamChannelCastProxy broadcast stream 
//Als de client een RTSP client is bv VLC
func (obj *StorageST) StreamCastProxy(key string,  val *[]byte) {
	obj.mutex.Lock()
	//fmt.Println("StreamCastProxy")
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[key]; ok {
		if len(streamTmp.clients) > 0 {
			for _, i2 := range streamTmp.clients {
				if i2.mode != RTSP {
					continue
				}
				if len(i2.outgoingRTPPacket) < 1000 {
					
					i2.outgoingRTPPacket <- val   //send to channel
				} else if len(i2.signals) < 10 {
					i2.signals <- SignalStreamStop
				}
			}
		}
		streamTmp.ack = time.Now()		 
		obj.Streams[key] = streamTmp	 
	}
}

// StreamChannelCodecsUpdate update stream codec storage
func (obj *StorageST) StreamCodecsUpdate(streamID string,val []av.CodecData, sdp []byte) {
	obj.mutex.Lock()

	fmt.Println("StreamCodecsUpdate ID: ",streamID, " val:", len(val), " type:", val[0].Type().String()  )
	//
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[streamID]; ok {	 
		streamTmp.codecs = val
		streamTmp.sdp = sdp
		obj.Streams[streamID] = streamTmp 
	}
}

// StreamChannelSDP codec storage or wait
/*
func (obj *StorageST) obsStreamSDP(streamID string) ([]byte, error) {
	for i := 0; i < 100; i++ {
		obj.mutex.RLock()
		streamTmp, ok := obj.Streams[streamID]
		obj.mutex.RUnlock()
		if !ok {
			return nil, ErrorStreamNotFound
		}

		if len(streamTmp.sdp) > 0 {
			return streamTmp.sdp, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil, ErrorStreamNotFound
}
 
// StreamChannelAdd add stream
func (obj *StorageST)  obsStreamAdd(uuid string, strm StreamST) error {
	
	fmt.Println("storageStreamChnl .StreamCAdd",uuid, " url:" , strm.URL)
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if _, ok := obj.Streams[uuid]; !ok {
		fmt.Println("storageStreamChnl return ErrorStreamNotFound :", uuid)
		return ErrorStreamNotFound
	}
	if _, ok := obj.Streams[uuid]; ok {
		fmt.Println("storageStreamChnl return ErrorStreamChannelAlreadyExists :", uuid)
		return ErrorStreamChannelAlreadyExists
	}
	strm = obj.StreamMake(strm)
	 
	if !strm.OnDemand {
		strm.runLock = true
		obj.Streams[uuid] = strm
		go StreamServerRunStreamDo(uuid )
	}
	obj.Streams[uuid] = strm
	err := obj.SaveConfig()
	if err != nil {
		return err
	}
	return nil
}

// StreamEdit edit stream
 
func (obj *StorageST) obsStreamEdit(uuid string,  strm StreamST) error {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
		 
			if streamTmp.runLock {
				streamTmp.signals <- SignalStreamStop
			}
			
			strm = obj.StreamMake(strm)
			obj.Streams[uuid]  = strm
			if !strm.OnDemand {
				strm.runLock = true
				obj.Streams[uuid]  = strm
				go StreamServerRunStreamDo(uuid )
			}
			err := obj.SaveConfig()
			if err != nil {
				return err
			}
			return nil
		 
	}
	return ErrorStreamNotFound
}
*/
// StreamChannelDelete stream
/* 
func (obj *StorageST) obsCStreamDelete(uuid string ) error {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
		 
			if streamTmp.runLock {
				streamTmp.signals <- SignalStreamStop
			}
			 
			delete(obj.Streams, uuid)
			err := obj.SaveConfig()
			if err != nil {
				return err
			}
			return nil
		 
	}
	return ErrorStreamNotFound
}
*/
 
func (obj *StorageST) StopAllStreams() {
	obj.mutex.RLock()
	defer obj.mutex.RUnlock()
	for _, streamTmp := range obj.Streams {
		if streamTmp.runLock {
			streamTmp.signals <- SignalStreamStop
		}
	
	}
}

// NewHLSMuxer new muxer init
func (obj *StorageST) NewHLSMuxer(uuid string ) {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
		 
			streamTmp.hlsMuxer = nil
			 
			obj.Streams[uuid] = streamTmp
		 
	}
}

// HlsMuxerSetFPS write packet
func (obj *StorageST) HlsMuxerSetFPS(uuid string,  fps int) {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
		if streamTmp.hlsMuxer != nil {
			streamTmp.hlsMuxer.SetFPS(fps)
			obj.Streams[uuid] = streamTmp
		}
	}
}

// HlsMuxerWritePacket write packet
func (obj *StorageST) HlsMuxerWritePacket(uuid string, packet *av.Packet) {
	obj.mutex.RLock()
	defer obj.mutex.RUnlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
		if  streamTmp.hlsMuxer != nil {
			streamTmp.hlsMuxer.WritePacket(packet)
		}
	}
}

// HLSMuxerClose close muxer
func (obj *StorageST) HLSMuxerClose(uuid string ) {
	obj.mutex.RLock()
	defer obj.mutex.RUnlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
		 
			streamTmp.hlsMuxer.Close()
		 
	}
}

// HLSMuxerM3U8 get m3u8 list
func (obj *StorageST) HLSMuxerM3U8(uuid string,  msn, part int) (string, error) {
	obj.mutex.Lock()
	streamTmp, ok := obj.Streams[uuid]
	obj.mutex.Unlock()
	if ok {
		 
			index, err := streamTmp.hlsMuxer.GetIndexM3u8(msn, part)
			return index, err
		 
	}
	return "", ErrorStreamNotFound
}

// HLSMuxerSegment get segment
func (obj *StorageST) HLSMuxerSegment(uuid string,   segment int) ([]*av.Packet, error) {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
		 
			return streamTmp.hlsMuxer.GetSegment(segment)
		 
	}
	return nil, ErrorStreamChannelNotFound
}

// HLSMuxerFragment get fragment
func (obj *StorageST) HLSMuxerFragment(uuid string,  segment, fragment int) ([]*av.Packet, error) {
	obj.mutex.Lock()
	streamTmp, ok := obj.Streams[uuid]
	obj.mutex.Unlock()
	if ok {
		 
			packet, err := streamTmp.hlsMuxer.GetFragment(segment, fragment)
			return packet, err
		 
	}
	return nil, ErrorStreamChannelNotFound
}
 
 