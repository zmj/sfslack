package workflow

import (
	"strings"

	"errors"
	"fmt"

	"github.com/zmj/sfslack/slack"
)

type Workflow interface {
	ID() int
}

func NewWorkflow(cmd slack.Command) (Workflow, error) {
	// common construction?
	switch strings.ToLower(cmd.Text) {
	case "send":
		return newSend(cmd), nil
	case "request":
		return newRequest(cmd), nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown command '%v'", cmd.Text))
	}
}
