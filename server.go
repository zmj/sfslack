package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"strconv"

	"time"

	"github.com/zmj/sfslack/secrets"
	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/wfutils"
	"github.com/zmj/sfslack/workflow"
)

const (
	publicHostHeader = "X-PUBLIC-HOST"
	wfidQueryKey     = "wfid"
	wfTypeQueryKey   = "wftype"
	redirectTimeout  = 3 * time.Second
)

type server struct {
	authCache *sharefile.AuthCache
	wfCache   *wfutils.Cache
}

func newServer(secrets secrets.Secrets) *server {
	return &server{
		authCache: sharefile.NewAuthCache(secrets.OAuthID, secrets.OAuthSecret),
		wfCache:   wfutils.NewCache(),
	}
}

func (srv *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", printReq)
	mux.HandleFunc(commandPath, srv.newCommand)
	mux.HandleFunc(commandClickPath, srv.newCommandClick)
	mux.HandleFunc(authPath, srv.authCallback)
	mux.HandleFunc(eventPath, srv.eventCallback)
	return mux
}

func (srv *server) startWorkflowForRedirect(builder *wfutils.Builder) string {
	resp, ok := srv.startWorkflow(builder, redirectTimeout)
	if !ok {
		return ""
	}
	return resp.URL
}

func (srv *server) startWorkflowForMessage(builder *wfutils.Builder) slack.Message {
	resp, ok := srv.startWorkflow(builder, slack.InitialReplyTimeout)
	if !ok {
		return timeoutMessage()
	}
	return resp.Msg
}

// move to runner?
func (srv *server) startWorkflow(builder *wfutils.Builder, timeout time.Duration) (workflow.Response, bool) {
	runner := wfutils.NewRunner(builder, srv.wfCache)
	response := make(chan workflow.Response, 1)
	accepted := make(chan bool, 1)
	cb := func(resp workflow.Response) bool {
		response <- resp
		return <-accepted
	}
	go func() {
		err := runner.Run(cb)
		srv.logErr(err)
		// cleanup wf id
	}()
	select {
	case resp := <-response:
		accepted <- true
		return resp, true
	case <-time.After(timeout):
		accepted <- false
		return workflow.Response{}, false
	}
}

func timeoutMessage() slack.Message {
	return slack.Message{Text: "Logging you in..."}
}

func workflowID(req *http.Request) (int, error) {
	values, err := httpValues(req)
	if err != nil {
		return 0, errors.New("Missing wfID")
	}
	wfidStr := values.Get(wfidQueryKey)
	if wfidStr == "" {
		return 0, errors.New("Missing wfID")
	}
	wfID, err := strconv.Atoi(wfidStr)
	if err != nil {
		return wfID, errors.New("Invalid wfID")
	}
	return wfID, nil
}

func httpValues(req *http.Request) (url.Values, error) {
	if req.Method == "GET" {
		return req.URL.Query(), nil
	} else if req.Method == "POST" {
		err := req.ParseForm()
		if err != nil {
			return url.Values{}, err
		}
		return req.PostForm, nil
	} else {
		return url.Values{}, fmt.Errorf("Unsupported HTTP method '%v'", req.Method)
	}
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

func (srv *server) logErr(err error) {
	if err == nil {
		return
	}
	// todo
}
