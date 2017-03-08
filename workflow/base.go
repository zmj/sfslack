package workflow

import (
	"time"

	"github.com/zmj/sfslack/slack"
)

type wfBase struct {
	Args
	started time.Time
	done    chan struct{}
	err     error
}

func newBase(args Args) *wfBase {
	return &wfBase{
		Args:    args,
		started: time.Now(),
		done:    make(chan struct{}),
	}
}

func (wf *wfBase) reply(msg slack.Message) {
	go wf.Reply(Response{Msg: msg})
}

func (wf *wfBase) replyOrRedirect(msg slack.Message, url string) {
	go wf.Reply(Response{Msg: msg, URL: url})
}

func (wf *wfBase) fatal(err error) {
	wf.err = err
	go wf.Reply(errorResponse(err))
	close(wf.done)
}

func errorResponse(err error) Response {
	return Response{Msg: errorMessage(err)}
}

func errorMessage(err error) slack.Message {
	return slack.Message{Text: err.Error()}
}

func (wf *wfBase) Done() <-chan struct{} {
	return wf.done
}

func (wf *wfBase) Err() error {
	return wf.err
}

func (wf *wfBase) Shutdown() {
	panic(nil)
	// todo
}
