package workflow

import (
	"strings"

	"errors"
	"fmt"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type Workflow interface {
	ID() int // can this be at main?
	Cmd() slack.Command
	Start(sf sharefile.Login, replyCbs ReplyCallbacks)
}

type ReplyCallbacks struct {
	Message  func(slack.Message) error
	Redirect func(string) error
}

func NewWorkflow(cmd slack.Command, id int) (Workflow, error) {
	// common construction?
	switch strings.ToLower(cmd.Text) {
	case "send":
		return newSend(cmd, id), nil
	case "request":
		return newRequest(cmd, id), nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown command '%v'", cmd.Text))
	}
}
