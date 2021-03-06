package workflow

import (
	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type Workflow interface {
	Setup() error
	Listen() error
	Event(sharefile.WebhookSubscriptionEvent)
	Cleanup() error
}

type Host interface {
	Authenticate() *sharefile.Login
	Reply(slack.Message) bool
	RedirectOrReply(string, slack.Message) bool
	ReplyErr(error) bool
	Name() string
	User() string
	EventCallbackURL() string
}

type Constructor func(Host) Workflow

type Definition struct {
	Arg         string
	Description string
	Constructor Constructor
}

var Definitions = struct{ Send, Request *Definition }{
	Send:    &Definition{"send", "Share Files", newSend},
	Request: &Definition{"request", "Request Files", newRequest},
}
