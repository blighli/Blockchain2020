package goPBFT

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	pb "github.com/zballs/goPBFT/types"
)

const (
	CHECKPOINT_PERIOD uint64 = 8
	CONSTANT_FACTOR   uint64 = 2
)

type Replica struct {
	//net.Listener 接口Accept()，Close()，Addr()，现用tcplistener代替
	tcpListener *net.TCPListener
	ID          uint64            //ID号
	replicas    map[uint64]string //map[KeyType]ValueType
	activeView  bool
	view        uint64                   //视图，为连续编号的整数
	sequence    uint64                   //序列号，从1开始，128轮回
	requestChan chan *pb.Request         //channel
	replyChan   chan *pb.Reply           //channel
	errChan     chan error               //channel
	requests    map[string][]*pb.Request //string为Request类型
	replies     map[string][]*pb.Reply   //string为Client地址
	lastReply   *pb.Reply
	pendingVC   []*pb.Request //viewchange
	executed    []uint64      //处理到的
	checkpoints []*pb.Checkpoint
}

type Client struct {
	tcpListener *net.TCPListener
	ID          uint64
	Addr        string
	replicas    map[uint64]string //map[KeyType]ValueType
	requestChan chan *pb.Request  //channel
	replyChan   chan *pb.Reply
	errChan     chan error
	requests    []*pb.Request
	replies     []*pb.Reply
	lastrequest *pb.Request
	execute     uint64
}

type Server struct {
	Port string
	Node *Replica
}

// Editing start
func newReplica(id uint64) *Replica {

	replica := &Replica{
		//net.Listener: net.Listener,
		ID: id,
		replicas: map[uint64]string{
			0: "127.0.0.1:1111",
			1: "127.0.0.1:1112",
			2: "127.0.0.1:1113",
			3: "127.0.0.1:1114",
		},
		activeView:  true,
		view:        0,
		sequence:    0,
		requestChan: make(chan *pb.Request, 10),
		replyChan:   make(chan *pb.Reply, 10),
		errChan:     make(chan error, 10),
		requests:    make(map[string][]*pb.Request),
		replies:     make(map[string][]*pb.Reply), //string 为client's IP:PORT
		lastReply:   new(pb.Reply),
		pendingVC:   make([]*pb.Request, 0),
		executed:    make([]uint64, 0),
	}
	replica.checkpoints = []*pb.Checkpoint{pb.ToCheckpoint(0, []byte(""))}
	return replica
}

// NewServer 初始化server
func NewServer(ID string) *Server {
	id, _ := strconv.ParseUint(ID, 10, 64) // 强制类型转换string->uint64得到id
	node := newReplica(id)
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1000000, 10) // 得到ms级时间戳
	Client := "127.0.0.1:9999"

	reply := pb.ToReply(
		0,
		timestamp, // timestamp
		Client,
		id,             //ID
		new(pb.Result)) // Result 是一个value string，如果成功，string值为"OK"

	//初始化一个checkpoint
	stateDigest := reply.Digest()
	checkpoint := pb.ToCheckpoint(0, stateDigest)
	node.addCheckpoint(checkpoint)

	//调用handlerequest时需要lastreply中有一个reply存在
	node.replies[Client] = append(node.replies[Client], reply)

	//node.replyChan <- reply node.requestChan <- req
	server := &Server{ID, node}
	fmt.Println("NewServer established, ID:", id)
	return server
}

// NewClient 初始化client
func NewClient() *Client {
	client := &Client{
		ID:   0,
		Addr: "127.0.0.1:9999",
		replicas: map[uint64]string{
			0: "127.0.0.1:1111",
			1: "127.0.0.1:1112",
			2: "127.0.0.1:1113",
			3: "127.0.0.1:1114",
		},
		requestChan: make(chan *pb.Request, 10),
		replyChan:   make(chan *pb.Reply, 10),
		errChan:     make(chan error, 10),
		requests:    make([]*pb.Request, CHECKPOINT_PERIOD), //128个
		replies:     make([]*pb.Reply, CHECKPOINT_PERIOD),
		lastrequest: new(pb.Request),
		execute:     0, //用于记录已经处理到的requests列表中的idx
	}
	return client
}

func (client *Client) Startlisten() {
	var tcpAddr *net.TCPAddr
	var err error
	if tcpAddr, err = net.ResolveTCPAddr("tcp", client.Addr); err != nil {
		fmt.Println("Resolve error\n", err)
	}
	if client.tcpListener, err = net.ListenTCP("tcp", tcpAddr); err != nil {
		fmt.Println("LISTEN_ERROR\n", err)
	}
	defer client.tcpListener.Close() //确保有错误时关闭监听

	for { // 无限循环监听reply
		conn, err := client.tcpListener.AcceptTCP()
		if err != nil {
			log.Panic(err)
			continue
		}
		rep := &pb.Reply{}
		err = pb.ReadMessage(conn, rep) // 读取reply
		if err != nil {
			log.Panic(err)
		}

		//我这里想做client端对reply的验证，但是会报错panic: runtime error: invalid memory address or nil pointer dereference

		/*client.replies = append(client.replies, rep) // log，之后可以改为logreply函数，不过现在已经被replica的方法占用
		count := 0                                   //记录成功的reply个数
		req := client.requests[client.execute]

		for _, reply := range client.replies {
			t := reply.Timestamp
			c := reply.Client
			id := reply.Replica
			result := reply.Result.Value // string type
			if t != req.GetClient().Timestamp || c != client.Addr || result != "OK" {
				continue
			}
			//走到这一步，说明这个reply是当前处理的req的reply，且为本client发送，且成功

			if id == rep.Replica { //如果这个reply是之前某节点发送过的，多发不算
				fmt.Printf("Replica %d sent multiple replies\n", id)
				continue
			}
			count++
			if count >= 2 { //4节点情况
				fmt.Println("Client " + client.Addr + " has accepted the reply.")
			}

		}*/
		fmt.Println("Client " + client.Addr + " has accepted the reply.")

	}
}

func (client *Client) SendRequest() {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1000000, 10) //得到ms时间戳
	//生成一个client request
	req := pb.ToRequestClient(
		&pb.Operation{1234567}, // operation struct uint64，以后需要修改为区块，先暂定为1234567
		timestamp,              // timestamp
		client.Addr)

	for _, replica := range client.replicas { //发送给所有节点
		err := pb.WriteMessage(replica, req)
		if err != nil {
			fmt.Println(err)
		}
	}
	client.requests = append(client.requests, req)
	client.lastrequest = req

	fmt.Println("Client " + client.Addr + " has sent a request at timestamp = " + timestamp)
}

