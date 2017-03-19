package workflow

import (
	"fmt"

	"time"

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
	<-time.After(10 * time.Second)
	uploadURL := requestShare.URI
	wf.RedirectOrReply(uploadURL, uploadMessage(uploadURL))

	// go event loop
	return nil
}

func (wf *sendWorkflow) Event() {

}

func (wf *sendWorkflow) Listen() {

}

func uploadMessage(uploadURL string) slack.Message {
	return slack.Message{
		Text: fmt.Sprintf("Upload your files: %v", uploadURL),
	}
}
