package workflow

import (
	"fmt"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type sendWorkflow struct {
	*wfBase
}

func newSend(wf *wfBase, cmd slack.Command) Workflow {
	return &sendWorkflow{
		wfBase: wf,
	}
}

// setup? init?
func (wf *sendWorkflow) Start(sf sharefile.Login, firstReply ReplyCallbacks) {
	var respondErr error
	defer logRespondError(respondErr)

	folder, err := wf.createWorkflowFolder(sf)
	if err != nil {
		respondErr = wf.firstReply(firstReply, errorMessage(err), "")
		return
	}
	// go subscribe
	requestShare, err := sf.CreateRequestShare(folder.ID)
	if err != nil {
		respondErr = wf.firstReply(firstReply, errorMessage(err), "")
		return // cancel sub - context close?
	}
	// wait for subscribe
	uploadURL := requestShare.URI
	respondErr = wf.firstReply(firstReply, wf.uploadMessage(uploadURL), uploadURL)
	if respondErr != nil {
		return
	}

}

func (wf *sendWorkflow) uploadMessage(uploadURL string) slack.Message {
	return slack.Message{
		Text: fmt.Sprintf("Upload your files: %v", uploadURL),
	}
}

func (wf *sendWorkflow) Event() {

}
