package workflow

import (
	"fmt"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type requestWorkflow struct {
	*wfBase
}

func newRequest(cmd slack.Command, id int) *requestWorkflow {
	return &requestWorkflow{
		wfBase: newBase(id, cmd),
	}
}

func (wf *requestWorkflow) Start(sf sharefile.Login, firstReply ReplyCallbacks) {
	fmt.Println("Request start!")
}
