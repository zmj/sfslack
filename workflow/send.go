package workflow

import (
	"context"
	"fmt"
	"strings"

	"time"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type sendWorkflow struct {
	*wfBase
	downloadShare *sharefile.Share
}

func newSend(host Host) Workflow {
	return &sendWorkflow{
		wfBase: newBase(host),
	}
}

func (wf *sendWorkflow) Setup() error {
	if wf.err != nil {
		return fmt.Errorf("Workflow already error: %v", wf.err)
	}
	folder, err := wf.createWorkflowFolder()
	if err != nil {
		wf.err = fmt.Errorf("Failed to create workflow folder: %v", err)
		return wf.err
	}

	c := make(chan struct{})
	go func() {
		err = wf.subscribe(folder)
		if err != nil {
			wf.err = fmt.Errorf("Failed to subscribe to workflow folder: %v", err)
		}
		c <- struct{}{}
	}()

	requestShare, err := wf.sf.CreateRequestShare(context.TODO(), folder.ID)
	if err != nil {
		wf.err = fmt.Errorf("Failed to create request share: %v", err)
		return wf.err // cancel sub - check done / shutdown called?
	}

	<-c
	if wf.err != nil {
		return wf.err
	}

	uploadURL := requestShare.URI
	wf.RedirectOrReply(uploadURL, uploadMessage(uploadURL))

	return nil
}

func (wf *sendWorkflow) Listen() error {
	if wf.err != nil {
		return fmt.Errorf("Workflow already error: %v", wf.err)
	}

	var notify <-chan time.Time
	done := time.After(10 * time.Minute)
	for {
		select {
		case <-wf.events:
			notify = time.After(1 * time.Second)
		case <-notify:
			notify = nil
			done = time.After(5 * time.Minute)
			newFiles, err := wf.getNewFiles()
			if err != nil {
				wf.err = fmt.Errorf("Failed to get info for upload notification: %v", err)
				return wf.err
			}
			if len(newFiles) == 0 {
				continue
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

func uploadMessage(uploadURL string) slack.Message {
	return slack.Message{
		Text: fmt.Sprintf("Upload your files: %v", uploadURL),
	}
}

func (wf *sendWorkflow) downloadMessage() slack.Message {
	user := wf.Host.User()
	files := wf.files // all files - new only instead?
	downloadURL := wf.sf.Account().DownloadAllURL(*wf.downloadShare)

	msg := slack.Message{ResponseType: slack.ResponseTypeInChannel}
	if len(files) == 1 {
		msg.Text = fmt.Sprintf("%v has shared %v: %v", user, files[0].FileName, downloadURL)
	} else {
		msg.Text = fmt.Sprintf("%v has shared %v files: %v", user, len(files), downloadURL)
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
