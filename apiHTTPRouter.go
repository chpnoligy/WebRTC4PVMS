package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"

)

//Message resp struct
type Message struct {
	Status  int         `json:"status"`
	Payload interface{} `json:"payload"`
}

//HTTPAPIServer start http server routes
func HTTPAPIServer() {

	log.Println("Server HTTP Server start")
	
	var public *gin.Engine

	gin.SetMode(gin.ReleaseMode)
	public = gin.New()

	public.Use(CrossOrigin())
	//Add private login password protect methods
	privat := public.Group("/")
	if Storage.ServerHTTPLogin() != "" && Storage.ServerHTTPPassword() != "" {
		log.Println("Server HTTP Server set username", Storage.ServerHTTPLogin(), " password:",Storage.ServerHTTPPassword())
		privat.Use(gin.BasicAuth(gin.Accounts{Storage.ServerHTTPLogin(): Storage.ServerHTTPPassword()}))
	}

	/*
		Static HTML Files 
	 

	//if Storage.ServerHTTPDemo() {
		public.LoadHTMLGlob(Storage.ServerHTTPDir() + "/templates/*")
		public.GET("/", HTTPAPIServerIndex)
		public.GET("/pages/stream/list", HTTPAPIStreamList)
		public.GET("/pages/stream/add", HTTPAPIAddStream)
		public.GET("/pages/stream/edit/:uuid", HTTPAPIEditStream)
		public.GET("/pages/player/hls/:uuid/:channel", HTTPAPIPlayHls)
		public.GET("/pages/player/mse/:uuid/:channel", HTTPAPIPlayMse)
		public.GET("/pages/player/webrtc/:uuid/:channel", HTTPAPIPlayWebrtc)
		public.StaticFS("/static", http.Dir(Storage.ServerHTTPDir()+"/static"))
	//}
*/
        public.LoadHTMLGlob(Storage.ServerHTTPDir() + "/templates/*")
		public.StaticFS("/static", http.Dir(Storage.ServerHTTPDir()+"/static"))

		public.GET("/", HTTPAPIServerIndex)
		privat.POST("/streams", HTTPAPIServerStreams)
		privat.POST("/stream/:uuid/add", HTTPAPIServerStreamAdd)
		privat.POST("/stream/:uuid/edit", HTTPAPIServerStreamEdit)
		privat.GET("/stream/:uuid/delete", HTTPAPIServerStreamDelete)

		public.POST("/stream/:uuid/webrtc", HTTPAPIServerStreamWebRTC_NOWS)
	/*
		Stream Control elements
	 

	privat.GET("/streams", HTTPAPIServerStreams)
	privat.POST("/stream/:uuid/add", HTTPAPIServerStreamAdd)
	privat.POST("/stream/:uuid/edit", HTTPAPIServerStreamEdit)
	privat.GET("/stream/:uuid/delete", HTTPAPIServerStreamDelete)
	privat.GET("/stream/:uuid/reload", HTTPAPIServerStreamReload)
	privat.GET("/stream/:uuid/info", HTTPAPIServerStreamInfo)
*/

	/*
		Stream video elements
	*/
	//HLS
	//public.GET("/stream/:uuid/channel/:channel/hls/live/index.m3u8", HTTPAPIServerStreamHLSM3U8)
	//public.GET("/stream/:uuid/channel/:channel/hls/live/segment/:seq/file.ts", HTTPAPIServerStreamHLSTS)
	//HLS LL
	//public.GET("/stream/:uuid/channel/:channel/hlsll/live/index.m3u8", HTTPAPIServerStreamHLSLLM3U8)
	//public.GET("/stream/:uuid/channel/:channel/hlsll/live/init.mp4", HTTPAPIServerStreamHLSLLInit)
	//public.GET("/stream/:uuid/channel/:channel/hlsll/live/segment/:segment/:any", HTTPAPIServerStreamHLSLLM4Segment)
	//public.GET("/stream/:uuid/channel/:channel/hlsll/live/fragment/:segment/:fragment/:any", HTTPAPIServerStreamHLSLLM4Fragment)
	//MSE
	//public.GET("/stream/:uuid/channel/:channel/mse", HTTPAPIServerStreamMSE)
	//WEBRTC
	//public.POST("/stream/:uuid/channel/:channel/webrtc", HTTPAPIServerStreamWebRTC)

	/*
		HTTPS Mode Cert
		# Key considerations for algorithm "RSA" ≥ 2048-bit
		openssl genrsa -out server.key 2048

		# Key considerations for algorithm "ECDSA" ≥ secp384r1
		# List ECDSA the supported curves (openssl ecparam -list_curves)
		#openssl ecparam -genkey -name secp384r1 -out server.key
		#Generation of self-signed(x509) public key (PEM-encodings .pem|.crt) based on the private (.key)

		openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
	*/
	if Storage.ServerHTTPS() {  //Https
		log.Println("Server HTTPS listen port:",Storage.ServerHTTPSPort())
		if Storage.ServerHTTPSAutoTLSEnable() {
			go func() {
				err := autotls.Run(public, Storage.ServerHTTPSAutoTLSName()+Storage.ServerHTTPSPort())
				if err != nil {
					log.Println("Start HTTPS ServerHTTPSAutoTLSEnable Error", err.Error())
				}
			}()
		} else {
			go func() {
				err := public.RunTLS(Storage.ServerHTTPSPort(), Storage.ServerHTTPSCert(), Storage.ServerHTTPSKey())
				if err != nil {
					log.Println("Start HTTPS ServerHTTPSPort Error", err.Error())
					os.Exit(1)
				}
			}()
		}
	}

	log.Println("Server HTTP listen port:",Storage.ServerHTTPPort() , " Webroot:", Storage.ServerHTTPDir() )
	err := public.Run(Storage.ServerHTTPPort())
	if err != nil {
		log.Println("Start HTTP ServerHTTPPort Error", err.Error())
		os.Exit(1)
	}

}
 
