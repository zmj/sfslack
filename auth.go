package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/workflow"
)

const (
	authPath         = "/sfslack/auth"
	publicHostHeader = "X-PUBLIC-HOST"
	wfidQueryKey     = "wfid"
)

func (srv *server) authCallback(wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	wf, status, err := srv.getWorkflow(req)
	if err != nil {
		http.Error(wr, err.Error(), status)
		return
	}
	login, err := srv.authCache.Add(wf.Cmd().User, req.URL.Query())
	if err != nil {
		http.Error(wr, err.Error(), http.StatusNotFound)
		return
	}
	redirectURL := startWorkflowForRedirect(wf, login)
	if redirectURL == "" {
		wr.Write([]byte("Logged in! You may close this page."))
		return
	}
	http.Redirect(wr, req, redirectURL, http.StatusFound)
}

func startWorkflowForRedirect(wf workflow.Workflow, login sharefile.Login) string {
	redirect := make(chan string, 1)
	accepted := make(chan error, 1)
	cb := func(url string) error {
		redirect <- url
		return <-accepted
	}
	go wf.Start(login, nil, cb)
	select {
	case url := <-redirect:
		accepted <- nil
		return url
	case <-time.After(3 * time.Second):
		accepted <- errors.New("Timed out")
		return ""
	}
}

func (srv *server) authCallbackURL(req *http.Request, wfID int) string {
	host := req.Header.Get(publicHostHeader)
	if host == "" {
		host = req.URL.Host
	}
	return fmt.Sprintf("https://%v%v?%v=%v",
		host,
		authPath,
		wfidQueryKey,
		wfID)
}
