package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"

	"github.com/zmj/sfslack/secrets"
	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

const (
	listenport = 8080
)

func main() {
	secrets, err := secrets.Load()
	if err != nil {
		fmt.Println(err)
		return
	}
	srv := newServer(secrets)
	err = (&http.Server{
		Addr:    fmt.Sprintf(":%v", listenport),
		Handler: srv.handler(),
	}).ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (srv *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", printReq)
	mux.HandleFunc(commandPath, srv.newCommand)
	mux.HandleFunc(authPath, srv.authCallback)
	mux.HandleFunc(eventPath, srv.eventCallback)
	return mux
}

type server struct {
	mu                *sync.Mutex
	workflows         map[int]workflow.Workflow
	currentWorkflowID int
	authCache         *sharefile.AuthCache
}

func newServer(secrets secrets.Secrets) *server {
	return &server{
		mu:        &sync.Mutex{},
		workflows: make(map[int]workflow.Workflow),
		authCache: sharefile.NewAuthCache(secrets.OAuthID, secrets.OAuthSecret),
	}
}

func (srv *server) newWorkflow(cmd slack.Command) (workflow.Workflow, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.currentWorkflowID += 1
	return workflow.NewWorkflow(cmd, srv.currentWorkflowID)
}

func printReq(wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))
	http.Error(wr, "", http.StatusNotFound)
}
