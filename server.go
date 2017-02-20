package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/workflow"
)

const (
	listenport = 8080
)

func main() {
	srv := newServer()
	err := (&http.Server{
		Addr:    fmt.Sprintf(":%v", listenport),
		Handler: srv.handler(),
	}).ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}

func (srv *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/sfslack/command", srv.newCommand)
	mux.HandleFunc("/sfslack/auth", srv.authCallback)
	mux.HandleFunc("/sfslack/event", srv.eventCallback)
	return mux
}

type server struct {
	mu                *sync.Mutex
	workflows         map[int]workflow.Workflow
	currentWorkflowID int
	authCache         *sharefile.AuthCache
}

func newServer() *server {
	return &server{
		mu:        &sync.Mutex{},
		workflows: make(map[int]workflow.Workflow),
		authCache: sharefile.NewAuthCache(),
	}
}

func (srv *server) nextWorkflowID() int {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.currentWorkflowID += 1
	return srv.currentWorkflowID
}
