package raft

import (
    "bytes"
    "../labgob"
   // "log"
    "math/rand"
    "sort"
    "sync"
    "sync/atomic"
	"time"
	"fmt"
)
import "../labrpc"

type ApplyMsg struct {
    CommandValid bool
    Command      interface{}
    CommandIndex int
}

type State int
const (
    Follower State = iota // value --> 0
    Candidate             // value --> 1
    Leader                // value --> 2
)

const NULL int = -1

type Log struct {
    Term    int         "term when entry was received by leader"
    Command interface{} "command for state machine,"
}
// A Go object implementing a single Raft peer.
type Raft struct {
    mu        sync.Mutex          // Lock to protect shared access to this peer's state
    peers     []*labrpc.ClientEnd // RPC end points of all peers
    persister *Persister          // Object to hold this peer's persisted state
    me        int                 // this peer's index into peers[]

    // state a Raft server must maintain.
    state     State

    //Persistent state on all servers:(Updated on stable storage before responding to RPCs)
    currentTerm int    "latest term server has seen (initialized to 0 increases monotonically)"
    votedFor    int    "candidateId that received vote in current term (or null if none)"
    log         []Log  "log entries;(first index is 1)"

    //Volatile state on all servers:
    commitIndex int    "index of highest log entry known to be committed (initialized to 0, increases monotonically)"
    lastApplied int    "index of highest log entry applied to state machine (initialized to 0, increases monotonically)"

    //Volatile state on leaders：(Reinitialized after election)
    nextIndex   []int  "for each server,index of the next log entry to send to that server"
    matchIndex  []int  "for each server,index of highest log entry known to be replicated on server(initialized to 0, im)"

    //channel
    applyCh     chan ApplyMsg // from Make()
    killCh      chan bool //for Kill()
    //handle rpc
    voteCh      chan bool
    appendLogCh chan bool

}

func  Min(a int, b int) int {
	if a > b{
		return b
	}
	return a
}

// return currentTerm and whether this server believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
    var term int
    var isleader bool
    rf.mu.Lock()
    defer rf.mu.Unlock()
    term = rf.currentTerm
    isleader = (rf.state == Leader)
    return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
func (rf *Raft) persist() {
    w := new(bytes.Buffer)
    e := labgob.NewEncoder(w)
    e.Encode(rf.currentTerm)
    e.Encode(rf.votedFor)
    e.Encode(rf.log)
    data := w.Bytes()
    rf.persister.SaveRaftState(data)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
    if data == nil || len(data) < 1 { // bootstrap without any state?
        return
    }
    r := bytes.NewBuffer(data)
    d := labgob.NewDecoder(r)
    var currentTerm int
    var voteFor int
    var clog []Log
    if  d.Decode(&currentTerm) != nil || d.Decode(&voteFor) != nil || d.Decode(&clog) != nil {
        //log.Fatal("readPersist ERROR for server %v",rf.me)
    } else {
        rf.mu.Lock()
        rf.currentTerm, rf.votedFor, rf.log = currentTerm, voteFor, clog
        rf.mu.Unlock()
    }
}

// RequestVote RPC arguments structure. field names must start with capital letters!
type RequestVoteArgs struct {
    Term            int "candidate’s term"
    CandidateId     int "candidate requesting vote"
    LastLogIndex    int "index of candidate’s last log entry (§5.4)"
    LastLogTerm     int "term of candidate’s last log entry (§5.4)"
}

// RequestVote RPC reply structure. field names must start with capital letters!
type RequestVoteReply struct {
    Term        int  "currentTerm, for candidate to update itself"
    VoteGranted bool "true means candidate received vote"
}

//RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
    rf.mu.Lock()
    defer rf.mu.Unlock()
    if (args.Term > rf.currentTerm) {//all server rule 1 If RPC request or response contains term T > currentTerm:
        rf.beFollower(args.Term) // set currentTerm = T, convert to follower (§5.1)
    }
    reply.Term = rf.currentTerm
    reply.VoteGranted = false
    if (args.Term < rf.currentTerm) || (rf.votedFor != NULL && rf.votedFor != args.CandidateId) {
        // Reply false if term < currentTerm (§5.1)  If votedFor is not null and not candidateId,
    } else if args.LastLogTerm < rf.getLastLogTerm() || (args.LastLogTerm == rf.getLastLogTerm() && args.LastLogIndex < rf.getLastLogIdx()){
        //If the logs have last entries with different terms, then the log with the later term is more up-to-date.
        // If the logs end with the same term, then whichever log is longer is more up-to-date.
        // Reply false if candidate’s log is at least as up-to-date as receiver’s log
    } else {
        //grant vote
        rf.votedFor = args.CandidateId
        reply.VoteGranted = true
        rf.state = Follower
        rf.persist()
        send(rf.voteCh) //because If election timeout elapses without receiving granting vote to candidate, so wake up

    }
}

