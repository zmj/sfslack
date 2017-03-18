package server

import (
	"sync"
	"time"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

type runner struct {
	srv        *server
	mu         *sync.Mutex
	wfID       int
	cmd        slack.Command
	def        *workflow.Definition
	defWait    chan *workflow.Definition
	login      *sharefile.Login
	loginWait  chan *sharefile.Login
	firstReply chan slack.Message
	next       *redirect
}

func (srv *server) get(wfID int) (*runner, bool) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	r, ok := srv.workflows[wfID]
	return r, ok
}

func (srv *server) put(r *runner) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.wfID++
	r.wfID = srv.wfID
	srv.workflows[srv.wfID] = r
}

func (srv *server) new(cmd slack.Command) (*runner, slack.Message) {
	r := &runner{
		cmd: cmd,
	}
	srv.put(r)
	go r.run()
	return r, slack.Message{}
}

func (r *runner) run() {
	def, ok := wfTypes[r.cmd.Text]
	if !ok {
		r.defWait = make(chan *workflow.Definition, 1)
		// Reply(helpmsg)
		def = <-r.defWait
	}

	login, ok := r.srv.authCache.TryGet(r.cmd.User)
	if !ok {
		r.loginWait = make(chan *sharefile.Login, 1)
		// RedirectOrReply(authmsg)
		login = <-r.loginWait
	}

}

func (r *runner) Reply(msg slack.Message) {
	if r.firstReply != nil {
		r.firstReply <- msg
		return
	}
	err := msg.RespondTo(r.cmd)
	// clear redirect?
}

func (r *runner) Redirect(url string) {

}

func (r *runner) RedirectOrReply(msg slack.Message, url string) {

}

func (r *runner) NextRedirect() *redirect {
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

func (r *runner) Name() string {
	time := time.Now().Format("2006-01-02 15:04:05")
	return r.cmd.Channel.Name + " " + time
}
