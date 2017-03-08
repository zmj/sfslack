package wfutils

import "github.com/zmj/sfslack/slack"
import "github.com/zmj/sfslack/workflow"

func (srv *server) startWorkflowForRedirect(builder *workflowBuilder) string {
	return ""
}

func (srv *server) startWorkflowForResponse(builder *workflowBuilder) slack.Message {
	return slack.Message{}
}

type runner struct {
	builder   *Builder
	cache     *Cache
	responses chan workflow.Response
	// authcache?
	redirect      chan<- string
	firstResponse chan<- slack.Message
	accepted      <-chan bool
}

func (r *runner) Run() {
	r.builder.Reply = r.receive
	wf := r.builder.definition.Constructor(*r.builder.Args)
	err := wf.Setup()
	if err != nil {
		// ??
		return
	}
	for {
		select {
		case resp := <-r.responses:
			err = r.reply(resp)
		case <-wf.Done():

		}
	}
}

func (r *runner) reply(resp workflow.Response) error {
	if r.redirect != nil {
		r.redirect <- resp.URL
		r.redirect = nil
	}
	if r.firstResponse != nil {
		r.firstResponse <- resp.Msg
		r.firstResponse = nil
	}
	if r.accepted != nil {
		accepted := <-r.accepted
		r.accepted = nil
		if accepted {
			return nil
		}
	}
	return resp.Msg.RespondTo(r.builder.Cmd)
}

func (r *runner) receive(resp workflow.Response) {
	go func() {
		r.responses <- resp
	}()
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
