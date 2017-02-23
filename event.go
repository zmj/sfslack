package main

import "net/http"

const (
	eventPath = "/sfslack/event"
)

func (srv *server) eventCallback(wr http.ResponseWriter, req *http.Request) {
	// parse event
	// get wfid
	// get wf
	// wf.event()
}
