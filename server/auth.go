package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func (srv *server) authCallback(wf *runner, wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	wf.SetLogin(req.URL.Query())
	srv.redirect(wf, wr, req)
}
