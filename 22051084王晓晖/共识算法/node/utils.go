package node

import (
	"sync"

	"github.com/pipapa/pbft/message"
)

type ExecuteOpNum struct {
	num    int
	locker *sync.RWMutex
}

func NewExecuteOpNum() *ExecuteOpNum {
	return &ExecuteOpNum{
		num:    0,
		locker: new(sync.RWMutex),
	}
}

func (n *ExecuteOpNum) Get() int {
	n.locker.RLock()
	ret := n.num
	n.locker.RUnlock()
	return ret
}

func (n *ExecuteOpNum) Add(i int) {
	n.locker.Lock()
	n.num = n.num + i
	n.locker.Unlock()
}

func (n *Node) GetPrimary() message.Identify {
	all := len(n.table)
	return message.Identify(int(n.view) % all)
}

func (n *Node) IsPrimary() bool {
	p := n.GetPrimary()
	if p == message.Identify(n.view) {
		return true
	}
	return false
}
