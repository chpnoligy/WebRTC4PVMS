package main

import (
	"sort"
	"strconv"
	"time"

	"github.com/deepch/vdk/av"
)

//StreamHLSAdd add hls seq to buffer
func (obj *StorageST) StreamHLSAdd(uuid string, val []*av.Packet, dur time.Duration) {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
		 
		streamTmp.hlsSegmentNumber++
		streamTmp.hlsSegmentBuffer[streamTmp.hlsSegmentNumber] = SegmentOld{data: val, dur: dur}
			if len(streamTmp.hlsSegmentBuffer) >= 6 {
				delete(streamTmp.hlsSegmentBuffer, streamTmp.hlsSegmentNumber-6-1)
			}
			obj.Streams[uuid] = streamTmp		 
	}
}

//StreamHLSm3u8 get hls m3u8 list
func (obj *StorageST) StreamHLSm3u8(uuid string, channelID string) (string, int, error) {
	obj.mutex.RLock()
	defer obj.mutex.RUnlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
		 
			var out string
			//TODO fix  it
			out += "#EXTM3U\r\n#EXT-X-TARGETDURATION:4\r\n#EXT-X-VERSION:4\r\n#EXT-X-MEDIA-SEQUENCE:" + strconv.Itoa(streamTmp.hlsSegmentNumber) + "\r\n"
			var keys []int
			for k := range streamTmp.hlsSegmentBuffer {
				keys = append(keys, k)
			}
			sort.Ints(keys)
			var count int
			for _, i := range keys {
				count++
				out += "#EXTINF:" + strconv.FormatFloat(streamTmp.hlsSegmentBuffer[i].dur.Seconds(), 'f', 1, 64) + ",\r\nsegment/" + strconv.Itoa(i) + "/file.ts\r\n"

			}
			return out, count, nil
		 
	}
	return "", 0, ErrorStreamNotFound
}

//StreamHLSTS send hls segment buffer to clients
func (obj *StorageST) StreamHLSTS(uuid string, channelID string, seq int) ([]*av.Packet, error) {
	obj.mutex.RLock()
	defer obj.mutex.RUnlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
			if tmp, ok := streamTmp.hlsSegmentBuffer[seq]; ok {
				return tmp.data, nil
			}
		 
	}
	return nil, ErrorStreamNotFound
}

//StreamHLSFlush delete hls cache
func (obj *StorageST) StreamHLSFlush(uuid string) {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
			streamTmp.hlsSegmentBuffer = make(map[int]SegmentOld)
			streamTmp.hlsSegmentNumber = 0
			obj.Streams[uuid] = streamTmp
	}
}
