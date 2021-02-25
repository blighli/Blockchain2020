package main

import (
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"time"
)

// 代表集群中的节点角色
type StateType int

const (
	StateFollower StateType = iota
	StateCandidate
	StateLeader
)

// 节点有以下三个状态:
//   0. Follower
//   1. Candidate
//   2. Leader
type Server struct {
	ID              int
	Term            int   // 任期
	Log             []Log // 日志
	Port            string
	State           StateType // 节点状态
	Servers         []*Server
	AliveServers    []bool
	Hb              chan int
	VoteRequested   chan bool
	VoteReceived    chan bool
	Voted           int
	TotalVotes      []bool
	NumAliveServers int
}

// 模拟创建创建一个raft节点
func CreateServer(id int, port string, startState int) *Server {
	server := new(Server)
	server.ID = id
	server.Term = 0
	server.Port = port
	server.State = StateType(startState)
	server.Hb = make(chan int)
	server.Voted = -1
	server.VoteRequested = make(chan bool)
	server.VoteReceived = make(chan bool)
	server.TotalVotes = []bool{false, false, false, false, false}
	server.AliveServers = []bool{true, true, true, true, true}
	server.NumAliveServers = numServers
	return server
}

// 启动raft节点
func Run(s *Server) {
	// Seed random for timeout
	rand.Seed(time.Now().UTC().UnixNano())

	go func() {
		address, err := net.ResolveTCPAddr("tcp", "127.0.0.1"+s.Port)
		if err != nil {
			log.Fatal(err)
		}

		inbound, err := net.ListenTCP("tcp", address)
		if err != nil {
			log.Fatal(err)
		}

		rpc.Register(s)
		rpc.Accept(inbound)
	}()

	for {
		switch s.State {
		case StateFollower:
			// Wait for heartbeat request
			if !RandomTimeout(s) {
				// Switch State
				s.State = StateCandidate
			}
		case StateCandidate:
			select {
			case <-s.VoteRequested:
				s.State = StateFollower
			default:
				// Start an election
				logger.Infof("%v: election started\n", s.ID)
				StartElection(s)
			}
		case StateLeader:
			select {
			case <-s.VoteRequested:
				s.Voted = -1
				s.TotalVotes = []bool{false, false, false, false, false}
				// If a vote is request while leader, surrender leadership
				s.State = StateFollower
			default:
				logger.Infof("%v: is leader\n", s.ID)
				time.Sleep(time.Second)
				// 发送心跳信息给其它节点
				SendHeartbeats(s)
			}

		}
	}
}
