package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/zmj/sfslack/server/wfhost"
)

func (srv *server) eventCallback(wf *wfhost.Runner, wr http.ResponseWriter, req *http.Request) {
	fmt.Println("")
	fmt.Println("event!")
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	// parse event
	// get wfid
	// get wf
	// wf.event()
}
