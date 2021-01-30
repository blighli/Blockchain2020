package node

import (
	"github.com/pipapa/pbft/message"
)

func (n *Node) commitRecvThread() {
	for {
		select {
		case msg := <-n.commit:
			if !n.checkCommitMsg(msg) {
				continue
			}
			n.buffer.BufferCommitMsg(msg)
			if n.buffer.IsReadyToExecute(msg.Digest, n.cfg.FaultNum, msg.View, msg.Sequence) {
				n.buffer.AppendToExecuteQueue(msg)
				n.executeNotify <- true
			}
		}
	}
}

func (n *Node) checkCommitMsg(msg *message.Commit) bool {
	if n.view != msg.View {
		return false
	}
	if !n.sequence.CheckBound(msg.Sequence) {
		return false
	}
	return true
}
