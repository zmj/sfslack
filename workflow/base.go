package workflow

import (
	"time"

	"github.com/zmj/sfslack/slack"
)

type wfBase struct {
	id      int
	cmd     slack.Command
	started time.Time
}

func newBase(id int, cmd slack.Command) *wfBase {
	return &wfBase{
		id:      id,
		cmd:     cmd,
		started: time.Now(),
	}
}

func (wf *wfBase) ID() int {
	return wf.id
}

func (wf *wfBase) Cmd() slack.Command {
	return wf.cmd
}
