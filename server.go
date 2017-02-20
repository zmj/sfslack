package main

import (
	"fmt"
	"net/http"
	"sync"

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
	mu        *sync.Mutex
	workflows map[int]workflow.Workflow
}

func newServer() *server {
	return &server{
		mu:        &sync.Mutex{},
		workflows: make(map[int]workflow.Workflow),
	}
}
