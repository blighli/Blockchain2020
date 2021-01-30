package node

import (
	"log"

	"github.com/pipapa/pbft/message"
)

func (n *Node) executeAndReplyThread() {
	for {
		select {
		case <-n.executeNotify:
			batchs, lastSeq := n.buffer.BatchExecute(n.sequence.GetLastSequence())
			n.sequence.SetLastSequence(lastSeq)
			requestBatchs := make([]*message.Request, 0)
			for _, b := range batchs {
				reqs := n.buffer.FetchRequest(b)
				requestBatchs = append(requestBatchs, reqs...)
			}
			for _, r := range requestBatchs {
				log.Printf("do the opreation %s - %d", r.Op, r.TimeStamp)
			}

		}
	}
}
