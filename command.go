package main

import (
	"errors"
	"net/http"
	"net/url"

	"fmt"
	"time"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

var (
	firstResponseTimeout = 2 * time.Second
	helpMessage          = &slack.Message{}
	workingMessage       = &slack.Message{}
)

func (srv *server) newCommand(wr http.ResponseWriter, req *http.Request) {
	var respondErr error
	defer logRespondError(respondErr)

	cmd, err := parseCommand(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}
	wr.Header().Add("Content-Type", "application/json")
	wf, err := srv.newWorkflow(cmd)
	if err != nil {
		_, respondErr = helpMessage.WriteTo(wr)
		return
	}
	userAuth, authFound := srv.authCache.TryGet(cmd.User)
	var firstResponse slack.Message
	if authFound {
		firstResponse = startAuthenticatedWorkflow(wf, userAuth)
	} else {
		go wf.Start(userAuth, nil)
		loginURL := srv.authCache.LoginURL(wf.ID())
		_, respondErr := loginMessage(loginURL).WriteTo(wr)
	}
	_, respondErr = firstResponse.WriteTo(wr)
}

func startAuthenticatedWorkflow(wf workflow.Workflow, login sharefile.Login) slack.Message {
	response := make(chan slack.Message)
	accepted := make(chan error, 1)
	cb := func(msg slack.Message) error {
		response <- msg
		return <-accepted
	}
	go wf.Start(login, cb)
	select {
	case msg := <-response:
		accepted <- nil
		return msg
	case <-time.After(firstResponseTimeout):
		accepted <- errors.New("Timed out")
		return *workingMessage
	}
}

func parseCommand(req *http.Request) (slack.Command, error) {
	var values url.Values
	if req.Method == "GET" {
		values = req.URL.Query()
	} else if req.Method == "POST" {
		err := req.ParseForm()
		if err != nil {
			return slack.Command{}, err
		}
		values = req.PostForm
	} else {
		return slack.Command{}, errors.New("Unsupported HTTP method " + req.Method)
	}
	return slack.ParseCommand(values)
}

func logRespondError(err error) {
	if err == nil {
		return
	}
	fmt.Printf("%v Response failure: %v", time.Now(), err.Error())
}

func loginMessage(loginURL string) slack.Message {
	return slack.Message{}
}