func HTTPAPIServerIndex(c *gin.Context) { //index pagina
	//fmt.Println("HTTPAPIServerIndex" )
	c.HTML(http.StatusOK, "index.html", gin.H{
		"version": time.Now().String(),
		"streams": Storage.Streams,
		"title": "ControlBase 2.0",
	})
}

func HTTPAPIServerStreams(c *gin.Context) {
	//fmt.Println("HTTPAPIServerStreams" )
	c.IndentedJSON(200, Message{Status: 1, Payload: Storage.StreamsList()})
}

//HTTPAPIServerStreamAdd function add new stream
func HTTPAPIServerStreamAdd(c *gin.Context) {
	fmt.Println("HTTPAPIServerStreamAdd" )
	
	var payload StreamST
	err := c.BindJSON(&payload)
	if err != nil {
		c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
		log.Println("HTTPAPIServerStreamAdd", err.Error())
		return
	}
	err = Storage.SStreamAdd(c.Param("uuid"), payload)
	if err != nil {
		c.IndentedJSON(500, Message{Status: 0, Payload: err.Error()})
		log.Println("HTTPAPIServerStreamAdd", err.Error())
		return
	}
	fmt.Println("HTTPAPIServerStreamAdd payload:", payload )
	c.IndentedJSON(200, Message{Status: 1, Payload: Success})
}

//HTTPAPIServerStreamEdit function edit stream
func HTTPAPIServerStreamEdit(c *gin.Context) {
	fmt.Println("HTTPAPIServerStreamEdit" )
	var payload StreamST
	err := c.BindJSON(&payload)
	if err != nil {
		c.IndentedJSON(400, Message{Status: 0, Payload: err.Error()})
		log.Println("HTTPAPIServerStreamEdit", err.Error())
		return
	}
	err = Storage.StreamEdit(c.Param("uuid"), payload)
	if err != nil {
		c.IndentedJSON(500, Message{Status: 0, Payload: err.Error()})
		log.Println("HTTPAPIServerStreamEdit", err.Error())
		return
	}
	c.IndentedJSON(200, Message{Status: 1, Payload: Success})
}

//HTTPAPIServerStreamDelete function delete stream
func HTTPAPIServerStreamDelete(c *gin.Context) {
	fmt.Println("HTTPAPIServerStreamDelete" )
	err := Storage.StreamDelete(c.Param("uuid"))
	if err != nil {
		c.IndentedJSON(500, Message{Status: 0, Payload: err.Error()})
		log.Println("HTTPAPIServerStreamDelete", err.Error())
		return
	}
	c.IndentedJSON(200, Message{Status: 1, Payload: Success})
}





func HTTPAPIServerDocumentation(c *gin.Context) {
	c.HTML(http.StatusOK, "documentation.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "documentation",
	})
}

func HTTPAPIStreamList(c *gin.Context) {
	c.HTML(http.StatusOK, "stream_list.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "stream_list",
	})
}

func HTTPAPIPlayHls(c *gin.Context) {
	c.HTML(http.StatusOK, "play_hls.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "play_hls",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}
func HTTPAPIPlayMse(c *gin.Context) {
	c.HTML(http.StatusOK, "play_mse.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "play_mse",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}
func HTTPAPIPlayWebrtc(c *gin.Context) {
	fmt.Println("apiHTTPRouter HTTPAPIPlayWebrtc")
	c.HTML(http.StatusOK, "play_webrtc.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "play_webrtc",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}
func HTTPAPIAddStream(c *gin.Context) {
	c.HTML(http.StatusOK, "add_stream.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "add_stream",
	})
}
func HTTPAPIEditStream(c *gin.Context) {
	c.HTML(http.StatusOK, "edit_stream.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "edit_stream",
		"uuid":    c.Param("uuid"),
	})
}

func HTTPAPIMultiview(c *gin.Context) {
	c.HTML(http.StatusOK, "multiview.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "multiview",
	})
}

func HTTPAPIPlayAll(c *gin.Context) {
	c.HTML(http.StatusOK, "play_all.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "play_all",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}

type MultiViewOptions struct {
	Grid   int                             `json:"grid"`
	Player map[string]MultiViewOptionsGrid `json:"player"`
}
type MultiViewOptionsGrid struct {
	UUID       string `json:"uuid"`
	Channel    int    `json:"channel"`
	PlayerType string `json:"playerType"`
}

 

//CrossOrigin Access-Control-Allow-Origin any methods
func CrossOrigin() gin.HandlerFunc {
	 
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
