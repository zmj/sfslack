package workflow

import (
	"context"
	"fmt"
	"strings"

	"time"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type requestWorkflow struct {
	*wfBase
	downloadShare *sharefile.Share
}

func newRequest(host Host) Workflow {
	return &requestWorkflow{
		wfBase: newBase(host),
	}
}

func (wf *requestWorkflow) Setup() error {
	folder, err := wf.createWorkflowFolder()
	if err != nil {
		wf.err = err
		return err
	}
	// go subscribe
	requestShare, err := wf.sf.CreateRequestShare(context.TODO(), folder.ID)
	if err != nil {
		wf.err = err
		return err // cancel sub - check done / shutdown called?
	}
	// wait for subscribe
	uploadURL := requestShare.URI
	wf.Reply(wf.requestMessage(uploadURL))

	return nil
}

func (wf *requestWorkflow) Listen() error {
	if wf.err != nil {
		return wf.err
	}
	var notify <-chan time.Time
	done := time.After(30 * time.Minute)
	for {
		select {
		case <-wf.events:
			notify = time.After(2 * time.Second)
		case <-notify:
			notify = nil
			newFiles, err := wf.getNewFiles()
			if err != nil {
				wf.err = fmt.Errorf("Failed to get info for upload notification: %v", err)
				return wf.err
			}
			wf.downloadShare, err = addToShare(wf.sf, wf.downloadShare, newFiles)
			if err != nil {
				wf.err = fmt.Errorf("Failed to make share for notification: %v", err)
				return wf.err
			}
			accepted := wf.Host.Reply(wf.downloadMessage())
			if !accepted {
				return nil
			}
		case <-done:
			return nil
		}
	}
}

func (wf *requestWorkflow) requestMessage(uploadURL string) slack.Message {
	return slack.Message{
		ResponseType: slack.ResponseTypeInChannel,
		Text: fmt.Sprintf("%v has requested files: %v",
			wf.Host.User(),
			uploadURL),
	}
}

func (wf *requestWorkflow) downloadMessage() slack.Message {
	files := wf.files
	downloadURL := wf.sf.Account().DownloadAllURL(*wf.downloadShare)

	var msg slack.Message
	if len(files) == 1 {
		msg.Text = fmt.Sprintf("Received %v: %v", files[0].FileName, downloadURL)
	} else {
		msg.Text = fmt.Sprintf("Received %v files: %v", len(files), downloadURL)
		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, file.FileName)
		}
		msg.Attachments = []slack.Attachment{
			slack.Attachment{
				Text:     strings.Join(fileNames, "\n"),
				Fallback: strings.Join(fileNames, " "),
			},
		}
	}
	return msg
}
