package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

const (
	authPath = "/sfslack/auth"
)

func (srv *server) authCallback(wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	wfID, err := workflowID(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	builder, ok := srv.wfCache.GetBuilder(wfID)
	if !ok {
		http.Error(wr, "Unknown workflow ID", http.StatusInternalServerError)
		return
	}

	login, err := srv.authCache.Add(builder.Cmd.User, req.URL.Query())
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	builder.Sf = login

	redirectURL := srv.startWorkflowForRedirect(builder)
	if redirectURL == "" {
		wr.Write([]byte("Logged in! You may close this page."))
		return
	}
	http.Redirect(wr, req, redirectURL, http.StatusFound)
}

func authCallbackURL(host string, wfID int) string {
	return fmt.Sprintf("https://%v%v?%v=%v",
		host,
		authPath,
		wfidQueryKey,
		wfID)
}
