package node

import (
	"bytes"
	"log"
	"net/http"
)

func (n *Node) BroadCast(content []byte, handle string) {
	for k, v := range n.table {
		if k == n.id {
			continue
		}
		go SendPost(content, v+handle)
	}
}

func SendPost(content []byte, url string) {
	buff := bytes.NewBuffer(content)
	if _, err := http.Post(url, "application/json", buff); err != nil {
		log.Printf("[Send] send to %s error: %s", url, err)
	}
}
