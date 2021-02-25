package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/pipapa/pbft/cmd"
	"github.com/pipapa/pbft/message"
)

const (
	RequestEntry    = "/request"
	PrePrepareEntry = "/preprepare"
	PrepareEntry    = "/prepare"
	CommitEntry     = "/commit"
)

type HttpServer struct {
	port   int
	server *http.Server

	requestRecv    chan *message.Request
	prePrepareRecv chan *message.PrePrepare
	prepareRecv    chan *message.Prepare
	commit         chan *message.Commit
}

func NewServer(cfg *cmd.SharedConfig) *HttpServer {
	httpServer := &HttpServer{
		port:   cfg.Port,
		server: nil,
	}
	return httpServer
}

func (s *HttpServer) RegisterChan(r chan *message.Request, pre chan *message.PrePrepare,
	p chan *message.Prepare, c chan *message.Commit) {
	log.Printf("[Server] register the chan for listen func")
	s.requestRecv = r
	s.prePrepareRecv = pre
	s.prepareRecv = p
	s.commit = c
}

func (s *HttpServer) Run() {
	log.Printf("[Node] start the listen server")
	s.registerServer()
}

func (s *HttpServer) registerServer() {
	log.Printf("[Server] set listen port:%d\n", s.port)

	httpRegister := map[string]func(http.ResponseWriter, *http.Request){
		RequestEntry:    s.HttpRequest,
		PrePrepareEntry: s.HttpPrePrepare,
		PrepareEntry:    s.HttpPrepare,
		CommitEntry:     s.HttpCommit,
	}

	mux := http.NewServeMux()
	for k, v := range httpRegister {
		log.Printf("[Server] register the func for %s", k)
		mux.HandleFunc(k, v)
	}

	s.server = &http.Server{
		Addr:    ":" + strconv.Itoa(s.port),
		Handler: mux,
	}

	if err := s.server.ListenAndServe(); err != nil {
		log.Printf("[Server Error] %s", err)
		return
	}
}
