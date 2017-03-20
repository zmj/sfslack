package workflow

import (
	"time"

	"github.com/zmj/sfslack/sharefile"
)

type wfBase struct {
	Host
	sf      *sharefile.Login
	started time.Time
	err     error
}

func newBase(host Host) *wfBase {
	return &wfBase{
		Host:    host,
		sf:      host.Authenticate(),
		started: time.Now(),
	}
}

func (wf *wfBase) Err() error {
	return wf.err
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
