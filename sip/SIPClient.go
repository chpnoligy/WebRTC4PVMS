package sip

import (
	"fmt"
	"net"
	"strings"
	"time"

	//"github.com/cloudwebrtc/go-sip-ua/examples/mock"

	"github.com/cloudwebrtc/go-sip-ua/examples/mock"
	"github.com/cloudwebrtc/go-sip-ua/pkg/account"
	"github.com/cloudwebrtc/go-sip-ua/pkg/media/rtp"
	"github.com/cloudwebrtc/go-sip-ua/pkg/session"
	"github.com/cloudwebrtc/go-sip-ua/pkg/stack"
	"github.com/cloudwebrtc/go-sip-ua/pkg/ua"
	"github.com/cloudwebrtc/go-sip-ua/pkg/utils"

	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/sip/parser"
)

var (
	logger log.Logger	 	 
	udp    *rtp.RtpUDPStream
	SIP_UDP_Port    int
	SIP_TCP_Port    int
	SIP_WSS_Port    int
	SIP_User     	string
	SIP_Password    string
	SIP_Server    	string
	SIP_Transport   string
	SIP_Realm    	string
 	useragent   	*ua.UserAgent
	register 		*ua.Register

	remote_addr 	net.Addr
	profile 		*account.Profile
)

 

func createUdp() *rtp.RtpUDPStream {

fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!createUdp!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			//udp = rtp.NewRtpUDPStream("127.0.0.1", rtp.DefaultPortMin, rtp.DefaultPortMax, func(data []byte, raddr net.Addr) {
	
		udp = rtp.NewRtpUDPStream("192.168.178.209", 
		rtp.DefaultPortMin, 
		rtp.DefaultPortMax, func(data []byte, raddr net.Addr) {
			// er komt data binnen
			remote_addr = raddr
			//fmt.Println("rtp.NewRtpUDPStream raddr:", raddr )
			//logger.Infof("Rtp recevied: %v, laddr %s : raddr %s", len(data), udp.LocalAddr().String(), raddr)
			dest, _ := net.ResolveUDPAddr(raddr.Network(), raddr.String())
			//logger.Infof("Echo rtp to %v", raddr)
			//fmt.Println("rtp.NewRtpUDPStream dest:", dest )
			udp.Send(data, dest)
		 
	})
	 
	go udp.Read()  //thread 

	return udp
}


func Send_To_UPD(data []byte)  {

	if(remote_addr != nil){

	 
	//dest, _ := net.ResolveUDPAddr(raddr.Network(), raddr.String())
	//logger.Infof("Echo rtp to %v", raddr)
	dest, _ := net.ResolveUDPAddr(remote_addr.Network(), remote_addr.String())
	udp.Send(data, dest)
	fmt.Printf("Send_To_UPD %s %02x %02x %02x %02x\n", dest, data[0], data[1], data[2], data[3])
	}	    

}



