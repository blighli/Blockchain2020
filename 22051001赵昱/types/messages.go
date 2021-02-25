package types

import (
	"crypto/md5"
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"
	"github.com/tendermint/go-wire"
)

// Digest摘要

func EQ(d1 []byte, d2 []byte) bool { //判断两个byte摘要是否相等
	if len(d1) != len(d2) {
		return false
	}
	for idx, b := range d1 {
		if b != d2[idx] {
			return false
		}
	}
	return true
}

// Checkpoint

func ToCheckpoint(sequence uint64, digest []byte) *Checkpoint {
	return &Checkpoint{sequence, digest}
}

// Entry

func ToEntry(sequence uint64, digest []byte, view uint64) *Entry {
	return &Entry{sequence, digest, view}
}

// ViewChange

func ToViewChange(viewchanger uint64, digest []byte) *ViewChange {
	return &ViewChange{viewchanger, digest}
}

// Summary

func ToSummary(sequence uint64, digest []byte) *Summary {
	return &Summary{sequence, digest}
}

// Request

func ToRequestClient(op *Operation, timestamp, client string) *Request { // client 是客户端ip地址
	return &Request{
		Value: &Request_Client{
			&RequestClient{op, timestamp, client}},
	}
}

func ToRequestPreprepare(view, sequence uint64, digest []byte, replica uint64) *Request {
	return &Request{
		Value: &Request_Preprepare{
			&RequestPrePrepare{view, sequence, digest, replica}},
	}
}

func ToRequestPrepare(view, sequence uint64, digest []byte, replica uint64) *Request {
	return &Request{
		Value: &Request_Prepare{
			&RequestPrepare{view, sequence, digest, replica}},
	}
}

func ToRequestCommit(view, sequence, replica uint64) *Request {
	return &Request{
		Value: &Request_Commit{
			&RequestCommit{view, sequence, replica}},
	}
}

func ToRequestCheckpoint(sequence uint64, digest []byte, replica uint64) *Request {
	return &Request{
		Value: &Request_Checkpoint{
			&RequestCheckpoint{sequence, digest, replica}},
	}
}

func ToRequestViewChange(view, sequence uint64, checkpoints []*Checkpoint, preps, prePreps []*Entry, replica uint64) *Request {
	return &Request{
		Value: &Request_Viewchange{
			&RequestViewChange{view, sequence, checkpoints, preps, prePreps, replica}},
	}
}

func ToRequestAck(view, replica, viewchanger uint64, digest []byte) *Request {
	return &Request{
		Value: &Request_Ack{
			&RequestAck{view, replica, viewchanger, digest}},
	}
}

func ToRequestNewView(view uint64, viewChanges []*ViewChange, summaries []*Summary, replica uint64) *Request {
	return &Request{
		Value: &Request_Newview{
			&RequestNewView{view, viewChanges, summaries, replica}},
	}
}

// Request Methods

func (req *Request) Digest() []byte {
	if req == nil {
		return nil
	}
	bytes := md5.Sum([]byte(req.String()))
	return bytes[:]
}

func (req *Request) LowWaterMark() uint64 {
	// only for requestViewChange
	reqViewChange := req.GetViewchange()
	checkpoints := reqViewChange.GetCheckpoints()
	lastStable := checkpoints[len(checkpoints)-1]
	lwm := lastStable.Sequence
	return lwm
}

// ToReply 生成reply
func ToReply(view uint64, timestamp, client string, replica uint64, result *Result) *Reply {
	return &Reply{view, timestamp, client, replica, result}
	// 注意timestamp是来自相应请求的时间戳，而非返回reply的时间戳
}

// Reply Methods
func (reply *Reply) Digest() []byte {
	if reply == nil {
		return nil
	}
	bytes := md5.Sum([]byte(reply.String()))
	return bytes[:]
}

// Write proto message
/* Origin
func WriteMessage(addr string, msg proto.Message) error { // addr是replica
	conn, err := net.Dial("tcp", addr)
	defer conn.Close()
	if err != nil {
		return err
	}
	bz, err := proto.Marshal(msg) // bz 返回的是data
	if err != nil {
		return err
	}
	var n int
	wire.WriteBinary(bz, conn, &n, &err)
	return err
}*/

// WriteMessage对每个端口addr发送消息，这是我修改tcpconn之后的代码
func WriteMessage(addr string, msg proto.Message) error { // addr是replica
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	//fmt.Println("connected!")

	bz, err := proto.Marshal(msg)
	if err != nil {
		fmt.Println("proto.Marshal error: ", err)
		return err
	}
	var n int
	wire.WriteBinary(bz, conn, &n, &err)
	return err
}

// Read proto message

func ReadMessage(conn net.Conn, msg proto.Message) error {
	n, err := int(0), error(nil)
	buf := wire.ReadByteSlice(conn, 0, &n, &err)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(buf, msg)
	return err
}