package server

import (
	"fmt"
	"net/http"
	"sync"

	"gogw/common"
	"gogw/logger"
	"gogw/schema"
)

type Server struct {
	ServerAddr    string
	Clients *sync.Map
}

func NewServer(serverAddr string) *Server {
	return &Server{
		ServerAddr:    serverAddr,
		Clients:       &sync.Map{},
	}
}

//client register
func (s *Server) registerHandler(w http.ResponseWriter, req *http.Request) {
	defer func(){
		req.Body.Close()
	}()

	msgPack, err := schema.ReadMsg(req.Body)
	if err != nil {
		return
	}

	msg, ok := msgPack.Msg.(*schema.RegisterRequest)
	if ! ok {
		return
	}

	clientId := common.UUID("clientid")

	client := NewClient(
		clientId,
		req.RemoteAddr,
		msg.ToPort,
		msg.Direction,
		msg.Protocol,
		msg.SourceAddr,
		msg.Description,
	)

	s.Clients.Store(clientId, client)
	err = client.Start()
	logger.Error(err)
}

//client register
func (s *Server) msgHandler(w http.ResponseWriter, req *http.Request) {
	defer func(){
		req.Body.Close()
	}()

	if its, ok := req.URL.Query()["clientid"]; ok && len(its[0]) > 0 {
		clientId := its[0]
		if ci, ok := s.Clients.Load(clientId); ok {
			client, _ := ci.(*Client)
			client.HttpHandler(w, req)
		}
	}
}


func (s *Server) Start() {
	logger.Info(fmt.Sprintf("\nserver start\nAddr: %v\n", s.ServerAddr))

	http.HandleFunc("/register", s.registerHandler)
	http.HandleFunc("/msg", s.msgHandler)
	http.HandleFunc("/monitor", s.monitorHandler)
	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("./ui"))))
	http.ListenAndServe(s.ServerAddr, nil)
}
