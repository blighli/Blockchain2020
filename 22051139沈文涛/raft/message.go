package main

type Message struct {
	SourceID     int
	Source       string
	Destination  string
	Index        int
	Term         int
	Vote         bool
	NumServers   int
	ServerStatus []bool
}
