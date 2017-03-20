package server

import (
	"sync"

	"net/url"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

type runner struct {
	*replier
	srv  *server
	wfID int
	wf   workflow.Workflow
	urls callbackURLs

	def     *workflow.Definition
	defWait chan struct{}

	login       *sharefile.Login
	loginValues url.Values
	loginWait   chan struct{}
}

func (srv *server) new(cmd slack.Command, host string) (*runner, slack.Message) {
	first := make(chan slack.Message, 1)
	r := &runner{
		replier: &replier{
			mu:         &sync.Mutex{},
			firstReply: first,
			cmd:        cmd,
			replies:    make(chan reply),
			done:       make(chan struct{}),
		},
		srv: srv,
	}
	srv.put(r)
	r.urls = srv.callbackURLs(host, r.wfID)
	go r.run()
	go r.sendReplies()
	return r, <-first
}

func (r *runner) run() {
	defer func() {
		if r.wf != nil {
			err := r.wf.Cleanup()
			r.srv.logErr(err)
		}
		close(r.done)
	}()
	// need shutdown on these two waits
	// what's the case for external shtudown?
	r.def = r.getDefinition()

	login, err := r.getLogin()
	if err != nil {
		r.Reply(errorMessage(err))
		r.srv.logErr(err)
		return
	}
	r.login = login

	r.wf = r.def.Constructor(r)
	err = r.wf.Setup()
	if err != nil {
		r.Reply(errorMessage(err))
		r.srv.logErr(err)
		return
	}

	err = r.wf.Listen()
	if err != nil {
		r.ReplyErr(err)
		r.srv.logErr(err)
		return
	}
}

func (r *runner) getDefinition() *workflow.Definition {
	def, ok := wfTypes[r.cmd.Text]
	if !ok {
		r.mu.Lock()
		r.defWait = make(chan struct{})
		r.mu.Unlock()

		r.Reply(helpMessage(r.urls.CommandClick))
		<-r.defWait
		def = r.def
	}
	return def
}

func (r *runner) getLogin() (*sharefile.Login, error) {
	login, ok := r.srv.authCache.TryGet(r.cmd.User)
	if !ok {
		r.mu.Lock()
		r.loginWait = make(chan struct{})
		r.mu.Unlock()

		loginURL := r.srv.authCache.LoginURL(r.urls.AuthCallback)
		r.RedirectOrReply(loginURL, loginMessage(loginURL))
		<-r.loginWait
		return r.srv.authCache.Add(r.cmd.User, r.loginValues)
	}
	return login, nil
}

func (r *runner) SetDefinition(def *workflow.Definition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.def != nil {
		return
	}
	r.def = def
	r.useCurrent = false
	close(r.defWait)
}

func (r *runner) SetLogin(cbValues url.Values) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.loginValues != nil {
		return
	}
	r.loginValues = cbValues
	r.useCurrent = false
	close(r.loginWait)
}

func (r *runner) Name() string {
	time := r.cmd.Received.Format("2006-01-02 15:04:05")
	return r.cmd.Channel.Name + " " + time
}

func (r *runner) User() string {
	return r.cmd.User.Name
}

func (r *runner) Authenticate() *sharefile.Login {
	return r.login
}

func (r *runner) EventCallbackURL() string {
	return r.urls.EventWebhook
}
