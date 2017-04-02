package workflow

import (
	"context"
	"fmt"

	"github.com/zmj/sfslack/slack"
)

type sendWorkflow struct {
	*wfBase
}

func newSend(host Host) Workflow {
	return &sendWorkflow{
		wfBase: newBase(host),
	}
}

func (wf *sendWorkflow) Setup() error {
	if wf.err != nil {
		return fmt.Errorf("Workflow already error\n%v", wf.err)
	}
	folder, err := wf.createWorkflowFolder()
	if err != nil {
		wf.err = fmt.Errorf("Failed to create workflow folder\n%v", err)
		return wf.err
	}

	// go subscribe
	err = wf.subscribe(folder)
	if err != nil {
		wf.err = fmt.Errorf("Failed to subscribe to workflow folder\n%v", err)
		return wf.err
	}

	requestShare, err := wf.sf.CreateRequestShare(context.TODO(), folder.ID)
	if err != nil {
		wf.err = fmt.Errorf("Failed to create request share\n%v", err)
		return wf.err // cancel sub - check done / shutdown called?
	}
	// wait for subscribe
	uploadURL := requestShare.URI
	wf.RedirectOrReply(uploadURL, uploadMessage(uploadURL))

	return nil
}

func (wf *sendWorkflow) Event() {

}

func (wf *sendWorkflow) Listen() error {
	if wf.err != nil {
		return fmt.Errorf("Workflow already error\n%v", wf.err)
	}
	return nil
}

func uploadMessage(uploadURL string) slack.Message {
	return slack.Message{
		Text: fmt.Sprintf("Upload your files: %v", uploadURL),
	}
}
