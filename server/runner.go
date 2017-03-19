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
	srv       *server
	wfID      int
	defWait   chan *workflow.Definition
	login     *sharefile.Login
	loginWait chan url.Values
	wf        workflow.Workflow
	urls      callbackURLs
}

func (srv *server) new(cmd slack.Command, host string) (*runner, slack.Message) {
	first := make(chan slack.Message, 1)
	r := &runner{
		replier: &replier{
			mu:         &sync.Mutex{},
			firstReply: first,
			cmd:        cmd,
			replies:    make(chan reply),
		},
		srv: srv,
		// done?
	}
	srv.put(r)
	r.urls = srv.callbackURLs(host, r.wfID)
	go r.run()
	go r.sendReplies()
	return r, <-first
}

func (r *runner) run() {
	// need shutdown on these two waits
	def := r.getDefinition()

	login, err := r.getLogin()
	if err != nil {
		r.Reply(errorMessage(err))
		r.srv.logErr(err)
		return
	}
	r.login = login

	r.wf = def.Constructor(r)
	r.setWorking()
	err = r.wf.Setup()
	if err != nil {
		r.Reply(errorMessage(err))
		r.srv.logErr(err)
		return
	}

	r.wf.Listen()
}

func (r *runner) getDefinition() *workflow.Definition {
	def, ok := wfTypes[r.cmd.Text]
	if !ok {
		r.defWait = make(chan *workflow.Definition, 1)
		msg := helpMessage(r.urls.CommandClick)
		r.Reply(msg)
		def = <-r.defWait
	}
	return def
}

func (r *runner) getLogin() (*sharefile.Login, error) {
	login, ok := r.srv.authCache.TryGet(r.cmd.User)
	if !ok {
		r.loginWait = make(chan url.Values, 1)
		

		loginURL := r.srv.authCache.LoginURL(r.urls.AuthCallback)
		msg := loginMessage(loginURL)
		r.RedirectOrReply(loginURL, msg)
		authValues := <-r.loginWait
		r.setWorking()
		return r.srv.authCache.Add(r.cmd.User, authValues)
	}
	return login, nil
}

func (r *runner) SetDefinition(def *workflow.Definition) {
	c := r.defWait
	if c == nil {
		return
	}
	c <- def
}

func (r *runner) SetLogin(cbValues url.Values) {
	c := r.loginWait
	if c == nil {
		return
	}
	c <- cbValues
}

func (r *runner) Name() string {
	time := r.cmd.Received.Format("2006-01-02 15:04:05")
	return r.cmd.Channel.Name + " " + time
}

func (r *runner) Authenticate() *sharefile.Login {
	return r.login
}