////RequestVote RPC sender.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
    ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
    return ok
}

type AppendEntriesArgs struct {
    Term         int    "leader’s term"
    LeaderId     int    "so follower can redirect clients"
    PrevLogIndex int    "index of log entry immediately preceding new ones"
    PrevLogTerm  int    "term of prevLogIndex entry"
    Entries      []Log  "log entries to store (empty for heartbeat;may send more than one for efficiency)"
    LeaderCommit int    "leader’s commitIndex"
}

type AppendEntriesReply struct {
    Term          int   "currentTerm, for leader to update itself"
    Success       bool  "true if follower contained entry matching prevLogIndex and prevLogTerm"
    ConflictIndex int   "the first index it stores for that conflict term"
    ConflictTerm  int   "the term of the conflicting entry"
}

//AppendEntries RPC handler.
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {//now only for heartbeat
    rf.mu.Lock()
    defer rf.mu.Unlock()
    defer send(rf.appendLogCh) //If election timeout elapses without receiving AppendEntries RPC from current leader
    if args.Term > rf.currentTerm { //all server rule 1 If RPC request or response contains term T > currentTerm:
        rf.beFollower(args.Term) // set currentTerm = T, convert to follower (§5.1)
    }
    reply.Term = rf.currentTerm
    reply.Success = false
    reply.ConflictTerm = NULL
    reply.ConflictIndex = 0
    //1. Reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm (§5.3)
    prevLogIndexTerm := -1
    logSize := len(rf.log)
    if args.PrevLogIndex >= 0 && args.PrevLogIndex < len(rf.log) {
        prevLogIndexTerm = rf.log[args.PrevLogIndex].Term
    }
    if prevLogIndexTerm != args.PrevLogTerm {
        reply.ConflictIndex = logSize
        if prevLogIndexTerm == -1 {//If a follower does not have prevLogIndex in its log,
            //it should return with conflictIndex = len(log) and conflictTerm = None.
        } else { //If a follower does have prevLogIndex in its log, but the term does not match
            reply.ConflictTerm = prevLogIndexTerm //it should return conflictTerm = log[prevLogIndex].Term,
            i := 0
            for ; i < logSize; i++ {//and then search its log for
                if rf.log[i].Term == reply.ConflictTerm {//the first index whose entry has term equal to conflictTerm
                    reply.ConflictIndex = i
                    break
                }
            }
        }
        return
    }
    //2. Reply false if term < currentTerm (§5.1)
    if args.Term < rf.currentTerm {return}

    index := args.PrevLogIndex
    for i := 0; i < len(args.Entries); i++ {
        index++
        if index < logSize {
            if rf.log[index].Term == args.Entries[i].Term {
                continue
            } else {//3. If an existing entry conflicts with a new one (same index but different terms),
                rf.log = rf.log[:index]//delete the existing entry and all that follow it (§5.3)
            }
        }
        rf.log = append(rf.log,args.Entries[i:]...) //4. Append any new entries not already in the log
        rf.persist()
        break;
    }
    //5. If leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
    if args.LeaderCommit > rf.commitIndex {
        rf.commitIndex = Min(args.LeaderCommit ,rf.getLastLogIdx())
        rf.updateLastApplied()
    }
    reply.Success = true
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
    ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
    return ok
}


