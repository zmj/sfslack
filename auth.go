package main

import "net/http"
import "net/http/httputil"
import "fmt"

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
	go wf.Start(login, nil)
	// redirect callback
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
