package workflow

import (
	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type Workflow interface {
	Setup() error
	Done() <-chan struct{}
	Err() error
	Shutdown()
}

type Args struct {
	Cmd      slack.Command
	Sf       sharefile.Login
	Reply    func(Response)
	EventURL string
}

type Response struct {
	Msg slack.Message
	URL string
}

type Constructor func(Args) Workflow

type Definition struct {
	Description string
	Constructor Constructor
}

var Definitions = struct{ Send, Request Definition }{
	Send:    Definition{"Share Files", newSend},
	Request: Definition{"Request Files", newRequest},
}
