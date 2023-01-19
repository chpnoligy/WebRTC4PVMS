package main

import (
	 
	"fmt"
	"time"

	//"mymodule/mypackage"

	"github.com/deepch/vdk/av"
)

 
//ClientAdd Add New Client to Translations
func (obj *StorageST) ClientAdd(streamID string, mode int) (string, chan *av.Packet, chan *[]byte, error) {
	//fmt.Println("ClientAdd" )

	obj.mutex.Lock()
	defer obj.mutex.Unlock()

	streamTmp, ok := obj.Streams[streamID]
	if !ok {
		return "", nil, nil, ErrorStreamNotFound
	}

	//Generate UUID client
	cid, err :=  GenerateUUID()
	if err != nil {
		return "", nil, nil, err
	}
	chAV := make(chan *av.Packet, 2000)
	chRTP := make(chan *[]byte, 2000)
	 
	streamTmp.clients[cid] = ClientST{mode: mode, outgoingAVPacket: chAV,
		                                    outgoingRTPPacket: chRTP, signals: make(chan int, 100)}
	streamTmp.ack = time.Now()
	 
	obj.Streams[streamID] = streamTmp
	return cid, chAV, chRTP, nil

}

//ClientDelete Delete Client
func (obj *StorageST) ClientDelete(streamID string, cid string) {
	//fmt.Println("ClientDelete " )
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if _, ok := obj.Streams[streamID]; ok {
		delete(obj.Streams[streamID].clients, cid)
	}
}

//ClientHas check is client ext
func (obj *StorageST) ClientHas(streamID string) bool {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	streamTmp, ok := obj.Streams[streamID]
	
	if !ok {
		return false
	}
	//if time.Now().Sub(channelTmp.ack).Seconds() > 30 {
	if time.Since(streamTmp.ack).Seconds() > 30 {
		fmt.Println("ClientHas Seconds() > 30 " )
		return false
	}
	return true
}

 