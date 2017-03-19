package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func (srv *server) sfAuth(wf *runner, wr http.ResponseWriter, req *http.Request) {
	wf.SetLogin(req.URL.Query())
	srv.redirect(wf, wr, req)
}

func (srv *server) slackAuth(wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))
}

const buttonPageHTML = `<html><body>
<a href="https://slack.com/oauth/authorize?&client_id=52278651460.156614739669&scope=commands,bot"><img alt="Add to Slack" height="40" width="139" src="https://platform.slack-edge.com/img/add_to_slack.png" srcset="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" /></a>
</body></html>`
