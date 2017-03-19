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
	fmt.Println("redir in")
	var url string
	redir := make(chan string, 1)
	accept := make(chan bool, 1)
	cb := func(url string) bool {
		redir <- url
		return <-accept
	}
	fmt.Println("redir pre")
	wf.NextRedirect(cb)
	fmt.Println("redir mid")
	select {
	case url = <-redir:
		accept <- true
	case <-time.After(redirectTimeout):
		accept <- false
	}
	fmt.Println("redir post")

	if url == "" {
		wr.Write([]byte("Done! You may close this page."))
		return
	}
	http.Redirect(wr, req, url, http.StatusFound)
}
