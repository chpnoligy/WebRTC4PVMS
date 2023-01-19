package main

import (
	"errors"
 
	"sync"
	"time"
	 
	"github.com/deepch/vdk/av"
)

var Storage = NewStreamCore()

//Default stream  type
const (
	MSE = iota
	WEBRTC
	RTSP
)

//Default stream status type
const (
	OFFLINE = iota
	ONLINE
)

//Default stream errors
var (
	Success                         = "success"
	ErrorStreamNotFound             = errors.New("stream not found")
	ErrorStreamAlreadyExists        = errors.New("stream already exists")
	ErrorStreamChannelAlreadyExists = errors.New("stream channel already exists")
	ErrorStreamNotHLSSegments       = errors.New("stream hls not ts seq found")
	ErrorStreamNoVideo              = errors.New("stream no video")
	ErrorStreamNoClients            = errors.New("stream no clients")
	ErrorStreamRestart              = errors.New("stream restart")
	ErrorStreamStopCoreSignal       = errors.New("stream stop core signal")
	ErrorStreamStopRTSPSignal       = errors.New("stream stop rtsp signal")
	ErrorStreamChannelNotFound      = errors.New("stream channel not found")
	ErrorStreamChannelCodecNotFound = errors.New("stream channel codec not ready, possible stream offline")
	ErrorStreamsLen0                = errors.New("streams len zero")
)



//StorageST main storage struct
/*
type obsStorageST struct {
	mutex           sync.RWMutex
	Server          ServerST            `json:"server" groups:"api,config"`
	Streams         map[string]obsStreamST `json:"streams,omitempty" groups:"api,config"`
	//ChannelDefaults ChannelST           `json:"channel_defaults,omitempty" groups:"api,config"`
}
*/

type StorageST struct {
	mutex           sync.RWMutex
	Server          ServerST             `json:"server" groups:"api,config"`
	Streams         map[string]StreamST  `json:"streams,omitempty" groups:"api,config"`
	StreamDefaults   StreamST            `json:"channel_defaults,omitempty" groups:"api,config"`
}

//ServerST server storage section
type ServerST struct {
	Verbose           bool         		`json:"verbose" groups:"api,config"`
	WS_URL            string            `json:"websocket_url" groups:"api,config"`
	Site_RefID        string            `json:"site_refid" groups:"api,config"`
	Site_Desc         string            `json:"site_desc" groups:"api,config"`
	Item_RefID        string            `json:"item_refid" groups:"api,config"`
	Item_Desc         string            `json:"item_desc" groups:"api,config"`
	Facil_RefID        string           `json:"facility_refid" groups:"api,config"`
	Facil_Desc         	string          `json:"facility_desc" groups:"api,config"`
	SIP_UDP_Port       	int           	`json:"sip_udp_port" groups:"api,config"`
	SIP_TCP_Port       	int           	`json:"sip_tcp_port" groups:"api,config"`
	SIP_WSS_Port       	int           	`json:"sip_wss_port" groups:"api,config"`
	SIP_User      		string          `json:"sip_user" groups:"api,config"`
	SIP_Password       	string         	`json:"sip_password" groups:"api,config"`
	SIP_Server       	string         	`json:"sip_server" groups:"api,config"`
	SIP_Transport       string          `json:"sip_transport" groups:"api,config"`
	SIP_Realm       	string          `json:"sip_realm" groups:"api,config"`
	Debug              bool         `json:"debug" groups:"api,config"`
	HTTPDemo           bool         `json:"http_demo" groups:"api,config"`
	HTTPDebug          bool         `json:"http_debug" groups:"api,config"`
	HTTPUsername          string       `json:"http_username" groups:"api,config"`
	HTTPPassword       string       `json:"http_password" groups:"api,config"`
	HTTPDir            string       `json:"http_dir" groups:"api,config"`
	HTTPPort           string       `json:"http_port" groups:"api,config"`
	RTSPPort           string       `json:"rtsp_port" groups:"api,config"`
	HTTPS              bool         `json:"https" groups:"api,config"`
	HTTPSPort          string       `json:"https_port" groups:"api,config"`
	HTTPSCert          string       `json:"https_cert" groups:"api,config"`
	HTTPSKey           string       `json:"https_key" groups:"api,config"`
	HTTPSAutoTLSEnable bool         `json:"https_auto_tls" groups:"api,config"`
	HTTPSAutoTLSName   string       `json:"https_auto_tls_name" groups:"api,config"`
	ICEServers         []string     `json:"ice_servers" groups:"api,config"`
	ICEUsername        string       `json:"ice_username" groups:"api,config"`
	ICECredential      string       `json:"ice_credential" groups:"api,config"`
	Token              Token        `json:"token,omitempty" groups:"api,config"`
	WebRTCPortMin      uint16       `json:"webrtc_port_min" groups:"api,config"`
	WebRTCPortMax      uint16       `json:"webrtc_port_max" groups:"api,config"`
}

//Token auth
type Token struct {
	Enable  bool   `json:"enable" groups:"api,config"`
	Backend string `json:"backend" groups:"api,config"`
}


type StreamST struct {
	Name               string `json:"name,omitempty" groups:"api,config"`
	URL                string `json:"url,omitempty" groups:"api,config"`
	OnDemand           bool   `json:"ondemand,omitempty" groups:"api,config"`
	Onvif              bool   `json:"onvif,omitempty" groups:"api,config"`
	Username           string `json:"username,omitempty" groups:"api,config"`
	Password           string `json:"password,omitempty" groups:"api,config"`
	Debug              bool   `json:"debug,omitempty" groups:"api,config"`    
	AStatus            int    `json:"status,omitempty" groups:"api"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify,omitempty" groups:"api,config"` 
	Audio              bool   `json:"audio,omitempty" groups:"api,config"`
	runLock            bool
	codecs             []av.CodecData
	sdp                []byte
	signals            chan int
	hlsSegmentBuffer   map[int]SegmentOld
	hlsSegmentNumber   int
	clients            map[string]ClientST
	ack                time.Time
	hlsMuxer           *MuxerHLS `json:"-"`
}

//ClientST client storage section
type ClientST struct {
	mode              int
	signals           chan int
	outgoingAVPacket  chan *av.Packet //tbv WebRTC Client
	outgoingRTPPacket chan *[]byte  // tbv RTSP Client
	//socket            net.Conn
}

//SegmentOld HLS cache section
type SegmentOld struct {
	dur  time.Duration
	data []*av.Packet
}

 
 