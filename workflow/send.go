package workflow

import (
	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type sendWorkflow struct {
	id int
}

func newSend(cmd slack.Command, id int) *sendWorkflow {
	return &sendWorkflow{
		id: id,
	}
}

func (wf *sendWorkflow) ID() int {
	return wf.id
}

func (wf *sendWorkflow) Start(sf sharefile.Login, firstResponse ResponseCallback) {

}
