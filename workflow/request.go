package workflow

import (
	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type requestWorkflow struct {
	id int
}

func newRequest(cmd slack.Command, id int) *requestWorkflow {
	return &requestWorkflow{
		id: id,
	}
}

func (wf *requestWorkflow) ID() int {
	return wf.id
}

func (wf *requestWorkflow) Start(sf sharefile.Login, firstResponse ResponseCallback) {

}
