package node

import (
	"log"

	"github.com/pipapa/pbft/cmd"
	"github.com/pipapa/pbft/message"
	"github.com/pipapa/pbft/server"
)

type Node struct {
	cfg    *cmd.SharedConfig
	server *server.HttpServer

	id       message.Identify
	view     message.View
	table    map[message.Identify]string
	faultNum uint

	lastReply  *message.LastReply
	sequence   *Sequence
	executeNum *ExecuteOpNum

	buffer *message.Buffer

	requestRecv    chan *message.Request
	prePrepareRecv chan *message.PrePrepare
	prepareRecv    chan *message.Prepare
	commit         chan *message.Commit

	prePrepareSendNotify chan bool
	executeNotify        chan bool
}

func NewNode(cfg *cmd.SharedConfig) *Node {
	node := &Node{
		cfg:                  cfg,
		server:               server.NewServer(cfg),
		id:                   cfg.Id,
		view:                 cfg.View,
		table:                cfg.Table,
		faultNum:             cfg.FaultNum,
		lastReply:            message.NewLastReply(),
		sequence:             NewSequence(cfg),
		executeNum:           NewExecuteOpNum(),
		buffer:               message.NewBuffer(),
		requestRecv:          make(chan *message.Request),
		prePrepareRecv:       make(chan *message.PrePrepare),
		prepareRecv:          make(chan *message.Prepare),
		commit:               make(chan *message.Commit),
		prePrepareSendNotify: make(chan bool),
		executeNotify:        make(chan bool),
	}
	log.Printf("[Node] the node id:%d, view:%d, fault number:%d\n", node.id, node.view, node.faultNum)
	return node
}

func (n *Node) Run() {
	n.server.RegisterChan(n.requestRecv, n.prePrepareRecv, n.prepareRecv, n.commit)
	go n.server.Run()
	go n.requestRecvThread()
	go n.prePrepareSendThread()
	go n.prePrepareRecvAndPrepareSendThread()
	go n.prepareRecvAndCommitSendThread()
	go n.commitRecvThread()
	go n.executeAndReplyThread()
}
