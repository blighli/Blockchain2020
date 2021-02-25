package main

import (
	"flag"
	"raft/logging"
)

const (
	numServers = 5
	timeout    = 5
)

var logger = logging.MustGetLogger("raft.main")

func main() {
	id := flag.Int("id", 0, "Usage: -id <id>")
	port := flag.String("port", ":8081", "Usage: -port <portnumber>")
	state := flag.Int("state", 0, "Usage: -state <start-state>")
	flag.Parse()

	server := CreateServer(*id, *port, *state)

	// 使程序保持活动
	exit := make(chan bool)

	// 跟踪所有已声明服务器的状态
	server0 := CreateServer(0, ":8081", 0)
	server1 := CreateServer(1, ":8082", 0)
	server2 := CreateServer(2, ":8083", 0)
	server3 := CreateServer(3, ":8084", 0)
	server4 := CreateServer(4, ":8085", 0)

	servers := []*Server{server0, server1, server2, server3, server4}
	server.Servers = servers

	go Run(server)

	<-exit
}