// Start the server
func (server *Server) Start() {
	//runtime.GOMAXPROCS(2)
	var tcpAddr *net.TCPAddr
	var err error
	if tcpAddr, err = net.ResolveTCPAddr("tcp", server.Node.replicas[server.Node.ID]); err != nil {
		fmt.Println("Resolve Error\n", err)
	}
	if server.Node.tcpListener, err = net.ListenTCP("tcp", tcpAddr); err != nil {
		fmt.Println("LISTEN_ERROR\n", err)
	}
	defer server.Node.tcpListener.Close()
	go server.Node.sendRoutine()    //本地生成的request、reply、error，都放在chan里发送
	server.Node.acceptConnections() //由ReadMessage得到req，然后直接handle
}

func (rep *Replica) acceptConnections() {
	fmt.Println("Accepting Listening...")
	for {
		//conn, err := rep.Accept() //net.Listener的Accept() (Conn, error)    origin
		conn, err := rep.tcpListener.AcceptTCP()
		if err != nil {
			log.Panic(err) //  Panic is equivalent to Print() followed by a call to panic(). log里的函数，打印错误
			continue
		}

		req := &pb.Request{}
		err = pb.ReadMessage(conn, req)
		if err != nil {
			log.Panic(err)
		}
		rep.handleRequest(req) //处理request
	}
}

