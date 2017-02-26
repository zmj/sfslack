package workflow

import (
	"fmt"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type requestWorkflow struct {
	*wfBase
}

func newRequest(wf *wfBase, cmd slack.Command) Workflow {
	return &requestWorkflow{
		wfBase: wf,
	}
}

func (wf *requestWorkflow) Start(sf sharefile.Login, firstReply ReplyCallbacks) {
	fmt.Println("Request start!")
}
