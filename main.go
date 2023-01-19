package main

import (
	 
	"log"
	"os"
	"os/signal"
	"syscall"
	"WebRTC4PVMS/sip"
	  
)


//var Storage = NewStreamCore()

func main() {
	 
	//fmt.Println("Hello, Modules!")
	//mypackage.PrintHello()
	log.Println("Server Started V1.0")
	//var Storage StorageST
	//go Storage.StreamRunAll()
	go HTTPAPIServer()
	 
	//sip.SIP_Connect()
    sip.SIP_Register()
    WSClient_Connect( WS_Url)

	 

	signalChanel := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(signalChanel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalChanel

		log.Println("Server receive signal:", sig)
		done <- true
	}()
	
	log.Println("Server Start success a wait signals ")
	<-done
	
	sip.SIP_DisConnect()
 	Storage.StopAllStreams()

	WSCLient_Disconnect(); 
	log.Println("Server stopped working")

}