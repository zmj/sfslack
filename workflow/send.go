package workflow

import (
	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type sendWorkflow struct {
	*wfBase
}

func newSend(cmd slack.Command, id int) *sendWorkflow {
	return &sendWorkflow{
		wfBase: newBase(id),
	}
}

func (wf *sendWorkflow) Start(sf sharefile.Login, firstResponse ResponseCallback) {

}
