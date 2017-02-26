package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"

	"strconv"

	"errors"

	"github.com/zmj/sfslack/secrets"
	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

const (
	listenport = 8080

	publicHostHeader = "X-PUBLIC-HOST"
	wfidQueryKey     = "wfid"
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

func (srv *server) newWorkflow(cmd slack.Command) (workflow.Workflow, int, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.currentWorkflowID++
	wf, err := workflow.NewWorkflow(cmd, "")
	if err != nil {
		return nil, -1, err
	}
	srv.workflows[srv.currentWorkflowID] = wf
	return wf, srv.currentWorkflowID, nil
}

func (srv *server) getWorkflow(req *http.Request) (workflow.Workflow, int, error) {
	values, err := httpValues(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	wfidStr := values.Get(wfidQueryKey)
	if wfidStr == "" {
		return nil, http.StatusBadRequest, err
	}
	wfid, err := strconv.Atoi(wfidStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("Invalid wfID")
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	wf, exists := srv.workflows[wfid]
	if !exists {
		return nil, http.StatusNotFound, errors.New("Workflow not found")
	}
	return wf, http.StatusOK, nil
}

func publicHost(req *http.Request) string {
	host := req.Header.Get(publicHostHeader)
	if host == "" {
		host = req.URL.Host
	}
	return host
}

func printReq(wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))
	http.Error(wr, "", http.StatusNotFound)
}
