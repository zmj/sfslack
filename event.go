package main

import "net/http"
import "net/http/httputil"
import "fmt"

const (
	eventPath = "/sfslack/event"
)

func (srv *server) eventCallback(wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	// parse event
	// get wfid
	// get wf
	// wf.event()
}
