package workflow

import (
	"time"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type wfBase struct {
	Host
	sf      *sharefile.Login
	started time.Time
	done    chan struct{}
	err     error
}

func newBase(host Host) *wfBase {
	return &wfBase{
		Host:    host,
		sf:      host.Authenticate(),
		started: time.Now(),
		done:    make(chan struct{}),
	}
}

func (wf *wfBase) fatal(err error) {
	wf.err = err
	go wf.Reply(errorMessage(err))
	close(wf.done)
}

func errorMessage(err error) slack.Message {
	return slack.Message{Text: err.Error()}
}

func (wf *wfBase) Done() <-chan struct{} {
	return wf.done
}

func (wf *wfBase) Err() error {
	return wf.err
}

func (wf *wfBase) Shutdown() {
	panic(nil)
	// todo
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
