package wfutils

import "github.com/zmj/sfslack/slack"
import "github.com/zmj/sfslack/workflow"
import "time"

type Runner interface {
	StartAndRedirect() string
	StartAndRespond() slack.Message
	Done() <-chan struct{}
}

func NewRunner(builder *Builder, cache *Cache) Runner {
	return &runner{
		builder:   builder,
		cache:     cache,
		responses: make(chan workflow.Response),
		done:      make(chan struct{}),
	}
}

func (r *runner) Done() <-chan struct{} {
	return r.done
}

func (r *runner) StartAndRedirect() string {
	redirect := make(chan string, 1)
	r.redirect = redirect
	accepted := make(chan bool, 1)
	r.accepted = accepted
	go r.run()
	select {
	case url := <-redirect:
		accepted <- true
		return url
	case <-time.After(3 * time.Second):
		accepted <- false
		return ""
	}
}

func (r *runner) StartAndRespond() slack.Message {
	response := make(chan slack.Message, 1)
	r.firstResponse = response
	accepted := make(chan bool, 1)
	r.accepted = accepted
	go r.run()
	select {
	case msg := <-response:
		accepted <- true
		return msg
	case <-time.After(slack.InitialReplyTimeout):
		accepted <- false
		return timeoutMessage()
	}
}

type runner struct {
	builder   *Builder
	cache     *Cache
	responses chan workflow.Response
	done      chan struct{}
	// authcache for logout?
	redirect      chan<- string
	firstResponse chan<- slack.Message
	accepted      <-chan bool
	// done callback?
}

func (r *runner) run() {
	r.builder.Reply = r.receive
	wf := r.builder.Definition.Constructor(*r.builder.Args)
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

func timeoutMessage() slack.Message {
	return slack.Message{Text: "Logging you in..."}
}