// Sends
func (rep *Replica) multicast(REQ *pb.Request) error { //广播
	fmt.Println("Multicasting to the other")
	for _, replica := range rep.replicas { //replica 是string地址 是每个节点监听的端口
		err := pb.WriteMessage(replica, REQ)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rep *Replica) sendRoutine() { //用于接收各种chan中的消息，并发给所有节点
	fmt.Println("SendRoutine Listening...")
	for {
		select { // 选择接收到是request 还是 reply
		case REQ := <-rep.requestChan: //rep.requestChan是chan *pb.Request  这里REQ为发出信道
			switch REQ.Value.(type) { //type-switch,
			case *pb.Request_Ack: //如果是ack
				//fmt.Printf("REQ is Ack")
				view := REQ.GetAck().View
				primaryID := rep.newPrimary(view)    // 得到新的主节点id, view是0，ID就是map中的0项，ID需要从0开始编号，为0
				primary := rep.replicas[primaryID]   // 得到0对应的“127.0.0.1:1111”
				err := pb.WriteMessage(primary, REQ) // 调用writemessage，返回错误，并引用调用REQ
				if err != nil {
					go func() { //有错误，则传入error channel
						rep.errChan <- err
					}()
				}
			default: //其他
				//fmt.Println("REQ is Default Case") // 除非是ask请求，其余请求均需要广播，multicast()
				err := rep.multicast(REQ)
				if err != nil {
					go func() {
						fmt.Println(err)
						rep.errChan <- err
					}()
				}
			}
		case reply := <-rep.replyChan:
			fmt.Println("reply")
			client := reply.Client
			err := pb.WriteMessage(client, reply)
			if err != nil {
				fmt.Println(err)
				go func() {
					rep.errChan <- err
				}()
			}
		}
	}
}

// Basic operations 结构体Replica的方法：
func (rep *Replica) primary() uint64 { //view取模副本数+1的值为primary（主节点）
	return rep.view % uint64(len(rep.replicas)+1) //p = v mod |R|，  **为什么+1？
}

func (rep *Replica) newPrimary(view uint64) uint64 { //从参数输入新的view，得到newprimary
	return view % uint64(len(rep.replicas)+1)
}

// 以上函数的rep.replicas+1 把+1都去掉了

func (rep *Replica) isPrimary(ID uint64) bool { //判断副本ID是否为主节点
	return ID == rep.primary()
}

func (rep *Replica) oneThird(count int) bool { //判断参数count是否>=副本数+1再/3
	return count >= (len(rep.replicas)+1)/3
}

func (rep *Replica) overOneThird(count int) bool { //判断参数count是否>副本数+1再/3
	return count > (len(rep.replicas)+1)/3
}

func (rep *Replica) twoThirds(count int) bool { //判断参数count是否>=副本数+1再*2/3
	return count >= 2*(len(rep.replicas))/3
}

func (rep *Replica) overTwoThirds(count int) bool { //判断参数count是否>副本数+1再*2/3
	return count > 2*(len(rep.replicas))/3
}

func (rep *Replica) lowWaterMark() uint64 { //watermark下限
	return rep.lastStable().Sequence //预准备消息的序号n必须在watermark上下限之间，防止一个失效节点使用一个很大的序号消耗序号空间。
}

func (rep *Replica) highWaterMark() uint64 { //watermark上限
	return rep.lowWaterMark() + CHECKPOINT_PERIOD*CONSTANT_FACTOR //下限+128*2
}

func (rep *Replica) sequenceInRange(sequence uint64) bool { //判断序号是否在watermark中
	return sequence > rep.lowWaterMark() && sequence <= rep.highWaterMark()
}

func (rep *Replica) lastExecuted() uint64 {
	return rep.executed[len(rep.executed)-1]
}

func (rep *Replica) lastStable() *pb.Checkpoint { //返回最后一个checkpoint
	return rep.checkpoints[len(rep.checkpoints)-1]
}

func (rep *Replica) theLastReply() *pb.Reply {
	lastReply := rep.lastReply
	for _, replies := range rep.replies { // 遍历replies map[string][]*pb.Reply  ，key应该是client
		reply := replies[len(replies)-1]           // 每个client对于map中的[]*pb.Reply中最后一项
		if reply.Timestamp > lastReply.Timestamp { //如果时间戳延后，则更新lastreply
			lastReply = reply
		}
	}
	return lastReply
}

func (rep *Replica) lastReplyToClient(client string) *pb.Reply { //返回给某个client的最后一个reply
	return rep.replies[client][len(rep.replies[client])-1]
}

func (rep *Replica) stateDigest() []byte { //返回最后reply的摘要，不过Digest()方法没有找到？应该是对应到Reply struct的方法吧？
	return rep.theLastReply().Digest()
}

func (rep *Replica) isCheckpoint(sequence uint64) bool { //每CHECKPOINT_PERIOD=128设置一个检查点
	return sequence%CHECKPOINT_PERIOD == 0
}

func (rep *Replica) addCheckpoint(checkpoint *pb.Checkpoint) {
	rep.checkpoints = append(rep.checkpoints, checkpoint) //checkpoints []*pb.Checkpoint
}

func (rep *Replica) allRequests() []*pb.Request {
	var requests []*pb.Request
	for _, reqs := range rep.requests { //把map[string][]*pb.Request 所有request列进一个list
		requests = append(requests, reqs...)
	}
	return requests
}

// Log
// !hasRequest before append?

func (rep *Replica) logRequest(REQ *pb.Request) { //分类
	switch REQ.Value.(type) {
	case *pb.Request_Client:
		rep.requests["client"] = append(rep.requests["client"], REQ)
	case *pb.Request_Preprepare:
		rep.requests["pre-prepare"] = append(rep.requests["pre-prepare"], REQ)
	case *pb.Request_Prepare:
		rep.requests["prepare"] = append(rep.requests["prepare"], REQ)
	case *pb.Request_Commit:
		rep.requests["commit"] = append(rep.requests["commit"], REQ)
	case *pb.Request_Checkpoint:
		rep.requests["checkpoint"] = append(rep.requests["checkpoint"], REQ)
	case *pb.Request_Viewchange:
		rep.requests["view-change"] = append(rep.requests["view-change"], REQ)
	case *pb.Request_Ack:
		rep.requests["ack"] = append(rep.requests["ack"], REQ)
	case *pb.Request_Newview:
		rep.requests["new-view"] = append(rep.requests["new-view"], REQ)
	default:
		log.Printf("Replica %d tried logging unrecognized request type\n", rep.ID)
	}
	//EDITING
	//fmt.Println("Logged")
}

func (rep *Replica) logPendingVC(REQ *pb.Request) error {
	switch REQ.Value.(type) {
	case *pb.Request_Viewchange:
		rep.pendingVC = append(rep.pendingVC, REQ) //pendingVC []*pb.Request
		return nil
	default:
		return errors.New("Request is wrong type") // 见error package的定义，返回结构体地址
	}
}

func (rep *Replica) logReply(client string, reply *pb.Reply) bool { // 记录最新的reply到 rep的 replies二维数组
	lastReplyToClient := rep.lastReplyToClient(client)
	//fmt.Println(lastReplyToClient.Timestamp, reply.Timestamp)
	if lastReplyToClient == nil {
		fmt.Println("none lastreply")
		return false
	}
	if lastReplyToClient.Timestamp < reply.Timestamp {
		rep.replies[client] = append(rep.replies[client], reply)
		return false
	}
	return true
}

// Has requests or reply

/*func (rep *Replica) hasReply(REP *pb.Reply, client string) bool {
	if rep.lastReplyToClient(client).Timestamp >= REP.Timestamp {
		return true
	}
	return false
}*/

func (rep *Replica) hasRequest(REQ *pb.Request) bool { //各个类型的has函数见以下定义

	switch REQ.Value.(type) {
	case *pb.Request_Preprepare:
		return rep.hasRequestPreprepare(REQ)
	case *pb.Request_Prepare:
		return rep.hasRequestPrepare(REQ)
	case *pb.Request_Commit:
		return rep.hasRequestCommit(REQ)
	case *pb.Request_Viewchange:
		return rep.hasRequestViewChange(REQ)
	case *pb.Request_Ack:
		return rep.hasRequestAck(REQ)
	case *pb.Request_Newview:
		return rep.hasRequestNewView(REQ)
	default:
		return false
	}
}

func (rep *Replica) hasRequestPreprepare(REQ *pb.Request) bool {
	view := REQ.GetPreprepare().View
	sequence := REQ.GetPreprepare().Sequence
	digest := REQ.GetPreprepare().Digest
	for _, req := range rep.requests["pre-prepare"] {
		v := req.GetPreprepare().View
		s := req.GetPreprepare().Sequence
		d := req.GetPreprepare().Digest
		if v == view && s == sequence && pb.EQ(d, digest) {
			return true
		}
	}
	return false
}

func (rep *Replica) hasRequestPrepare(REQ *pb.Request) bool {
	view := REQ.GetPrepare().View
	sequence := REQ.GetPrepare().Sequence
	digest := REQ.GetPrepare().Digest
	replica := REQ.GetPrepare().Replica
	for _, req := range rep.requests["prepare"] {
		v := req.GetPrepare().View
		s := req.GetPrepare().Sequence
		d := req.GetPrepare().Digest
		r := req.GetPrepare().Replica
		if v == view && s == sequence && pb.EQ(d, digest) && r == replica {
			return true
		}
	}
	return false
}

func (rep *Replica) hasRequestCommit(REQ *pb.Request) bool {
	view := REQ.GetCommit().View
	sequence := REQ.GetCommit().Sequence
	replica := REQ.GetCommit().Replica
	for _, req := range rep.requests["commit"] {
		v := req.GetCommit().View
		s := req.GetCommit().Sequence
		r := req.GetCommit().Replica
		if v == view && s == sequence && r == replica {
			return true
		}
	}
	return false
}

func (rep *Replica) hasRequestViewChange(REQ *pb.Request) bool {
	view := REQ.GetViewchange().View
	replica := REQ.GetViewchange().Replica
	for _, req := range rep.requests["view-change"] {
		v := req.GetViewchange().View
		r := req.GetViewchange().Replica
		if v == view && r == replica {
			return true
		}
	}
	return false
}

func (rep *Replica) hasRequestAck(REQ *pb.Request) bool {
	view := REQ.GetAck().View
	replica := REQ.GetAck().Replica
	viewchanger := REQ.GetAck().Viewchanger
	for _, req := range rep.requests["ack"] {
		v := req.GetAck().View
		r := req.GetAck().Replica
		vc := req.GetAck().Viewchanger
		if v == view && r == replica && vc == viewchanger {
			return true
		}
	}
	return false
}

func (rep *Replica) hasRequestNewView(REQ *pb.Request) bool {
	view := REQ.GetNewview().View
	for _, req := range rep.requests["new-view"] {
		v := req.GetNewview().View
		if v == view {
			return true
		}
	}
	return false
}

// Clear requests

func (rep *Replica) clearRequestsBySeq(sequence uint64) { //调用所有clear函数
	rep.clearRequestClients()
	rep.clearRequestPrepreparesBySeq(sequence)
	rep.clearRequestPreparesBySeq(sequence)
	rep.clearRequestCommitsBySeq(sequence)
	//rep.clearRequestCheckpointsBySeq(sequence)
}

func (rep *Replica) clearRequestClients() {
	clientReqs := rep.requests["client"]
	lastTimestamp := rep.theLastReply().Timestamp
	for idx, req := range rep.requests["client"] {
		timestamp := req.GetClient().Timestamp // 表示这个client发送请求的时间戳
		if lastTimestamp >= timestamp && idx < (len(clientReqs)-1) {
			//如果 最后一个时间戳大于当前请求的时间戳 且 idx小于最后一个的idx，
			clientReqs = append(clientReqs[:idx], clientReqs[idx+1:]...)
		} else {
			clientReqs = clientReqs[:idx-1]
		}
	}
	rep.requests["client"] = clientReqs
}

func (rep *Replica) clearRequestPrepreparesBySeq(sequence uint64) {
	prePrepares := rep.requests["pre-prepare"]
	for idx, req := range rep.requests["pre-prepare"] {
		s := req.GetPreprepare().Sequence
		if s <= sequence {
			prePrepares = append(prePrepares[:idx], prePrepares[idx+1:]...)
		}
	}
	rep.requests["pre-prepare"] = prePrepares
}

func (rep *Replica) clearRequestPreparesBySeq(sequence uint64) {
	prepares := rep.requests["prepare"]
	for idx, req := range rep.requests["prepare"] {
		s := req.GetPrepare().Sequence
		if s <= sequence {
			prepares = append(prepares[:idx], prepares[idx+1:]...)
		}
	}
	rep.requests["prepare"] = prepares
}

func (rep *Replica) clearRequestCommitsBySeq(sequence uint64) {
	commits := rep.requests["commit"]
	for idx, req := range rep.requests["commit"] {
		s := req.GetCommit().Sequence
		if s <= sequence {
			commits = append(commits[:idx], commits[idx+1:]...)
		}
	}
	rep.requests["commit"] = commits
}

func (rep *Replica) clearRequestCheckpointsBySeq(sequence uint64) {
	checkpoints := rep.requests["checkpoint"]
	for idx, req := range rep.requests["checkpoint"] {
		s := req.GetCheckpoint().Sequence
		if s <= sequence {
			checkpoints = append(checkpoints[:idx], checkpoints[idx+1:]...)
		}
	}
	rep.requests["checkpoint"] = checkpoints
}

func (rep *Replica) clearRequestsByView(view uint64) {
	rep.clearRequestPrepreparesByView(view)
	rep.clearRequestPreparesByView(view)
	rep.clearRequestCommitsByView(view)
	//add others
}

func (rep *Replica) clearRequestPrepreparesByView(view uint64) {
	prePrepares := rep.requests["pre-prepare"]
	for idx, req := range rep.requests["pre-prepare"] {
		v := req.GetPreprepare().View
		if v < view {
			prePrepares = append(prePrepares[:idx], prePrepares[idx+1:]...)
		}
	}
	rep.requests["pre-prepare"] = prePrepares
}

func (rep *Replica) clearRequestPreparesByView(view uint64) {
	prepares := rep.requests["prepare"]
	for idx, req := range rep.requests["prepare"] {
		v := req.GetPrepare().View
		if v < view {
			prepares = append(prepares[:idx], prepares[idx+1:]...)
		}
	}
	rep.requests["prepare"] = prepares
}

func (rep *Replica) clearRequestCommitsByView(view uint64) {
	commits := rep.requests["commit"]
	for idx, req := range rep.requests["commit"] {
		v := req.GetCommit().View
		if v < view {
			commits = append(commits[:idx], commits[idx+1:]...)
		}
	}
	rep.requests["commit"] = commits
}

// Handle requests

func (rep *Replica) handleRequest(REQ *pb.Request) {
	//switch type，对于不同的request类型交付给不同的函数处理
	switch REQ.Value.(type) {
	case *pb.Request_Client:
		rep.handleRequestClient(REQ)

	case *pb.Request_Preprepare:
		rep.handleRequestPreprepare(REQ)

	case *pb.Request_Prepare:
		rep.handleRequestPrepare(REQ)

	case *pb.Request_Commit:
		rep.handleRequestCommit(REQ)

	case *pb.Request_Checkpoint:
		rep.handleRequestCheckpoint(REQ)

	case *pb.Request_Viewchange:
		rep.handleRequestViewChange(REQ)

	case *pb.Request_Ack:
		rep.handleRequestAck(REQ)

	default:
		fmt.Printf("Replica %d received unrecognized request type\n", rep.ID)
	}
}

func (rep *Replica) handleRequestClient(REQ *pb.Request) {
	fmt.Printf("Replica %d received request client \n", rep.ID)
	rep.sequence++                   // 序列号+1，初始为0，第一个req进来时为1
	client := REQ.GetClient().Client // REQ的client和timestamp
	timestamp := REQ.GetClient().Timestamp
	lastReplyToClient := rep.lastReplyToClient(client) //返回一个reply结构，是当前client最后的reply
	if lastReplyToClient.Timestamp == timestamp {      //如果最后一个reply的ts和当前请求的ts一样，说明已经处理过了
		fmt.Println("Has handled the RequestClient")
		reply := pb.ToReply(rep.view, timestamp, client, rep.ID, lastReplyToClient.Result)
		rep.logReply(client, reply)
		go func() {
			rep.replyChan <- reply // 重新发送reply给client
		}()
		return
	}
	rep.logRequest(REQ)
	if !rep.isPrimary(rep.ID) {
		fmt.Printf("%d is not the primary \n", rep.ID)
		return
	} // 如果不是主节点，则返回，由主节点来发送preprepare请求

	req := pb.ToRequestPreprepare(rep.view, rep.sequence, REQ.Digest(), rep.ID)
	// pre-prepare中的摘要是client req计算的摘要
	//把replica的sequence n分配给m，m是指request消息，然后求这个m的digest
	rep.logRequest(req)
	log.Println("=====[Cliented]=====")
	go func() {
		rep.requestChan <- req
	}()
}

func (rep *Replica) handleRequestPreprepare(REQ *pb.Request) {
	replica := REQ.GetPreprepare().Replica //得到id
	fmt.Printf("Replica %d received request preprepare from Replica %d\n", rep.ID, replica)
	//判断request是否为主节点，preprepare是primary给其他备份节点发的，所以REQ的发送方如不是primary，就return
	if !rep.isPrimary(replica) {
		return
	}
	view := REQ.GetPreprepare().View
	if rep.view != view {
		return
	}
	sequence := REQ.GetPreprepare().Sequence
	//fmt.Println(view, sequence)  测试过了，得到的是初始replica时的view和sequence
	if !rep.sequenceInRange(sequence) {
		return
	}
	digest := REQ.GetPreprepare().Digest
	accept := true
	for _, req := range rep.requests["pre-prepare"] {
		v := req.GetPreprepare().View
		s := req.GetPreprepare().Sequence
		d := req.GetPreprepare().Digest
		if v == view && s == sequence && !pb.EQ(d, digest) {
			// 当该备份节点之前已经接收到了一条在同一view下并且编号也是s，但是digest不同的PRE-PREPARE信息时，就会拒绝。
			accept = false
			break
		}
	}
	if !accept {
		return
	}
	rep.logRequest(REQ) // 记录到request['preprepare'][]*pb.Request中
	log.Println("=====[PrePrepared]=====")
	req := pb.ToRequestPrepare(view, sequence, digest, rep.ID)
	if rep.hasRequest(req) {
		//如果log中已经有了，就不发，如果还没有就记录在log中（request['prepare'][]*pb.Request列表中）
		return
	}
	rep.logRequest(req) //发前已经log，所以在收到自己的req时会有提示multisent
	go func() {
		rep.requestChan <- req
	}()
}

func (rep *Replica) handleRequestPrepare(REQ *pb.Request) {
	replica := REQ.GetPrepare().Replica
	fmt.Printf("Replica %d received request prepare from Replica %d \n", rep.ID, replica)
	if rep.isPrimary(replica) { // prepare只需要副本发送，主节点不发送，所以如果REQ的发送方是primary，就return
		return
	}
	view := REQ.GetPrepare().View
	if rep.view != view {
		return
	}
	sequence := REQ.GetPrepare().Sequence
	if !rep.sequenceInRange(sequence) {
		return
	}
	digest := REQ.GetPrepare().Digest
	rep.logRequest(REQ) // 要先记录请求，因为要计数，其他线程需要查询replica.requests
	//prePrepared := make(chan bool, 1)
	twoThirds := make(chan bool, 1)

	go func() {
		count := 0
		for _, req := range rep.requests["prepare"] {
			v := req.GetPrepare().View
			s := req.GetPrepare().Sequence
			d := req.GetPrepare().Digest
			r := req.GetPrepare().Replica
			if v != view || s != sequence || !pb.EQ(d, digest) {
				//反复查找已收到的prepare请求，如果v和当前请求的view相同，且序列号，摘要都相同，则继续往下，否则重新查找
				continue
			}
			if r == replica { //如果当前的请求发送方已经发送过，输出这句话，且prePrepared置为true
				fmt.Printf("Replica %d sent multiple prepare requests\n", replica)
				continue
			}
			count++
			if rep.twoThirds(count) { //如果4节点，那至少要收到2条才行
				twoThirds <- true
				return
			}
		}
		twoThirds <- false
	}()

	//用两个bool变量记录两个判断条件的返回值
	var twothirds bool
	twothirds = <-twoThirds

	if !twothirds { //用channel等待线程
		log.Println("twothirds:", twothirds)
		return
	}
	req := pb.ToRequestCommit(view, sequence, rep.ID)
	if rep.hasRequest(req) { //这里如果发过了commit了，就return
		return
	}
	log.Printf("=====[Prepared]=====")
	rep.logRequest(req)
	go func() {
		rep.requestChan <- req
	}()
}

func (rep *Replica) handleRequestCommit(REQ *pb.Request) {
	//runtime.GOMAXPROCS(4)
	replica := REQ.GetCommit().Replica // commit请求的发送方
	fmt.Printf("Replica %d received request commit from %d \n", rep.ID, replica)
	view := REQ.GetCommit().View
	if rep.view != view {
		return
	}
	sequence := REQ.GetCommit().Sequence
	log.Println("Commit sequence:", sequence)
	if !rep.sequenceInRange(sequence) {
		return
	}
	rep.logRequest(REQ)
	twoThirds := make(chan bool, 1)
	go func() { // 这里仿照prepare阶段，另起线程来判断个数
		count := 0
		for _, req := range rep.requests["commit"] {
			v := req.GetCommit().View
			s := req.GetCommit().Sequence
			rr := req.GetCommit().Replica
			if v != view || s != sequence {
				continue
			}
			if rr == replica {
				fmt.Printf("Replica %d sent multiple commit requests\n", replica) //这句话貌似多余，之前都log了，所以这个请求肯定会有的，而不是多发了。
				continue
			}
			count++
			if rep.overTwoThirds(count) { // 4节点的话就必须要收到4个才行，包括本地
				twoThirds <- true
				return
			}
		}
		twoThirds <- false
	}()

	if !<-twoThirds {
		return
	}

	/*for _, req := range rep.requests["commit"] {
	v := req.GetCommit().View
	s := req.GetCommit().Sequence
	rr := req.GetCommit().Replica
	if v != view || s != sequence {

		continue
	}
	if rr == replica {
		//fmt.Printf("Replica %d sent multiple commit requests\n", replica) 这句话多余，之前都log了，所以这个请求肯定会有的，而不是多发了。
		continue
	}
	count++
	if !rep.overTwoThirds(count) { // 4节点的话就必须要收到4个才行，包括本地
		continue
	}*/

	var digest []byte
	digests := make(map[string]struct{})
	for _, req := range rep.requests["pre-prepare"] {
		v := req.GetPreprepare().View
		s := req.GetPreprepare().Sequence
		if v == view && s <= sequence { //相同view下，序列号小于当前请求序列号的所有preprepare请求
			//digests[string(req.Digest())] = struct{}{}
			digests[string(req.GetPreprepare().Digest)] = struct{}{}
			if s == sequence {
				//digest = req.Digest()
				digest = req.GetPreprepare().Digest
				break
			}
		}
	}
	for _, req := range rep.requests["client"] {
		d := req.Digest()
		if _, exists := digests[string(d)]; !exists { //如果不存在，继续搜rep.requests["client"]
			log.Printf("continue\n")
			continue
		}
		if !pb.EQ(d, digest) { //preprepare请求中v，s都与当前请求相同的消息的digest，如果和client中相同
			// Unexecuted message op with
			// lower sequence number...
			//log.Printf("digest not equal\n")
			//break
			continue
		}

		// op := req.GetClient().Op
		timestamp := req.GetClient().Timestamp
		client := req.GetClient().Client
		result := &pb.Result{"OK"} //result定义，如果确认为"OK"

		rep.executed = append(rep.executed, sequence) //当前序列号的请求被处理好了

		reply := pb.ToReply(view, timestamp, client, rep.ID, result)

		/*if rep.hasReply(reply, client) {
			return
		}*/

		hasReply := rep.logReply(client, reply)
		if hasReply {
			//log.Println("hasReply")
			return
		}

		log.Printf("=====[Commited]=====")
		rep.lastReply = reply

		go func() {
			rep.replyChan <- reply
		}()

		if !rep.isCheckpoint(sequence) {
			// 如果没到checkpoint，直接返回，否则需要发送请求，清除已有的消息log，执行下k个请求
			return
		}
		stateDigest := rep.stateDigest()                             // 返回最后reply的摘要
		req := pb.ToRequestCheckpoint(sequence, stateDigest, rep.ID) // 生成checkpoint的请求
		rep.logRequest(req)
		//checkpoint := pb.ToCheckpoint(sequence, stateDigest)
		//rep.addCheckpoint(checkpoint)
		go func() {
			rep.requestChan <- req
		}()
		return
	}
}

func (rep *Replica) handleRequestCheckpoint(REQ *pb.Request) {

	sequence := REQ.GetCheckpoint().Sequence
	if !rep.sequenceInRange(sequence) {
		return
	}
	digest := REQ.GetCheckpoint().Digest
	replica := REQ.GetCheckpoint().Replica
	rep.logRequest(REQ)
	count := 0
	for _, req := range rep.requests["checkpoint"] {
		s := req.GetCheckpoint().Sequence
		d := req.GetCheckpoint().Digest
		r := req.GetCheckpoint().Replica
		if s != sequence || !pb.EQ(d, digest) {
			continue
		}
		if r == replica && len(rep.replicas) != 1 {
			fmt.Printf("Replica %d sent multiple checkpoint requests\n", replica)
			continue
		}
		count++
		if !rep.overTwoThirds(count) { //超过2/3的，4个里面需要4个
			continue
		}
		for _, checkpoint := range rep.checkpoints {
			if sequence == checkpoint.Sequence && pb.EQ(digest, checkpoint.Digest) {
				return
			}
		}
		// rep.clearEntries(sequence)
		rep.clearRequestsBySeq(sequence)
		checkpoint := pb.ToCheckpoint(sequence, digest)
		rep.addCheckpoint(checkpoint)
		log.Printf("=====[Checkpoint and clear request done]=====\n")
		return
	}
	return
}

/* type中RequestViewChange的定义:
type RequestViewChange struct {
	View        uint64        `protobuf:"varint,1,opt,name=view" json:"view,omitempty"`
	Sequence    uint64        `protobuf:"varint,2,opt,name=sequence" json:"sequence,omitempty"`
	Checkpoints []*Checkpoint `protobuf:"bytes,3,rep,name=checkpoints" json:"checkpoints,omitempty"`
	Preps       []*Entry      `protobuf:"bytes,4,rep,name=preps" json:"preps,omitempty"`
	Prepreps    []*Entry      `protobuf:"bytes,5,rep,name=prepreps" json:"prepreps,omitempty"`
	Replica     uint64        `protobuf:"varint,6,opt,name=replica" json:"replica,omitempty"`
}
*/

func (rep *Replica) handleRequestViewChange(REQ *pb.Request) {
	view := REQ.GetViewchange().View
	if view < rep.view { // 如果当前view是大于请求的view的，return
		return
	}
	reqViewChange := REQ.GetViewchange()
	for _, prep := range reqViewChange.GetPreps() {
		v := prep.View
		s := prep.Sequence
		if v >= view || !rep.sequenceInRange(s) {
			return
		}
	}
	for _, prePrep := range reqViewChange.GetPrepreps() {
		v := prePrep.View
		s := prePrep.Sequence
		if v >= view || !rep.sequenceInRange(s) {
			return
		}
	}
	for _, checkpoint := range reqViewChange.GetCheckpoints() {
		s := checkpoint.Sequence
		if !rep.sequenceInRange(s) {
			return
		}
	}
	if rep.hasRequest(REQ) {
		return
	}
	rep.logRequest(REQ)
	viewchanger := reqViewChange.Replica

	req := pb.ToRequestAck(
		view,
		rep.ID,
		viewchanger,
		REQ.Digest())

	go func() {
		rep.requestChan <- req
	}()
}

func (rep *Replica) handleRequestAck(REQ *pb.Request) {

	view := REQ.GetAck().View
	primaryID := rep.newPrimary(view)

	if rep.ID != primaryID {
		return
	}

	if rep.hasRequest(REQ) {
		return
	}

	rep.logRequest(REQ)

	replica := REQ.GetAck().Replica
	viewchanger := REQ.GetAck().Viewchanger
	digest := REQ.GetAck().Digest

	reqViewChange := make(chan *pb.Request, 1)
	twoThirds := make(chan bool, 1)

	go func() {
		for _, req := range rep.requests["view-change"] {
			v := req.GetViewchange().View
			vc := req.GetViewchange().Replica
			if v == view && vc == viewchanger {
				reqViewChange <- req
			}
		}
		reqViewChange <- nil
	}()

	go func() {
		count := 0
		for _, req := range rep.requests["ack"] {
			v := req.GetAck().View
			r := req.GetAck().Replica
			vc := req.GetAck().Viewchanger
			d := req.GetAck().Digest
			if v != view || vc != viewchanger || !pb.EQ(d, digest) {
				continue
			}
			if r == replica {
				fmt.Printf("Replica %d sent multiple ack requests\n", replica)
				continue
			}
			count++
			if rep.twoThirds(count) {
				twoThirds <- true
				return
			}
		}
		twoThirds <- false
	}()

	req := <-reqViewChange

	if req == nil || !<-twoThirds {
		return
	}

	rep.logPendingVC(req)

	// When to send new view?
	rep.requestNewView(view)
}

func (rep *Replica) handleRequestNewView(REQ *pb.Request) {

	view := REQ.GetNewview().View

	if view == 0 || view < rep.view {
		return
	}

	replica := REQ.GetNewview().Replica
	primary := rep.newPrimary(view)

	if replica != primary {
		return
	}

	if rep.hasRequest(REQ) {
		return
	}

	rep.logRequest(REQ)

	rep.processNewView(REQ)
}

func (rep *Replica) correctViewChanges(viewChanges []*pb.ViewChange) (requests []*pb.Request) {

	// Returns requests if correct, else returns nil

	valid := false
	for _, vc := range viewChanges {
		for _, req := range rep.requests["view-change"] {
			d := req.Digest()
			if !pb.EQ(d, vc.Digest) {
				continue
			}
			requests = append(requests, req)
			v := req.GetViewchange().View
			// VIEW or rep.view??
			if v == rep.view {
				valid = true
				break
			}
		}
		if !valid {
			return nil
		}
	}

	if rep.isPrimary(rep.ID) {
		reps := make(map[uint64]int)
		valid = false
		for _, req := range rep.requests["ack"] {
			reqAck := req.GetAck()
			reps[reqAck.Replica]++
			if rep.twoThirds(reps[reqAck.Replica]) { //-2
				valid = true
				break
			}
		}
		if !valid {
			return nil
		}
	}

	return
}

func (rep *Replica) correctSummaries(requests []*pb.Request, summaries []*pb.Summary) (correct bool) {

	// Verify SUMMARIES

	var start uint64
	var digest []byte
	digests := make(map[uint64][]byte)

	for _, summary := range summaries {
		s := summary.Sequence
		d := summary.Digest
		if _d, ok := digests[s]; ok && !pb.EQ(_d, d) {
			return
		} else if !ok {
			digests[s] = d
		}
		if s < start || start == uint64(0) {
			start = s
			digest = d
		}
	}

	var A1 []*pb.Request
	var A2 []*pb.Request

	valid := false
	for _, req := range requests {
		reqViewChange := req.GetViewchange()
		s := reqViewChange.Sequence
		if s <= start {
			A1 = append(A1, req)
		}
		checkpoints := reqViewChange.GetCheckpoints()
		for _, checkpoint := range checkpoints {
			if checkpoint.Sequence == start && pb.EQ(checkpoint.Digest, digest) {
				A2 = append(A2, req)
				break
			}
		}
		if rep.twoThirds(len(A1)) && rep.oneThird(len(A2)) {
			valid = true
			break
		}
	}

	if !valid {
		return
	}

	end := start + CHECKPOINT_PERIOD*CONSTANT_FACTOR

	for seq := start; seq <= end; seq++ {

		valid = false

		for _, summary := range summaries {

			if summary.Sequence != seq {
				continue
			}

			if summary.Digest != nil {

				var view uint64

				for _, req := range requests {
					reqViewChange := req.GetViewchange()
					preps := reqViewChange.GetPreps()
					for _, prep := range preps {
						s := prep.Sequence
						d := prep.Digest
						if s != summary.Sequence || !pb.EQ(d, summary.Digest) {
							continue
						}
						v := prep.View
						if v > view {
							view = v
						}
					}
				}

				verifiedA1 := make(chan bool, 1)

				// Verify A1
				go func() {

					var A1 []*pb.Request

				FOR_LOOP:
					for _, req := range requests {
						reqViewChange := req.GetViewchange()
						s := reqViewChange.Sequence
						if s >= summary.Sequence {
							continue
						}
						preps := reqViewChange.GetPreps()
						for _, prep := range preps {
							s = prep.Sequence
							if s != summary.Sequence {
								continue
							}
							d := prep.Digest
							v := prep.View
							if v > view || (v == view && !pb.EQ(d, summary.Digest)) {
								continue FOR_LOOP
							}
						}
						A1 = append(A1, req)
						if rep.twoThirds(len(A1)) {
							verifiedA1 <- true
							return
						}
					}
					verifiedA1 <- false
				}()

				verifiedA2 := make(chan bool, 1)

				// Verify A2
				go func() {

					var A2 []*pb.Request

					for _, req := range requests {
						reqViewChange := req.GetViewchange()
						prePreps := reqViewChange.GetPrepreps()
						for _, prePrep := range prePreps {
							s := prePrep.Sequence
							d := prePrep.Digest
							v := prePrep.View
							if s == summary.Sequence && pb.EQ(d, summary.Digest) && v >= view {
								A2 = append(A2, req)
								break
							}
						}
						if rep.oneThird(len(A2)) {
							verifiedA2 <- true
							return
						}
					}
					verifiedA2 <- false
				}()

				if !<-verifiedA1 || !<-verifiedA2 {
					continue
				}

				valid = true
				break

			} else {

				var A1 []*pb.Request

			FOR_LOOP:

				for _, req := range requests {

					reqViewChange := req.GetViewchange()

					s := reqViewChange.Sequence

					if s >= summary.Sequence {
						continue
					}

					preps := reqViewChange.GetPreps()
					for _, prep := range preps {
						if prep.Sequence == summary.Sequence {
							continue FOR_LOOP
						}
					}

					A1 = append(A1, req)
					if rep.twoThirds(len(A1)) {
						valid = true
						break
					}
				}
				if valid {
					break
				}
			}
		}
		if !valid {
			return
		}
	}

	return true
}

func (rep *Replica) processNewView(REQ *pb.Request) (success bool) {
	if rep.activeView {
		return
	}
	reqNewView := REQ.GetNewview()
	viewChanges := reqNewView.GetViewchanges()
	requests := rep.correctViewChanges(viewChanges)
	if requests == nil {
		return
	}
	summaries := reqNewView.GetSummaries()
	correct := rep.correctSummaries(requests, summaries)
	if !correct {
		return
	}
	var h uint64
	for _, checkpoint := range rep.checkpoints {
		if checkpoint.Sequence < h || h == uint64(0) {
			h = checkpoint.Sequence
		}
	}
	var s uint64
	for _, summary := range summaries {
		if summary.Sequence < s || s == uint64(0) {
			s = summary.Sequence
		}
		if summary.Sequence > h {
			valid := false
			for _, req := range rep.requests["view-change"] { //in
				if pb.EQ(req.Digest(), summary.Digest) {
					valid = true
					break
				}
			}
			if !valid {
				return
			}
		}
	}

	if h < s {
		return
	}

	// Process new view
	rep.activeView = true

	for _, summary := range summaries {

		if rep.ID != reqNewView.Replica {
			req := pb.ToRequestPrepare(
				reqNewView.View,
				summary.Sequence,
				summary.Digest,
				rep.ID) // the backup sends/logs prepare

			go func() {
				if !rep.hasRequest(req) {
					rep.requestChan <- req
				}
			}()
			if summary.Sequence <= h {
				continue
			}

			if !rep.hasRequest(req) {
				rep.logRequest(req)
			}
		} else {
			if summary.Sequence <= h {
				break
			}
		}

		req := pb.ToRequestPreprepare(
			reqNewView.View,
			summary.Sequence,
			summary.Digest,
			reqNewView.Replica) // new primary pre-prepares

		if !rep.hasRequest(req) {
			rep.logRequest(req)
		}
	}

	var maxSequence uint64
	for _, req := range rep.requests["pre-prepare"] {
		reqPrePrepare := req.GetPreprepare()
		if reqPrePrepare.Sequence > maxSequence {
			maxSequence = reqPrePrepare.Sequence
		}
	}
	rep.sequence = maxSequence
	return true
}

func (rep *Replica) prePrepBySequence(sequence uint64) []*pb.Entry {
	var view uint64
	var requests []*pb.Request
	for _, req := range rep.requests["pre-prepare"] {
		v := req.GetPreprepare().View
		s := req.GetPreprepare().Sequence
		if v >= view && s == sequence {
			view = v
			requests = append(requests, req)
		}
	}
	if requests == nil {
		return nil
	}
	var prePreps []*pb.Entry
	for _, req := range requests {
		v := req.GetPreprepare().View
		if v == view {
			s := req.GetPreprepare().Sequence
			d := req.GetPreprepare().Digest
			prePrep := pb.ToEntry(s, d, v)
			prePreps = append(prePreps, prePrep)
		}
	}
FOR_LOOP:
	for _, prePrep := range prePreps {
		for _, req := range rep.allRequests() { //TODO: optimize
			if pb.EQ(req.Digest(), prePrep.Digest) {
				continue FOR_LOOP
			}
		}
		return nil
	}
	return prePreps
}

func (rep *Replica) prepBySequence(sequence uint64) ([]*pb.Entry, []*pb.Entry) {

	prePreps := rep.prePrepBySequence(sequence)

	if prePreps == nil {
		return nil, nil
	}

	var preps []*pb.Entry

FOR_LOOP:
	for _, prePrep := range prePreps {

		view := prePrep.View
		digest := prePrep.Digest

		replicas := make(map[uint64]int)
		for _, req := range rep.requests["prepare"] {
			reqPrepare := req.GetPrepare()
			v := reqPrepare.View
			s := reqPrepare.Sequence
			d := reqPrepare.Digest
			if v == view && s == sequence && pb.EQ(d, digest) {
				r := reqPrepare.Replica
				replicas[r]++
				if rep.twoThirds(replicas[r]) {
					prep := pb.ToEntry(s, d, v)
					preps = append(preps, prep)
					continue FOR_LOOP
				}
			}
		}
		return prePreps, nil
	}

	return prePreps, preps
}

/*
func (rep *Replica) prepBySequence(sequence uint64) *pb.Entry {
	var REQ *pb.Request
	var view uint64
	// Looking for a commit that
	// replica sent with sequence#
	for _, req := range rep.requests["commit"] {
		v := req.GetCommit().View
		s := req.GetCommit().Sequence
		r := req.GetCommit().Replica
		if v >= view && s == sequence && r == rep.ID {
			REQ = req
			view = v
		}
	}
	if REQ == nil {
		return nil
	}
	entry := pb.ToEntry(sequence, nil, view)
	return entry
}

func (rep *Replica) prePrepBySequence(sequence uint64) *pb.Entry {
	var REQ *pb.Request
	var view uint64
	// Looking for prepare/pre-prepare that
	// replica sent with matching sequence
	for _, req := range rep.requests["pre-prepare"] {
		v := req.GetPreprepare().View
		s := req.GetPreprepare().Sequence
		r := req.GetPreprepare().Replica
		if v >= view && s == sequence && r == rep.ID {
			REQ = req
			view = v
		}
	}
	for _, req := range rep.requests["prepare"] {
		v := req.GetPreprepare().View
		s := req.GetPreprepare().Sequence
		r := req.GetPreprepare().Replica
		if v >= view && s == sequence && r == rep.ID {
			REQ = req
			view = v
		}
	}
	if REQ == nil {
		return nil
	}
	var digest []byte
	switch REQ.Value.(type) {
	case *pb.Request_Preprepare:
		digest = REQ.GetPreprepare().Digest
	case *pb.Request_Prepare:
		digest = REQ.GetPrepare().Digest
	}
	entry := pb.ToEntry(sequence, digest, view)
	return entry
}
*/

func (rep *Replica) requestViewChange(view uint64) { // 缺少vc条件。

	if view != rep.view+1 {
		return
	}
	rep.view = view
	rep.activeView = false

	var prePreps []*pb.Entry
	var preps []*pb.Entry

	start := rep.lowWaterMark() + 1
	end := rep.highWaterMark()

	for s := start; s <= end; s++ {
		_prePreps, _preps := rep.prepBySequence(s)
		if _prePreps != nil {
			prePreps = append(prePreps, _prePreps...)
		}
		if _preps != nil {
			preps = append(preps, _preps...)
		}
	}

	sequence := rep.lowWaterMark()

	req := pb.ToRequestViewChange(
		view,
		sequence,
		rep.checkpoints, //
		preps,
		prePreps,
		rep.ID)

	rep.logRequest(req)

	go func() {
		rep.requestChan <- req
	}()

	rep.clearRequestsByView(view)
}

func (rep *Replica) createNewView(view uint64) (request *pb.Request) {

	// Returns RequestNewView if successful, else returns nil
	// create viewChanges
	viewChanges := make([]*pb.ViewChange, len(rep.pendingVC))

	for idx, _ := range viewChanges {
		req := rep.pendingVC[idx]
		viewchanger := req.GetViewchange().Replica
		vc := pb.ToViewChange(viewchanger, req.Digest())
		viewChanges[idx] = vc
	}

	var summaries []*pb.Summary
	var summary *pb.Summary

	start := rep.lowWaterMark() + 1
	end := rep.highWaterMark()

	// select starting checkpoint
FOR_LOOP_1:
	for seq := start; seq <= end; seq++ {

		overLWM := 0
		var digest []byte
		digests := make(map[string]int)

		for _, req := range rep.pendingVC {
			reqViewChange := req.GetViewchange()
			if reqViewChange.Sequence <= seq {
				overLWM++
			}
			for _, checkpoint := range reqViewChange.GetCheckpoints() {
				if checkpoint.Sequence == seq {
					d := checkpoint.Digest
					digests[string(d)]++
					if rep.oneThird(digests[string(d)]) {
						digest = d
						break
					}
				}
			}
			if rep.twoThirds(overLWM) && rep.oneThird(digests[string(digest)]) {
				summary = pb.ToSummary(seq, digest)
				continue FOR_LOOP_1
			}
		}
	}

	if summary == nil {
		return
	}

	summaries = append(summaries, summary)

	start = summary.Sequence
	end = start + CHECKPOINT_PERIOD*CONSTANT_FACTOR

	// select summaries
	// TODO: optimize
FOR_LOOP_2:
	for seq := start; seq <= end; seq++ {

		for _, REQ := range rep.pendingVC {

			sequence := REQ.GetViewchange().Sequence

			if sequence != seq {
				continue
			}

			var A1 []*pb.Request
			var A2 []*pb.Request

			view := REQ.GetViewchange().View
			digest := REQ.Digest()

		FOR_LOOP_3:
			for _, req := range rep.pendingVC {

				reqViewChange := req.GetViewchange()

				if reqViewChange.Sequence < sequence {
					preps := reqViewChange.GetPreps()
					for _, prep := range preps {
						if prep.Sequence != sequence {
							continue
						}
						if prep.View > view || (prep.View == view && !pb.EQ(prep.Digest, digest)) {
							continue FOR_LOOP_3
						}
					}
					A1 = append(A1, req)
				}
				prePreps := reqViewChange.GetPrepreps()
				for _, prePrep := range prePreps {
					if prePrep.Sequence != sequence {
						continue
					}
					if prePrep.View >= view && pb.EQ(prePrep.Digest, digest) {
						A2 = append(A2, req)
						continue FOR_LOOP_3
					}
				}
			}

			if rep.twoThirds(len(A1)) && rep.oneThird(len(A2)) {
				summary = pb.ToSummary(sequence, digest)
				summaries = append(summaries, summary)
				continue FOR_LOOP_2
			}
		}
	}

	request = pb.ToRequestNewView(view, viewChanges, summaries, rep.ID)
	return
}

func (rep *Replica) requestNewView(view uint64) {

	req := rep.createNewView(view)

	if req == nil || rep.hasRequest(req) {
		return
	}

	// Process new view

	success := rep.processNewView(req)

	if !success {
		return
	}

	rep.logRequest(req)

	go func() {
		rep.requestChan <- req
	}()

}