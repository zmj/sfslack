package server

import "net/http"
import "net/http/httputil"
import "fmt"

const (
	eventPath = "/sfslack/event"
)

func (srv *server) eventCallback(wf *runner, wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	// parse event
	// get wfid
	// get wf
	// wf.event()
}

func eventCallbackURL(host string, wfID int) string {
	return fmt.Sprintf("https://%v%v?%v=%v",
		host,
		eventPath,
		wfidQueryKey,
		wfID)
}
