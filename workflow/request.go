package workflow

import (
	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type requestWorkflow struct {
	*wfBase
}

func newRequest(cmd slack.Command, id int) *requestWorkflow {
	return &requestWorkflow{
		wfBase: newBase(id),
	}
}

func (wf *requestWorkflow) Start(sf sharefile.Login, firstResponse ResponseCallback) {

}
