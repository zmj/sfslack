package workflow

import (
	"time"

	"github.com/zmj/sfslack/sharefile"
)

const (
	slackFolderName = ".slack"
)

func (wf *wfBase) createWorkflowFolder(sf sharefile.Login) (sharefile.Folder, error) {
	slackFolder, err := getOrCreateSlackFolder(sf)
	if err != nil {
		return sharefile.Folder{}, err
	}
	return sf.CreateFolder(wf.folderName(), slackFolder.ID)
}

func (wf *wfBase) folderName() string {
	time := time.Now().Format("2006-01-02 15:04:05")
	return wf.cmd.Channel.Name + " " + time
}

func getOrCreateSlackFolder(sf sharefile.Login) (sharefile.Folder, error) {
	children, err := sf.GetChildren("home")
	if err != nil {
		return sharefile.Folder{}, err
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
	return sf.CreateFolder(slackFolderName, "home")
}
