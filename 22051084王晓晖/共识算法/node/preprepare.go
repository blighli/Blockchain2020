package node

import (
	"log"
	"time"

	"github.com/pipapa/pbft/message"
	"github.com/pipapa/pbft/server"
)

func (n *Node) prePrepareSendThread() {
	duration := time.Second
	timer := time.After(duration)
	for {
		select {
		case <-n.prePrepareSendNotify:
			n.prePrepareSendHandleFunc()
		case <-timer:
			timer = nil
			n.prePrepareSendHandleFunc()
			timer = time.After(duration)
		}
	}
}

func (n *Node) prePrepareSendHandleFunc() {
	if n.executeNum.Get() >= n.cfg.ExecuteMaxNum {
		return
	}
	if n.buffer.SizeofRequestQueue() < 1 {
		return
	}
	batch := n.buffer.BatchRequest()
	if len(batch) < 1 {
		return
	}
	seq := n.sequence.Get()
	content, msg, digest, err := message.NewPreprepareMsg(n.view, seq, batch)
	if err != nil {
		log.Printf("[PrePrepare] generate pre-prepare message error")
		return
	}
	log.Printf("[PrePrepare] generate sequence(%d) for msg(%s)", seq, digest[0:9])
	n.buffer.BufferPreprepareMsg(msg)
	n.BroadCast(content, server.PrePrepareEntry)
}

func (n *Node) prePrepareRecvAndPrepareSendThread() {
	for {
		select {
		case msg := <-n.prePrepareRecv:
			if !n.checkPrePrepareMsg(msg) {
				continue
			}
			n.buffer.BufferPreprepareMsg(msg)
			content, prepare, err := message.NewPrepareMsg(n.id, msg)
			if err != nil {
				continue
			}
			n.buffer.BufferPrepareMsg(prepare)
			n.BroadCast(content, server.PrepareEntry)
		}
	}
}

func (n *Node) checkPrePrepareMsg(msg *message.PrePrepare) bool {
	if n.view != msg.View {
		return false
	}
	if n.buffer.IsExistPreprepareMsg(msg.View, msg.Sequence) {
		return false
	}
	d, err := message.Digest(msg.Message)
	if err != nil {
		return false
	}
	if d != msg.Digest {
		return false
	}
	if !n.sequence.CheckBound(msg.Sequence) {
		return false
	}
	return true
}
