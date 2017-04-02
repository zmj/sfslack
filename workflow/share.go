package workflow

import (
	"context"

	"fmt"

	"github.com/zmj/sfslack/sharefile"
)

const (
	slackFolderName = ".slack"
)

func (wf *wfBase) createWorkflowFolder() (sharefile.Folder, error) {
	slackFolder, err := getOrCreateSlackFolder(wf.sf)
	if err != nil {
		return sharefile.Folder{}, fmt.Errorf("Failed to get slack folder: %v", err)
	}
	return wf.sf.CreateFolder(context.TODO(), wf.Name(), slackFolder.ID)
}

func getOrCreateSlackFolder(sf *sharefile.Login) (sharefile.Folder, error) {
	children, err := sf.GetChildren(context.TODO(), "home")
	if err != nil {
		return sharefile.Folder{}, fmt.Errorf("Failed to get home folder children: %v", err)
	}
	for _, item := range children {
		folder, err := item.Folder()
		if err != nil {
			continue
		}
		if folder.FileName == slackFolderName {
			return folder, nil
		}
	}
	return sf.CreateFolder(context.TODO(), slackFolderName, "home")
}

func (wf *wfBase) subscribe(folder sharefile.Folder) error {
	// save sub on base for cleanup
	_, err := wf.sf.Subscribe(context.TODO(),
		folder,
		wf.Host.EventCallbackURL(),
		sharefile.OperationNameUpload)
	return err
}
