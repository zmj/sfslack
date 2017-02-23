package main

import "net/http"
import "fmt"

const (
	authPath         = "/sfslack/auth"
	publicHostHeader = "X-PUBLIC-HOST"
	wfidQueryKey     = "wfid"
)

func (srv *server) authCallback(wr http.ResponseWriter, req *http.Request) {
	// parse auth
	// get wfid
	// wfid -> user
	// user -> all wfs
	// cache.add(user, login)
	// wfid: wf.auth(login, redirectCallback)
	// not wfid: wf.auth(login, redirectCallback: nil)

	// simplify - don't share pending
	// wfid -> user
	// login := cache.add(user, values)
	// wf.auth(login, redirectCallback)
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
