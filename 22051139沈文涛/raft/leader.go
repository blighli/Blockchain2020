package main

import (
	"log"
	"net/rpc"
)

// 发送心跳消息
func SendHeartbeatRequest(source *Server, destination *Server) {
	var mes = new(Message)
	mes.SourceID = source.ID
	mes.Source = source.Port
	mes.Destination = destination.Port
	mes.Index = len(source.Log)
	mes.Term = source.Term
	mes.NumServers = source.NumAliveServers
	mes.ServerStatus = source.AliveServers

	client, err := rpc.Dial("tcp", mes.Destination)
	if err != nil {
		logger.Errorf("No response from %v\n", destination.ID)
		if source.AliveServers[destination.ID] == true {
			source.AliveServers[destination.ID] = false
			source.NumAliveServers--
		}
		return
	}

	var reply = new(Message)
	err = client.Call("Server.Heartbeat", mes, reply)
	if err != nil {
		log.Print(err)
	}

	logger.Infof("Heartbeat from %v, Epoch %v, Index %v, Num Servers %v\n", destination.ID, reply.Term, reply.Index, source.NumAliveServers)
}

// 发送心跳信息给其它节点
func SendHeartbeats(s *Server) {
	for _, val := range s.Servers {
		if val.ID != s.ID { // 跳过当前节点
			go SendHeartbeatRequest(s, val)
		}
	}
}
