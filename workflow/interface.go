package workflow

import (
	"strings"

	"fmt"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type Workflow interface {
	Cmd() slack.Command
	Start(sf sharefile.Login, replyCbs ReplyCallbacks)
}

type ReplyCallbacks struct {
	Message  func(slack.Message) error
	Redirect func(string) error
}

func NewWorkflow(cmd slack.Command, eventURL string) (Workflow, error) {
	var constructor func(*wfBase, slack.Command) Workflow
	switch strings.ToLower(cmd.Text) {
	case "send":
		constructor = newSend
	case "request":
		constructor = newRequest
	default:
		return nil, fmt.Errorf("Unknown command '%v'", cmd.Text)
	}
	wf := newBase(cmd, eventURL)
	return constructor(wf, cmd), nil
}
