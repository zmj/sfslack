package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/zmj/sfslack/server/wfhost"
	"github.com/zmj/sfslack/sharefile"
)

func (srv *server) eventCallback(wf *wfhost.Runner, wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	event, err := sharefile.ParseEvent(req.Body)
	if err != nil {
		err = fmt.Errorf("Failed to parse event: %v", err)
		srv.logErr(err)
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(event.WebhookSubscriptionID)
	fmt.Println(event.AccountID)
	fmt.Println(event.Event.ResourceType)
	fmt.Println(event.Event.OperationName)
	fmt.Println(event.Event.Timestamp)
	fmt.Println(event.Event.Resource.ID)
	fmt.Println(event.Event.Resource.Parent.ID)

	// hand to wf
}
