package server

import (
	"sync"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

type runner struct {
	mu       *sync.Mutex
	wfID     int
	cmd      slack.Command
	def      workflow.Definition
	login    sharefile.Login
	redirect *redirect
}

func newRunner(cmd slack.Command, id int) *runner {
	return &runner{
		mu:   &sync.Mutex{},
		wfID: id,
		cmd:  cmd,
		redirect: &redirect{
			done: make(chan struct{}),
		},
	}
}