//Leader Section:
func (rf *Raft) startAppendLog() {
    for i := 0; i < len(rf.peers); i++ {
        if i == rf.me {
            continue
        }
        go func(idx int) {
            for {
                rf.mu.Lock();
                if rf.state != Leader {
                    rf.mu.Unlock()
                    return
                } //send initial empty AppendEntries RPCs (heartbeat) to each server
                args := AppendEntriesArgs{
                    rf.currentTerm,
                    rf.me,
                    rf.getPrevLogIdx(idx),
                    rf.getPrevLogTerm(idx),
                    //If last log index ≥ nextIndex for a follower:send AppendEntries RPC with log entries starting at nextIndex
                    //nextIndex > last log index, rf.log[rf.nextIndex[idx]:] will be empty then like a heartbeat
                    append(make([]Log,0),rf.log[rf.nextIndex[idx]:]...),
                    rf.commitIndex,
                }
                rf.mu.Unlock()
                reply := &AppendEntriesReply{}
				ret := rf.sendAppendEntries(idx, &args, reply)
                rf.mu.Lock();
                if !ret || rf.state != Leader || rf.currentTerm != args.Term {
                    rf.mu.Unlock()
                    return
                }
                if reply.Term > rf.currentTerm {//all server rule 1 If RPC response contains term T > currentTerm:
                    rf.beFollower(reply.Term) // set currentTerm = T, convert to follower (§5.1)
                    rf.mu.Unlock()
                    return
                }
                if reply.Success {//If successful：update nextIndex and matchIndex for follower
                    rf.matchIndex[idx] = args.PrevLogIndex + len(args.Entries)
                    rf.nextIndex[idx] = rf.matchIndex[idx] + 1
                    rf.updateCommitIndex()
                    rf.mu.Unlock()
                    return
                } else { //If AppendEntries fails because of log inconsistency: decrement nextIndex and retry
                    tarIndex := reply.ConflictIndex //If it does not find an entry with that term
                    if reply.ConflictTerm != NULL {
                        logSize := len(rf.log) //first search its log for conflictTerm
                        for i := 0; i < logSize; i++ {//if it finds an entry in its log with that term,
                            if rf.log[i].Term != reply.ConflictTerm { continue }
                            for i < logSize && rf.log[i].Term == reply.ConflictTerm { i++ }//set nextIndex to be the one
                            tarIndex = i //beyond the index of the last entry in that term in its log
                        }
                    }
                    rf.nextIndex[idx] = tarIndex;
                    rf.mu.Unlock()
				}
				
			}	
			


        }(i)
    }
}

// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
    rf.mu.Lock()
    defer rf.mu.Unlock()
    index := -1
    term := rf.currentTerm
    isLeader := (rf.state == Leader)
    //If command received from client: append entry to local log, respond after entry applied to state machine (§5.3)
    if isLeader {
        index = rf.getLastLogIdx() + 1
        newLog := Log{
            rf.currentTerm,
            command,
        }
        rf.log = append(rf.log,newLog)
        rf.persist()
    }
    return index, term, isLeader
}

//If there exists an N such that N > commitIndex,
// a majority of matchIndex[i] ≥ N, and log[N].term == currentTerm: set commitIndex = N (§5.3, §5.4).
func (rf *Raft) updateCommitIndex() {
    rf.matchIndex[rf.me] = len(rf.log) - 1
    copyMatchIndex := make([]int,len(rf.matchIndex))
    copy(copyMatchIndex,rf.matchIndex)
    sort.Sort(sort.Reverse(sort.IntSlice(copyMatchIndex)))
    N := copyMatchIndex[len(copyMatchIndex)/2]
    if N > rf.commitIndex && rf.log[N].Term == rf.currentTerm {
        rf.commitIndex = N
        rf.updateLastApplied()
    }
}

func (rf *Raft) beLeader() {
    if rf.state != Candidate {
        return
	}
	fmt.Println(rf.me, " Become Leader !", rf.currentTerm)
    rf.state = Leader
    //initialize leader data
    rf.nextIndex = make([]int,len(rf.peers))
    rf.matchIndex = make([]int,len(rf.peers))//initialized to 0
    for i := 0; i < len(rf.nextIndex); i++ {//(initialized to leader last log index + 1)
        rf.nextIndex[i] = rf.getLastLogIdx() + 1
    }
}
//end Leader section


