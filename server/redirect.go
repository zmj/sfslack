package server

import (
	"fmt"
	"net/http"
	"time"
)

const (
	redirectTimeout = 3 * time.Second
)

func (srv *server) redirect(wf *runner, wr http.ResponseWriter, req *http.Request) {
	redir := make(chan string, 1)
	accept := make(chan bool, 1)
	cb := func(url string) bool {
		redir <- url
		return <-accept
	}
	wf.NextRedirect(cb)
	var url string
	select {
	case url = <-redir:
		accept <- true
		if url == "" {
			wr.Write([]byte("Done! You may close this page."))
			return
		}
		http.Redirect(wr, req, url, http.StatusFound)
	case <-time.After(redirectTimeout):
		accept <- false
		wr.Write([]byte(waitHTML(wf.urls.Waiting)))
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
	fmt.Println(s)
	return s
}
