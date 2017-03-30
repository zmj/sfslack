package server

import (
	"fmt"
	"net/http"

	"github.com/zmj/sfslack/server/wfhost"
	"github.com/zmj/sfslack/slack"
)

func (srv *server) newCommand(wr http.ResponseWriter, req *http.Request) {
	cmd, err := srv.parseCommand(req)
	if err != nil {
		err = fmt.Errorf("Command parse failed: %v", err)
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	_, msg := srv.wfSvc.New(cmd, callbackUrls{srv.publicHost(req)})
	wr.Header().Add("Content-Type", "application/json")
	_, err = msg.WriteTo(wr)
	if err != nil {
		srv.log.Err(err)
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (srv *server) newCommandClick(wf *wfhost.Runner, wr http.ResponseWriter, req *http.Request) {
	err := wf.SetDefinition(req.URL.Query())
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}
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
