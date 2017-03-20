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

func (wf *wfBase) Cleanup() error {
	return nil
}
