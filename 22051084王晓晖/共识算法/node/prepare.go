package node

import (
	"github.com/pipapa/pbft/message"
	"github.com/pipapa/pbft/server"
)

func (n *Node) prepareRecvAndCommitSendThread() {
	for {
		select {
		case msg := <-n.prepareRecv:
			if !n.checkPrepareMsg(msg) {
				continue
			}
			n.buffer.BufferPrepareMsg(msg)
			if n.buffer.IsTrueOfPrepareMsg(msg.Digest, n.cfg.FaultNum) {
				content, msg, err := message.NewCommitMsg(n.id, msg)
				if err != nil {
					continue
				}
				n.buffer.BufferCommitMsg(msg)
				n.BroadCast(content, server.CommitEntry)
			}
		}
	}
}

func (n *Node) checkPrepareMsg(msg *message.Prepare) bool {
	if n.view != msg.View {
		return false
	}
	if !n.sequence.CheckBound(msg.Sequence) {
		return false
	}
	return true
}
