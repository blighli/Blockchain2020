package message

import (
	"encoding/json"
	"strconv"
)

type TimeStamp uint64 // 时间戳格式
type Identify uint64  // 客户端标识格式
type View Identify    // 视图
type Sequence int64   // 序号

type Operation string
type Result struct {
}
type Request struct {
	Op        Operation `json:"operation"`
	TimeStamp TimeStamp `json:"timestamp"`
	ID        Identify  `json:"clientID"`
}
type Message struct {
	Requests []*Request `json:"requests"`
}
type PrePrepare struct {
	View     View     `json:"view"`
	Sequence Sequence `json:"sequence"`
	Digest   string   `json:"digest"`
	Message  Message  `json:"message"`
}
type Prepare struct {
	View     View     `json:"view"`
	Sequence Sequence `json:"sequence"`
	Digest   string   `json:"digest"`
	Identify Identify `json:"id"`
}
type Commit struct {
	View     View     `json:"view"`
	Sequence Sequence `json:"sequence"`
	Digest   string   `json:"digest"`
	Identify Identify `json:"id"`
}
type Reply struct {
	View      View      `json:"view"`
	TimeStamp TimeStamp `json:"timestamp"`
	Id        Identify  `json:"nodeID"`
	Result    Result    `json:"result"`
}

func NewPreprepareMsg(view View, seq Sequence, batch []*Request) ([]byte, *PrePrepare, string, error) {
	message := Message{Requests: batch}
	d, err := Digest(message)
	if err != nil {
		return []byte{}, nil, "", nil
	}
	prePrepare := &PrePrepare{
		View:     view,
		Sequence: seq,
		Digest:   d,
		Message:  message,
	}
	ret, err := json.Marshal(prePrepare)
	if err != nil {
		return []byte{}, nil, "", nil
	}
	return ret, prePrepare, d, nil
}

func NewPrepareMsg(id Identify, msg *PrePrepare) ([]byte, *Prepare, error) {
	prepare := &Prepare{
		View:     msg.View,
		Sequence: msg.Sequence,
		Digest:   msg.Digest,
		Identify: id,
	}
	content, err := json.Marshal(prepare)
	if err != nil {
		return []byte{}, nil, err
	}
	return content, prepare, nil
}

func NewCommitMsg(id Identify, msg *Prepare) ([]byte, *Commit, error) {
	commit := &Commit{
		View:     msg.View,
		Sequence: msg.Sequence,
		Digest:   msg.Digest,
		Identify: id,
	}
	content, err := json.Marshal(commit)
	if err != nil {
		return []byte{}, nil, err
	}
	return content, commit, nil
}

func ViewSequenceString(view View, seq Sequence) string {
	seqStr := strconv.Itoa(int(seq))
	viewStr := strconv.Itoa(int(view))
	seqLen := 4 - len(seqStr)
	viewLen := 28 - len(viewStr)
	for i := 0; i < seqLen; i++ {
		viewStr = "0" + viewStr
	}
	for i := 0; i < viewLen; i++ {
		seqStr = "0" + seqStr
	}
	return viewStr + seqStr
}
