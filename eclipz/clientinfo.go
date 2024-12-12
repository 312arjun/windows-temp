package eclipz

func sendClientInfo() {
	req := &RpcClientInfo{}
	req.Cmd = CmdClientInfo
	req.MsgId = getNextMsgId()
	req.Token = Client.AuthToken

	req.WgKey = WgGetMyPublicKey()
	req.WgPort = WgListenPort
	req.PublicIP = myPublicAddress()
	req.LocalIPs = myLocalAddresses()

	sendQueue<-req
}