func SIP_Register() {

	fmt.Println("!!!!!!Server SIP Client Register")
	 //
	// return
	//logger = utils.NewLogrusLogger(log.InfoLevel, "Client", nil)
	logger = utils.NewLogrusLogger(log.InfoLevel, "Client", nil)
 
	stack := stack.NewSipStack(&stack.SipStackConfig{
		UserAgent:  "Sip PVMS Client",
		Extensions: []string{"replaces", "outbound"},
		Dns:        "8.8.8.8"})
	
	fmt.Println("UDP:", SIP_UDP_Port, " TCP:", SIP_TCP_Port," WSS:", SIP_WSS_Port)

	if(SIP_UDP_Port == 0 && SIP_TCP_Port == 0 && SIP_WSS_Port == 0 ){
		return
	}
	var str string
	if(SIP_UDP_Port != 0){
		str = fmt.Sprintf("0.0.0.0:%d", SIP_UDP_Port)
		logger.Infof("UDP Listen => %s", str)
		if err := stack.Listen("udp", str); err != nil {
			logger.Panic(err)
		}
	}
	if(SIP_TCP_Port != 0){
		str = fmt.Sprintf("0.0.0.0:%d", SIP_TCP_Port)
		logger.Infof("TCP Listen => %s", str)
		if err := stack.Listen("tcp", str); err != nil {
			logger.Panic(err)
		}
	}
	if(SIP_WSS_Port != 0){
		str = fmt.Sprintf("0.0.0.0:%d", SIP_WSS_Port)
		logger.Infof("WSS Listen => %s", str)
		if err := stack.ListenTLS("wss", str, nil); err != nil {
			logger.Panic(err)
		}
	}

	//stack.OnConnectionError(func(err *transport.ConnectionError){

	//})
	//useragent aanmaken
	useragent = ua.NewUserAgent(&ua.UserAgentConfig{
		SipStack: stack,
	})


	//callback voor een INVITE message
	useragent.InviteStateHandler = func(sess *session.Session, req *sip.Request, resp *sip.Response, state session.Status) {
		//logger.Infof("InviteStateHandler: state => %v, type => %s", state, sess.Direction())
// callback als er een opreop/invite binnenkomt doen
		switch state {
		case session.InviteReceived:
			fmt.Println("session.InviteReceived")
			udp = createUdp()
			udpLaddr := udp.LocalAddr()
			sdp := mock.BuildLocalSdp(udpLaddr.IP.String(), udpLaddr.Port)
			
			sess.ProvideAnswer(sdp)
			sess.Accept(200)
		case session.Canceled:
			fallthrough

		case session.Failure:
			fallthrough
		case session.Terminated:
			udp.Close()
		}
	}

	//logger.Infof("SIPInit.go RegisterStateHandler ")
	//callback voor een REGISTER status message
	useragent.RegisterStateHandler = func(state account.RegisterState) {
		logger.Infof("RegisterStateHandler: user => %s, state => %v, expires => %v", state.Account.AuthInfo.AuthUser, state.StatusCode, state.Expiration)
	
	}
	
	//uri, err := parser.ParseUri("sip:100@1.2.3.4")
	str = fmt.Sprintf("sip:%s@%s",SIP_User,SIP_Server)
	uri, err := parser.ParseUri(str)
	if err != nil {
		logger.Error(err)
	}

	profile = account.NewProfile(uri.Clone(), "SIP/PVMS client",
		&account.AuthInfo{
			AuthUser: SIP_User,
			Password: SIP_Password,
			Realm:    SIP_Realm,
		},
		1800,
		stack,
	)

	switch(strings.ToLower(SIP_Transport)){
	case "udp":
		str = fmt.Sprintf("sip:%s@%s:%d;transport=%s", SIP_User, SIP_Server, SIP_UDP_Port,SIP_Transport)
		break
	case "tcp":
		str = fmt.Sprintf("sip:%s@%s:%d;transport=%s", SIP_User, SIP_Server, SIP_TCP_Port,SIP_Transport)
		break
	case "wss":
		str = fmt.Sprintf("sip:%s@%s:%d;transport=%s", SIP_User, SIP_Server, SIP_WSS_Port,SIP_Transport)
		break
	default:
		logger.Panic("Error SIP Transport ")
		break
	}

	logger.Infof("ParseSipUri: %s", str)
	//recipient, err := parser.ParseSipUri("sip:100@192.168.178.210:5060;transport=udp")
	recipient, err := parser.ParseSipUri(str)
	if err != nil {
		logger.Error(err)
	}

	register, _ = useragent.SendRegister(profile, recipient, profile.Expires, nil)
	
	 
	 
	time.Sleep(time.Second * 3)


	//SIP_Invite()

	//return
 
    //invite maken
	udp = createUdp()
	udpLaddr := udp.LocalAddr()
	sdp := mock.BuildLocalSdp(udpLaddr.IP.String(), udpLaddr.Port)

	called, err2 := parser.ParseUri("sip:101@192.168.178.210")
	if err2 != nil {
		logger.Error(err)
	}

	recipient, err = parser.ParseSipUri("sip:101@192.168.178.210:5060;transport=udp")
	if err != nil {
		logger.Error(err)
	}

	//fmt.Println("cd:",called, " sdp:", sdp)

	//time.Sleep(time.Second * 10)

	fmt.Println("useragent.Invite ")
	go useragent.Invite(profile, called, recipient, &sdp)
	fmt.Println("useragent.Invite done")
	 
 /*
	//<-stop


	fmt.Println("register.SendRegister(0)!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	register.SendRegister(0)

	ua.Shutdown()
	*/
}

func SIP_DisConnect() {
	 
	fmt.Println("Server SIP Client Stoppen")
	if register != nil {
		register.SendRegister(0)
	}
	if useragent != nil {
		useragent.Shutdown()
	}


}

func SIP_Invite(){
	
	fmt.Println("!!!!!!Server SIP Client Invite")
	udp = createUdp()   //om wat te ontvangen
	udpLaddr := udp.LocalAddr()
	sdp := mock.BuildLocalSdp(udpLaddr.IP.String(), udpLaddr.Port)

	var str string

	switch(strings.ToLower(SIP_Transport)){
	case "udp":
		str = fmt.Sprintf("sip:%s@%s:%d;transport=%s", SIP_User, SIP_Server, SIP_UDP_Port,SIP_Transport)
		break
	case "tcp":
		str = fmt.Sprintf("sip:%s@%s:%d;transport=%s", SIP_User, SIP_Server, SIP_TCP_Port,SIP_Transport)
		break
	case "wss":
		str = fmt.Sprintf("sip:%s@%s:%d;transport=%s", SIP_User, SIP_Server, SIP_WSS_Port,SIP_Transport)
		break
	default:
		logger.Panic("Error SIP Transport ")
		break
	}
	 
	  
	 
	called, err2 := parser.ParseUri("sip:101@192.168.178.210")
	if err2 != nil {
		fmt.Println("parser.ParseUri(str)" ,str, "error:", err2.Error())
	}

	recipient, err := parser.ParseSipUri("sip:101@192.168.178.210:5060;transport=udp")
	if err != nil {
		fmt.Println("parser.ParseSipUri(str)" ,str, "error:", err.Error())
	}


	fmt.Println("useragent.Invite ")
	go useragent.Invite(profile, called, recipient, &sdp)
	
	fmt.Println("useragent.Invite done")

}