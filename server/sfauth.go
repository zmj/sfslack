package server

import (
	"net/http"

	"github.com/zmj/sfslack/server/wfhost"
)

func (srv *server) sfAuth(wf *wfhost.Runner, wr http.ResponseWriter, req *http.Request) {
	wf.SetLogin(req.URL.Query())
	srv.redirect(wf, wr, req)
}
