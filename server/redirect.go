package server

import (
	"fmt"
	"net/http"
	"time"
)

const (
	redirectTimeout = 3 * time.Second
	redirectPath    = "/sfslack/wait"
)

type redirect struct {
	done chan struct{}
	url  string
	err  error
}

func (srv *server) redirect(wf *runner, wr http.ResponseWriter, req *http.Request) {
	var url string
	redir := wf.Redirect()
	select {
	case <-redir.done:
		if redir.err != nil {
			http.Error(wr, redir.err.Error(), http.StatusInternalServerError)
			return
		}
		url = redir.url
	case <-time.After(redirectTimeout):
		url = waitURL(publicHost(req), wf.wfID)
	}

	if url == "" {
		wr.Write([]byte("Done! You may close this page."))
		return
	}
	http.Redirect(wr, req, url, http.StatusFound)
}

func waitURL(host string, wfID int) string {
	return fmt.Sprintf("https://%v%v?%v=%v",
		host,
		redirectPath,
		wfidQueryKey,
		wfID)
}
