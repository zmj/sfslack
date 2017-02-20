package main

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
	"time"
	"fmt"
)

var (
    firstResponseTimeout = 2 * time.Second
    helpMessage = &slack.Message{}
    workingMessage = &slack.Message{}
)

func (srv *server) newCommand(wr http.ResponseWriter, req *http.Request) {
    var respondErr error
    defer logRespondError(respondErr)

	cmd, err := parseCommand()
	if err != nil {
		http.Error(wr, err, http.StatusBadRequest)
		return
	}
	wr.Header().Add("Content-Type", "application/json")
    wfID := srv.nextWorkflowID()
	wf, err := workflow.NewWorkflow(cmd, wfID)
	if err != nil {
		_, respondErr = helpMessage.WriteTo(wr)
		return
	}
	userAuth, authFound := srv.authCache.TryGet(cmd.User)
	if authFound {
        firstResponse := make(chan slack.Message, 1)
        cb := func(msg slack.Message) error {
            firstResponse <- msg
        }
		go wf.Start(userAuth, cb)
        select {
            case msg <- cb:
                _, respondErr = msg.WriteTo(wr)
            case <- time.After(firstResponseTimeout):
                _, respondErr = workingMessage.WriteTo(wr)
        }
	}
    else {        
        go wf.Start(userAuth, nil)
        loginURL := srv.authCache.LoginURL(wfID)
        _, respondErr := loginMessage(loginURL).WriteTo(wr)
    }
}

func parseCommand(req *http.Request) (slack.Command, error) {
	var values url.Values
	if req.Method == "GET" {
		values = req.URL.Query()
	} else if req.Method == "POST" {
		err = req.ParseForm()
		if err != nil {
			return slack.Command{}, err
		}
		values = req.PostForm
	} else {
		return slack.Command{}, errors.New("Unsupported HTTP method " + req.Method)
	}
	return slack.ParseCommand(values)
}

func logRespondError(cmd slack.Command, err error) {
    if err == nil {
        return
    }
    fmt.Printf("Response failure to %v: %v", cmd.User.Name, err.Error())
}

func loginMessage(loginURL string) slack.Message {

}