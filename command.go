package main

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"

	"fmt"
	"time"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

const (
	commandPath = "/sfslack/command"
)

var (
	firstResponseTimeout = 2 * time.Second
)

func (srv *server) newCommand(wr http.ResponseWriter, req *http.Request) {
	var respondErr error
	defer logRespondError(respondErr)

	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	cmd, err := parseCommand(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}
	wr.Header().Add("Content-Type", "application/json")
	wf, err := srv.newWorkflow(cmd)
	if err != nil {
		_, respondErr = helpMessage().WriteTo(wr)
		return
	}
	login, authFound := srv.authCache.TryGet(cmd.User)
	var response slack.Message
	if authFound {
		response = startAuthenticatedWorkflow(wf, login)
	} else {
		authCallbackURL := srv.authCallbackURL(req, wf.ID())
		loginURL := srv.authCache.LoginURL(authCallbackURL)
		response = loginMessage(loginURL)
	}
	_, respondErr = response.WriteTo(wr)
}

func startAuthenticatedWorkflow(wf workflow.Workflow, login sharefile.Login) slack.Message {
	response := make(chan slack.Message, 1)
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
		return workingMessage()
	}
}

func parseCommand(req *http.Request) (slack.Command, error) {
	values, err := httpValues(req)
	if err != nil {
		return slack.Command{}, err
	}
	return slack.ParseCommand(values)
}

func httpValues(req *http.Request) (url.Values, error) {
	if req.Method == "GET" {
		return req.URL.Query(), nil
	} else if req.Method == "POST" {
		err := req.ParseForm()
		if err != nil {
			return url.Values{}, err
		}
		return req.PostForm, nil
	} else {
		return url.Values{}, errors.New("Unsupported HTTP method " + req.Method)
	}
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

func helpMessage() slack.Message {
	return slack.Message{
		Text: "command args",
	}
}

func workingMessage() slack.Message {
	return slack.Message{
		Text: "Logging you in...",
	}
}
