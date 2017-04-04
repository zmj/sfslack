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
	folder, err := wf.sf.CreateFolder(context.TODO(), wf.Name(), slackFolder.ID)
	if err != nil {
		return sharefile.Folder{}, fmt.Errorf("Failed to create workflow folder: %v", err)
	}
	wf.folder = folder
	return folder, nil
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

func (wf *wfBase) getNewFiles() ([]sharefile.File, error) {
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

func addToShare(sf *sharefile.Login, share *sharefile.Share, newFiles []sharefile.File) (*sharefile.Share, error) {
	var result sharefile.Share
	var err error
	if share == nil {
		result, err = sf.CreateSendShare(context.TODO(), newFiles)
		if err != nil {
			err = fmt.Errorf("Failed to create share: %v", err)
		}
	} else {
		result, err = sf.UpdateSendShare(context.TODO(), *share, newFiles)
		if err != nil {
			err = fmt.Errorf("Failed to update share: %v", err)
		}
	}
	return &result, err
}
