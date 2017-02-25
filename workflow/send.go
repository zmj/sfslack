package workflow

import (
	"fmt"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type sendWorkflow struct {
	*wfBase
}

func newSend(cmd slack.Command, id int) *sendWorkflow {
	return &sendWorkflow{
		wfBase: newBase(id, cmd),
	}
}

func (wf *sendWorkflow) Start(sf sharefile.Login,
	firstResponse ResponseCallback,
	firstRedirect RedirectCallback) {
	fmt.Println("Send start!")
}
