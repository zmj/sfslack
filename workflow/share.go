package workflow

import "github.com/zmj/sfslack/sharefile"

const (
	slackFolderName = ".slack"
)

func (wf *wfBase) createWorkflowFolder() (sharefile.Folder, error) {
	slackFolder, err := getOrCreateSlackFolder(wf.sf)
	if err != nil {
		return sharefile.Folder{}, err
	}
	return wf.sf.CreateFolder(wf.Name(), slackFolder.ID)
}

func getOrCreateSlackFolder(sf *sharefile.Login) (sharefile.Folder, error) {
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

func (wf *wfBase) subscribe(folder sharefile.Folder) error {
	toCreate := sharefile.WebhookSubscription{
		SubscriptionContext: sharefile.SubscriptionContext{
			ResourceType: sharefile.ResourceTypeFolder,
			ResourceId:   folder.ID,
		},
		Events: []sharefile.SubscribedResourceEvent{
			sharefile.SubscribedResourceEvent{
				ResourceType:  sharefile.ResourceTypeFile,
				OperationName: sharefile.OperationNameUpload,
			},
		},
		WebhookURL: wf.Host.EventCallbackURL(),
	}
	// save sub on base for cleanup
	_, err := wf.sf.CreateSubscription(toCreate)
	return err
}
