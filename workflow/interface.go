package workflow

import (
	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type Workflow interface {
	Setup() error
	Listen()
}

type Host interface {
	Authenticate() *sharefile.Login
	Reply(slack.Message)
	RedirectOrReply(string, slack.Message)
	Name() string
	User() string
}

type Constructor func(Host) Workflow

type Definition struct {
	Description string
	Constructor Constructor
}

var Definitions = struct{ Send, Request *Definition }{
	Send:    &Definition{"Share Files", newSend},
	Request: &Definition{"Request Files", newRequest},
}
