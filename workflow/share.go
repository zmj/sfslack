package workflow

import (
	"github.com/zmj/sfslack/sharefile"
)

const (
	slackFolderName = ".slack"
)

func (wf *wfBase) createWorkflowFolder(sf sharefile.Login) (sharefile.Folder, error) {

}

func getOrCreateSlackFolder(sf sharefile.Login) (sharefile.Folder, error) {
	// get home children
	// if !exists slack
	// create home/slack
}