//Candidate Section:
// If AppendEntries RPC received from new leader: convert to follower implemented in AppendEntries RPC Handler
func (rf *Raft) beCandidate() { //Reset election timer are finished in caller
	fmt.Println(rf.me,"become Candidate", rf.currentTerm)
	rf.state = Candidate
    rf.currentTerm++ //Increment currentTerm
    rf.votedFor = rf.me //vote myself first
    rf.persist()
    //ask for other's vote
    go rf.startElection() //Send RequestVote RPCs to all other servers
}
//If election timeout elapses: start new election handled in caller
func (rf *Raft) startElection() {
    rf.mu.Lock()
    args := RequestVoteArgs{
        rf.currentTerm,
        rf.me,
        rf.getLastLogIdx(),
        rf.getLastLogTerm(),

    };
    rf.mu.Unlock()
    var votes int32 = 1;
    for i := 0; i < len(rf.peers); i++ {
        if i == rf.me {
            continue
        }
        go func(idx int) {
            reply := &RequestVoteReply{}
            ret := rf.sendRequestVote(idx,&args,reply)

            if ret {
                rf.mu.Lock()
                defer rf.mu.Unlock()
                if reply.Term > rf.currentTerm {
                    rf.beFollower(reply.Term)
                    return
                }
                if rf.state != Candidate || rf.currentTerm != args.Term{
                    return
                }
                if reply.VoteGranted {
                    atomic.AddInt32(&votes,1)
                } //If votes received from majority of servers: become leader
                if atomic.LoadInt32(&votes) > int32(len(rf.peers) / 2) {
                    rf.beLeader()
                    send(rf.voteCh) //after be leader, then notify 'select' goroutine will sending out heartbeats immediately
                }
            }
        }(i)
    }
}
//end Candidate section

//Follower Section:
func (rf *Raft) beFollower(term int) {
    rf.state = Follower
    rf.votedFor = NULL
    rf.currentTerm = term
    rf.persist()
}
//end Follower section

//all server rule : If commitIndex > lastApplied: increment lastApplied, apply log[lastApplied] to state machine
func (rf *Raft) updateLastApplied() {
    for rf.lastApplied < rf.commitIndex {
        rf.lastApplied++
        curLog := rf.log[rf.lastApplied]
        applyMsg := ApplyMsg{
            true,
            curLog.Command,
            rf.lastApplied,
        }
        rf.applyCh <- applyMsg
    }
}


// the tester calls Kill() when a Raft instance won't be needed again.
func (rf *Raft) Kill() {
    send(rf.killCh)
}

//Helper function
func send(ch chan bool) {
    select {
    case <-ch: //if already set, consume it then resent to avoid block
    default:
    }
    ch <- true
}

func (rf *Raft) getPrevLogIdx(i int) int {
    return rf.nextIndex[i] - 1
}

func (rf *Raft) getPrevLogTerm(i int) int {
    prevLogIdx := rf.getPrevLogIdx(i)
    if prevLogIdx < 0 {
        return -1
    }
    return rf.log[prevLogIdx].Term
}

func (rf *Raft) getLastLogIdx() int {
    return len(rf.log) - 1
}

func (rf *Raft) getLastLogTerm() int {
    idx := rf.getLastLogIdx()
    if idx < 0 {
        return -1
    }
    return rf.log[idx].Term
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
    persister *Persister, applyCh chan ApplyMsg) *Raft {
    rf := &Raft{}
    rf.peers = peers
    rf.persister = persister
    rf.me = me

    rf.state = Follower
    rf.currentTerm = 0
    rf.votedFor = NULL
    rf.log = make([]Log,1) //(first index is 1)

    rf.commitIndex = 0
    rf.lastApplied = 0

    rf.applyCh = applyCh
    //because gorountne only send the chan to below goroutine,to avoid block, need 1 buffer
    rf.voteCh = make(chan bool,1)
    rf.appendLogCh = make(chan bool,1)
    rf.killCh = make(chan bool,1)
    // initialize from state persisted before a crash
    rf.readPersist(persister.ReadRaftState())

    //because from hint The tester requires that the leader send heartbeat RPCs no more than ten times per second.
    heartbeatTime := time.Duration(100) * time.Millisecond

    //from hint :You'll need to write code that takes actions periodically or after delays in time.
    //  The easiest way to do this is to create a goroutine with a loop that calls time.Sleep().
    go func() {
        for {
            select {
            case <-rf.killCh:
                return
            default:
            }
            electionTime := time.Duration(rand.Intn(200) + 300) * time.Millisecond

            rf.mu.Lock()
            state := rf.state
            rf.mu.Unlock()

            switch state {
            case Follower, Candidate:
                select {
                case <-rf.voteCh:
                case <-rf.appendLogCh:
                case <-time.After(electionTime):
                    rf.mu.Lock()
                    rf.beCandidate() //becandidate, Reset election timer, then start election
                    rf.mu.Unlock()
                }
            case Leader:
                rf.startAppendLog()
                time.Sleep(heartbeatTime)
            }
        }
    }()
    return rf
}