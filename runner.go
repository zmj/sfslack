package main

import (
	"github.com/zmj/sfslack/slack"
)

func (srv *server) startWorkflowForRedirect() string {
	return ""
}

func (srv *server) startWorkflowForResponse() slack.Message {
	return slack.Message{}
}

/*
func startWorkflowForResponse(wf workflow.Workflow, login sharefile.Login) slack.Message {
	response := make(chan slack.Message, 1)
	accepted := make(chan error, 1)
	cb := func(msg slack.Message) error {
		response <- msg
		return <-accepted
	}
	go wf.Start(login, workflow.ReplyCallbacks{Message: cb})
	select {
	case msg := <-response:
		accepted <- nil
		return msg
	case <-time.After(2 * time.Second):
		accepted <- errors.New("Timed out")
		return workingMessage()
	}
}
*/

/*
func startWorkflowForRedirect(wf workflow.Workflow, login sharefile.Login) string {
	redirect := make(chan string, 1)
	accepted := make(chan error, 1)
	cb := func(url string) error {
		redirect <- url
		return <-accepted
	}
	go wf.Start(login, workflow.ReplyCallbacks{Redirect: cb})
	select {
	case url := <-redirect:
		accepted <- nil
		return url
	case <-time.After(3 * time.Second):
		accepted <- errors.New("Timed out")
		return ""
	}
}
*/
