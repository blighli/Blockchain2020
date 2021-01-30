package main

import (
	"log"
	"pbft/pbft"
)

func main() {
	//为四个节点生成公私钥
	pbft.GenRsaKeys()
	pbft.NodeTable = map[string]string{
		"N0": "127.0.0.1:8000",
		"N1": "127.0.0.1:8001",
		"N2": "127.0.0.1:8002",
		"N3": "127.0.0.1:8003",
	}
	//首先启动四个节点
	nodeID := "client"
	//nodeID := "N0"
	//nodeID := "N1"
	//nodeID := "N2"
	//nodeID := "N3"
	if nodeID == "client" {
		pbft.ClientSendMessageAndListen() //启动客户端程序
	} else if addr, ok := pbft.NodeTable[nodeID]; ok {
		p := pbft.NewPBFT(nodeID, addr)
		go p.TcpListen() //启动节点
	} else {
		log.Fatal("无此节点编号！")
	}
	select {}
}
