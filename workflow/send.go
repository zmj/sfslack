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
	folder        sharefile.Folder
	files         []sharefile.File
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
	wf.folder = folder

	// go subscribe
	err = wf.subscribe(folder)
	if err != nil {
		wf.err = fmt.Errorf("Failed to subscribe to workflow folder: %v", err)
		return wf.err
	}

	requestShare, err := wf.sf.CreateRequestShare(context.TODO(), folder.ID)
	if err != nil {
		wf.err = fmt.Errorf("Failed to create request share: %v", err)
		return wf.err // cancel sub - check done / shutdown called?
	}
	// wait for subscribe
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
			err = wf.addToShare(newFiles)
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

func (wf *sendWorkflow) addToShare(newFiles []sharefile.File) error {
	if wf.downloadShare == nil {
		share, err := wf.sf.CreateSendShare(context.TODO(), newFiles)
		if err != nil {
			wf.err = fmt.Errorf("Failed to create share: %v", err)
			return wf.err
		}
		wf.downloadShare = &share
	} else {
		share, err := wf.sf.UpdateSendShare(context.TODO(), *wf.downloadShare, newFiles)
		if err != nil {
			wf.err = fmt.Errorf("Failed to update share: %v", err)
			return wf.err
		}
		wf.downloadShare = &share
	}
	return nil
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

func (wf *sendWorkflow) getNewFiles() ([]sharefile.File, error) {
	children, err := wf.sf.GetChildren(context.TODO(), wf.folder.ID)
	if err != nil {
		wf.err = fmt.Errorf("New file check failed to get folder contents: %v", err)
		return nil, wf.err
	}
	var newChildren []sharefile.File
	for _, child := range children {
		fi, err := child.File()
		if err != nil {
			continue
		}
		known := false
		for _, existing := range wf.files {
			if existing.ID == child.ID {
				known = true
				break
			}
		}
		if known {
			continue
		}
		newChildren = append(newChildren, fi)
		wf.files = append(wf.files, fi)
	}
	return newChildren, nil
}
