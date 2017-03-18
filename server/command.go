package server

import (
	"net/http"
	"net/http/httputil"

	"fmt"

	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

var wfTypes = map[string]*workflow.Definition{
	"send":    workflow.Definitions.Send,
	"request": workflow.Definitions.Request,
}

func (srv *server) newCommand(wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	cmd, err := parseCommand(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	wr.Header().Add("Content-Type", "application/json")

	wf := srv.workflows.new(cmd)

	def, ok := wfTypes[cmd.Text]
	if !ok {
		url := commandClickURL(publicHost(req), wf.wfID)
		_, err = helpMessage(url).WriteTo(wr)
		// log err
		return
	}
	wf.SetDefinition(def)

	panic("wf auth msg")
}

func (srv *server) newCommandClick(wf *runner, wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	wfType := req.URL.Query().Get(wfTypeQueryKey)
	def, ok := wfTypes[wfType]
	if !ok {
		http.Error(wr, "Unknown workflow type", http.StatusBadRequest)
		return
	}
	wf.SetDefinition(def)
	srv.redirect(wf, wr, req)
}

func parseCommand(req *http.Request) (slack.Command, error) {
	values, err := httpValues(req)
	if err != nil {
		return slack.Command{}, err
	}
	return slack.ParseCommand(values)
}
