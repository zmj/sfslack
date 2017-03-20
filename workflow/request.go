package workflow

import (
	"fmt"

	"github.com/zmj/sfslack/slack"
)

type requestWorkflow struct {
	*wfBase
}

func newRequest(host Host) Workflow {
	return &requestWorkflow{
		wfBase: newBase(host),
	}
}

func (wf *requestWorkflow) Setup() error {
	folder, err := wf.createWorkflowFolder()
	if err != nil {
		wf.fatal(err)
		return err
	}
	// go subscribe
	requestShare, err := wf.sf.CreateRequestShare(folder.ID)
	if err != nil {
		wf.fatal(err)
		return err // cancel sub - check done / shutdown called?
	}
	// wait for subscribe
	uploadURL := requestShare.URI
	wf.Reply(wf.requestMessage(uploadURL))

	return nil
}

func (wf *requestWorkflow) Listen() {

}

func (wf *requestWorkflow) requestMessage(uploadURL string) slack.Message {
	return slack.Message{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         fmt.Sprintf("%v has requested files", wf.Host.User()),
	}
}
