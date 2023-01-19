package main

import (
	 
	 
)

//StreamsList list all stream


func (obj *StorageST) StreamsList() map[string]StreamST {
	obj.mutex.RLock()
	defer obj.mutex.RUnlock()
	tmp := make(map[string]StreamST)
	for i, i2 := range obj.Streams {
		tmp[i] = i2
	}
	return tmp
}

//SStreamAdd add stream
func (obj *StorageST) SStreamAdd(uuid string, strm StreamST) error {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	//TODO create empty map bug save https://github.com/liip/sheriff empty not nil map[] != {} json
	//data, err := sheriff.Marshal(&sheriff.Options{
	//		Groups:     []string{"config"},
	//		ApiVersion: v2,
	//	}, obj)
	//Not Work map[] != {}
	if obj.Streams == nil {
		obj.Streams = make(map[string]StreamST)
	}
	if _, ok := obj.Streams[uuid]; ok {
		return ErrorStreamAlreadyExists
	}
	strm = obj.StreamMake(strm)
	if !strm.OnDemand { // een niet ondemand stream meteen opstarten
		strm.runLock = true
		go StreamServerRunStreamDo(uuid)
	}
	//fmt.Println("SStreamAdd 1:", len(obj.Streams))
	obj.Streams[uuid] = strm
	//fmt.Println("SStreamAdd 2:", len(obj.Streams))
	err := obj.SaveConfig()
	if err != nil {
		return err
	}
	return nil
}

//StreamEdit edit stream
func (obj *StorageST) StreamEdit(uuid string, strm StreamST) error {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
		
			if streamTmp.runLock {
				streamTmp.signals <- SignalStreamStop
			}
		 
			strm := obj.StreamMake(strm)
			if !strm.OnDemand {
				strm.runLock = true 
				go StreamServerRunStreamDo(uuid)
			} 
		 
		obj.Streams[uuid] = strm
		err := obj.SaveConfig()
		if err != nil {
			return err
		}
		return nil
	}
	return ErrorStreamNotFound
}

 
 

//StreamReload reload stream
func (obj *StorageST) AStreamReload(uuid string) error {
	obj.mutex.RLock()
	defer obj.mutex.RUnlock()
	if streamTmp, ok := obj.Streams[uuid]; ok {
			if streamTmp.runLock {
				streamTmp.signals <- SignalStreamRestart
			}
		return nil
	}
	return ErrorStreamNotFound
}

//StreamDelete stream
func (obj *StorageST) StreamDelete(uuid string) error {
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

//StreamInfo return stream info
func (obj *StorageST) AStreamInfo(uuid string) (*StreamST, error) {
	obj.mutex.RLock()
	defer obj.mutex.RUnlock()
	if tmp, ok := obj.Streams[uuid]; ok {
		return &tmp, nil
	}
	return nil, ErrorStreamNotFound
}
