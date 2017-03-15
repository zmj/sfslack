package server

import (
	"sync"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

type runner struct {
	mu    *sync.Mutex
	wfID  int
	cmd   slack.Command
	def   *workflow.Definition
	login *sharefile.Login
	next  *redirect
}

func newRunner(cmd slack.Command, id int) *runner {
	return &runner{
		mu:   &sync.Mutex{},
		wfID: id,
		cmd:  cmd,
		next: &redirect{
			done: make(chan struct{}),
		},
	}
}

func (r *runner) Redirect() *redirect {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.next
}

func (r *runner) SetDefinition(def *workflow.Definition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.def != nil {
		return
	}
	r.def = def
	// construct workflow
	// launch workflow
}

func (r *runner) SetLogin(login *sharefile.Login) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.login != nil {
		return
	}
	r.login = login
	// signal login set?
	// pass on channel instead of set?
}

func (r *runner) setNext(url string, err error) {
	// lock? always from runner task?
	r.next.url = url
	r.next.err = err
	close(r.next.done)
	r.next = &redirect{
		done: make(chan struct{}),
	}
}
