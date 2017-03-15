package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

const (
	authPath = "/sfslack/auth"
)

func (srv *server) authCallback(wf *runner, wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	login, err := srv.authCache.Add(wf.cmd.User, req.URL.Query())
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	wf.SetLogin(login)
	srv.redirect(wf, wr, req)
}

func authCallbackURL(host string, wfID int) string {
	return fmt.Sprintf("https://%v%v?%v=%v",
		host,
		authPath,
		wfidQueryKey,
		wfID)
}
