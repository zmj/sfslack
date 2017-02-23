package workflow

import (
	"time"
)

type wfBase struct {
	id      int
	started time.Time
}

func newBase(id int) *wfBase {
	return &wfBase{
		id:      id,
		started: time.Now(),
	}
}

func (wf *wfBase) ID() int {
	return wf.id
}
