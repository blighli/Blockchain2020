package main

import (
	"fmt"
	"log"
	"net/rpc"
)

// 开始参与leader选举
func StartElection(s *Server) {
	s.Voted = s.ID
	s.TotalVotes[s.ID] = true
	logger.Infof("%v alive servers\n", s.NumAliveServers)
	for _, val := range s.Servers {
		if val.ID != s.ID && s.AliveServers[val.ID] {
			go RequestVote(s, val)
		}
	}

	for i := 0; i < s.NumAliveServers; i++ {
		<-s.VoteReceived
		if x := CheckVotes(s); x > s.NumAliveServers/2 {
			s.State = StateLeader
			return
		}
	}

}

// 检查投票
func CheckVotes(s *Server) int {
	votes := 0
	fmt.Println(s.TotalVotes)
	for _, val := range s.TotalVotes {
		if val {
			votes++
		}
	}
	return votes
}

// 发送投票请求
func RequestVote(source *Server, destination *Server) {
	var mes = new(Message)
	mes.Source = source.Port
	mes.Destination = destination.Port
	mes.Index = len(source.Log)
	mes.Term = source.Term

	// send response
	client, err := rpc.Dial("tcp", mes.Destination)
	if err != nil {
		logger.Errorf("Cannot connect to %v for vote\n", destination.ID)
		return
	}

	var reply = new(Message)
	reply.Vote = true
	err = client.Call("Server.Elect", mes, reply)
	if err != nil {
		log.Fatal(err)
	}

	if reply.Vote {
		logger.Infof("%v: yes vote from %v\n", source.ID, destination.ID)
		source.TotalVotes[destination.ID] = true
		source.VoteReceived <- true
	}
	logger.Infof("%v: no vote from %v\n", source.ID, destination.ID)
}
