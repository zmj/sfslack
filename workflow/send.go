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
	wf.RedirectOrReply(uploadURL, uploadMessage(uploadURL))

	// go event loop
	return nil
}

func uploadMessage(uploadURL string) slack.Message {
	return slack.Message{
		Text: fmt.Sprintf("Upload your files: %v", uploadURL),
	}
}

func (wf *sendWorkflow) Event() {

}
