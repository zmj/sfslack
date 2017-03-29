package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zmj/sfslack/server/wfhost"
)

const (
	redirectTimeout = 3 * time.Second
)

func (srv *server) redirect(wf *wfhost.Runner, wr http.ResponseWriter, req *http.Request) {
	type redirect struct {
		url string
		err error
	}
	redir := make(chan redirect, 1)
	accept := make(chan bool, 1)
	cb := func(url string, err error) bool {
		redir <- redirect{url, err}
		return <-accept
	}
	wf.NextRedirect(cb)
	select {
	case r := <-redir:
		accept <- true
		if r.err != nil {
			wr.Write([]byte(wf.ErrorText(r.err)))
			return
		}
		if r.url == "" {
			wr.Write([]byte("Done! You may close this page."))
			return
		}
		http.Redirect(wr, req, r.url, http.StatusFound)
	case <-time.After(redirectTimeout):
		accept <- false
		wr.Write([]byte(waitHTML(wf.WaitingURL())))
	}
}

const (
	waitMessage = "Working on it..."
	waitHTMLfmt = `<html><head>
<meta http-equiv="refresh" content="%v; url=%v">
</head><body>
%v
</body></html>`
)

func waitHTML(nextURL string) string {
	s := fmt.Sprintf(waitHTMLfmt,
		0,
		nextURL,
		waitMessage)
	return s
}
