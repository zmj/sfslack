package server

import (
	"fmt"
	"net/http"

	"github.com/zmj/sfslack/server/wfhost"
	"github.com/zmj/sfslack/sharefile"
)

func (srv *server) eventCallback(wf *wfhost.Runner, wr http.ResponseWriter, req *http.Request) {
	event, err := sharefile.ParseEvent(req.Body)
	if err != nil {
		err = fmt.Errorf("Failed to parse event: %v", err)
		srv.logErr(err)
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("a")
	wf.Event(event)
	fmt.Println("b")
}
