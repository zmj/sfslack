package server

import (
	"fmt"
	"net/http"

	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

var wfTypes = map[string]*workflow.Definition{
	"send":    workflow.Definitions.Send,
	"request": workflow.Definitions.Request,
}

func (srv *server) newCommand(wr http.ResponseWriter, req *http.Request) {
	cmd, err := srv.parseCommand(req)
	err = fmt.Errorf("Command parse failed: %v", err)
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

func (srv *server) parseCommand(req *http.Request) (slack.Command, error) {
	values, err := httpValues(req)
	if err != nil {
		return slack.Command{}, err
	}
	cmd, err := slack.ParseCommand(values)
	if err != nil {
		return slack.Command{}, err
	}

	if cmd.Command != slashCommand {
		return slack.Command{}, fmt.Errorf("Unexpected command %v", cmd.Command)
	}
	if cmd.Token != srv.config.SlackVerificationToken {
		return slack.Command{}, fmt.Errorf("Unexpected token %v", cmd.Token)
	}

	return cmd, nil
}
