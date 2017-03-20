package server

import (
	"fmt"
	"net/http"
)

func (srv *server) sfAuth(wf *runner, wr http.ResponseWriter, req *http.Request) {
	wf.SetLogin(req.URL.Query())
	srv.redirect(wf, wr, req)
}

func (srv *server) slackAuth(wr http.ResponseWriter, req *http.Request) {
	fmt.Printf("slack auth code %v\n", req.URL.Query().Get("code"))

	// swap for slack team + bot oauth tokens
	// save forever?
}

const buttonPageHTML = `<html><body>
<a href="https://slack.com/oauth/authorize?&client_id=52278651460.156614739669&scope=commands,bot"><img alt="Add to Slack" height="40" width="139" src="https://platform.slack-edge.com/img/add_to_slack.png" srcset="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" /></a>
</body></html>`

func (srv *server) install(wr http.ResponseWriter, req *http.Request) {
	wr.Write([]byte(buttonPageHTML))
}
