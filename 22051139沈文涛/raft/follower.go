package main

import (
	"math/rand"
	"time"
)

// 随机设置超时时间
func RandomTimeout(s *Server) bool {
	// Milliseconds, Min = timeout, max = timeout*2
	sleep := rand.Int()%(timeout*1000) + (timeout * 1000)

	for i := sleep; i > 0; i-- {
		// Every 100 milliseconds, check for an update
		if i%100 == 0 {
			// Check the channel for some input, but don't block on the channel
			select {
			case value := <-s.Hb:
				logger.Infof("%v: heartbeat to %v\n", s.ID, value)
				return true
			case <-s.VoteRequested:
				return true
			default:
				// Do nothing for now
			}
		} else {
			time.Sleep(time.Millisecond)
		}
	}
	return false
}

// 响应心跳信息
func (s *Server) Heartbeat(message *Message, response *Message) error {

	response.Source = s.Port
	response.Destination = message.Source
	response.Term = s.Term
	response.Index = len(s.Log)
	if message.NumServers != s.NumAliveServers {
		s.NumAliveServers = message.NumServers
		logger.Infof("Server Status updated\n")
		s.AliveServers = message.ServerStatus
	}
	// Flip bool to let client thread know we sent a heartbeat

	s.Hb <- message.SourceID
	return nil
}

// 响应对投票信息
func (s *Server) Elect(message *Message, response *Message) error {
	response.SourceID = s.ID
	response.Source = s.Port
	response.Destination = message.Source
	response.Term = s.Term
	response.Index = len(s.Log)

	// If you haven't already voted
	if s.Voted != -1 {
		response.Vote = false
	} else {
		// We've voted
		response.Vote = true
		s.VoteRequested <- true
	}
	return nil
}
