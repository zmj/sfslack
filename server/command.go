package server

import (
	"net/http"

	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

var wfTypes = map[string]*workflow.Definition{
	"send":    workflow.Definitions.Send,
	"request": workflow.Definitions.Request,
}

func (srv *server) newCommand(wr http.ResponseWriter, req *http.Request) {
	cmd, err := parseCommand(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	_, msg := srv.new(cmd, publicHost(req))
	wr.Header().Add("Content-Type", "application/json")
	_, err = msg.WriteTo(wr)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (srv *server) newCommandClick(wf *runner, wr http.ResponseWriter, req *http.Request) {
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
