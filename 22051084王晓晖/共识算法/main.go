package main

import (
	"github.com/pipapa/pbft/cmd"
	"github.com/pipapa/pbft/node"
	"time"
)

func main() {
	config := cmd.ReadConfig()
	n := node.NewNode(config)
	n.Run()
	for {
		t := time.After(time.Second)
		<-t
	}
}