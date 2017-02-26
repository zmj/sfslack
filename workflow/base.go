package workflow

import (
	"fmt"
	"time"

	"github.com/zmj/sfslack/slack"
)

type wfBase struct {
	id                 int
	cmd                slack.Command
	started            time.Time
	delayedRepliesSent int
}

func newBase(id int, cmd slack.Command) *wfBase {
	return &wfBase{
		id:      id,
		cmd:     cmd,
		started: time.Now(),
	}
}

func (wf *wfBase) ID() int {
	return wf.id
}

func (wf *wfBase) Cmd() slack.Command {
	return wf.cmd
}

func errorMessage(err error) slack.Message {
	return slack.Message{Text: err.Error()}
}

func (wf *wfBase) respond(msg slack.Message) error {
	// todo
	return nil
}

func logRespondError(err error) {
	if err == nil {
		return
	}
	fmt.Printf("%v Response failure: %v", time.Now(), err.Error())
}

func (wf *wfBase) firstReply(rcb ReplyCallbacks, msg slack.Message, url string) error {
	var cb func() error
	if rcb.Message != nil {
		cb = func() error {
			return rcb.Message(msg)
		}
	} else if rcb.Redirect != nil {
		cb = func() error {
			return rcb.Redirect(url)
		}
	}

	err := cb()
	if err != nil {
		err = wf.respond(msg)
	}
	return err
}
