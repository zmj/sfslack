package main

import "net/http"

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
	// cache.add(user, login)
	// wf.auth(login, redirectCallback)
}
