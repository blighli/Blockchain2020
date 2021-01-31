package main

import (
	"fmt"
	"os"
	"time"

	gp "github.com/zballs/goPBFT/goPBFT"
)

func main() {
	ID := os.Args[1]
	if ID == "client" {
		client := gp.NewClient()
		go client.Startlisten()
		/*for {
			var msg string
			fmt.Scanln(&msg)
			if msg == "quit" { //在客户端命令行中输入quit退出，否则每次回车则发送一次请求
				break
			}
			client.SendRequest()
		}*/
		var msgnum int
		fmt.Scanln(&msgnum)
		for i := 0; i < msgnum; i++ {
			client.SendRequest()
			time.Sleep(100 * time.Millisecond)
		}

	} else {
		server := gp.NewServer(ID)
		server.Start()
	}
}