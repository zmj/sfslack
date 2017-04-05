package workflow

import (
	"context"
	"time"

	"fmt"

	"github.com/zmj/sfslack/sharefile"
)

type wfBase struct {
	Host
	sf           *sharefile.Login
	started      time.Time
	err          error
	events       chan sharefile.WebhookSubscriptionEvent
	folder       sharefile.Folder
	files        []sharefile.File
	subscription *sharefile.WebhookSubscription
}

func newBase(host Host) *wfBase {
	return &wfBase{
		Host:    host,
		sf:      host.Authenticate(),
		started: time.Now(),
		events:  make(chan sharefile.WebhookSubscriptionEvent),
	}
}

func (wf *wfBase) Err() error {
	return wf.err
}

func (wf *wfBase) Cleanup() error {
	var err error
	if wf.subscription != nil && wf.sf != nil {
		e := wf.sf.DeleteSubscription(context.TODO(), wf.subscription.ID)
		if e != nil {
			err = fmt.Errorf("Failed to unsubscribe: %v", e)
		}
	}

	return err
}

func (wf *wfBase) Event(event sharefile.WebhookSubscriptionEvent) {
	if wf.err != nil {
		return // err->unsub?
	}
	go func() {
		wf.events <- event
	}()
}
