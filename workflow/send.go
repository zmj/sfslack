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

// setup? init?
func (wf *sendWorkflow) Start(sf sharefile.Login,
	firstResponse ResponseCallback,
	firstRedirect RedirectCallback) {
	fmt.Println("Send start!")
	// create slack folder if necessary
	// create workflow folder
	// go create request share
	// go subscribe to workflow folder
	// respond or redirect or msg to request share

}

func (wf *sendWorkflow) Event() {

}
