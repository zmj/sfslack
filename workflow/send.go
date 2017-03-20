package workflow

import (
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
		return wf.err
	}
	folder, err := wf.createWorkflowFolder()
	if err != nil {
		wf.err = err
		return err
	}

	// go subscribe
	err = wf.subscribe(folder)
	if err != nil {
		wf.err = err
		return err
	}

	requestShare, err := wf.sf.CreateRequestShare(folder.ID)
	if err != nil {
		wf.err = err
		return err // cancel sub - check done / shutdown called?
	}
	// wait for subscribe
	uploadURL := requestShare.URI
	wf.RedirectOrReply(uploadURL, uploadMessage(uploadURL))

	return nil
}

func (wf *sendWorkflow) Event() {

}

func (wf *sendWorkflow) Listen() error {
	if wf.Err != nil {
		return wf.err
	}
	return nil
}

func uploadMessage(uploadURL string) slack.Message {
	return slack.Message{
		Text: fmt.Sprintf("Upload your files: %v", uploadURL),
	}
}
