package node

import (
	"log"
)

func (n *Node) requestRecvThread() {
	log.Printf("[Node] start recv the request thread")
	for {
		msg := <-n.requestRecv
		if !n.IsPrimary() {
			if n.lastReply.Equal(msg) {
			} else {
			}
		}
		n.buffer.AppendToRequestQueue(msg)
		n.prePrepareSendNotify <- true
	}
}
