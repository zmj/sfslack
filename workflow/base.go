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
	events  chan sharefile.WebhookSubscriptionEvent
	folder  sharefile.Folder
	files   []sharefile.File
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
	return nil
}

func (wf *wfBase) Event(event sharefile.WebhookSubscriptionEvent) {
	if wf.err != nil {
		return // err->unsub?
	}
	go func() {
		wf.events <- event
	}()
}
