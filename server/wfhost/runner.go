package wfhost

import (
	"net/url"

	"fmt"

	"github.com/zmj/sfslack/sharefile"

	"github.com/zmj/sfslack/sharefile/sfauth"
	"github.com/zmj/sfslack/workflow"
)

type Runner struct {
	*replier
	wfID    int
	urls    CallbackURLs
	authSvc *sfauth.Cache

	def     *workflow.Definition
	defWait chan struct{}

	login       *sharefile.Login
	loginValues url.Values
	loginWait   chan struct{}
}

func (r *Runner) run() {
	var err error
	defer func() {
		if err != nil {
			r.ReplyErr(err)
			r.log.Err(err)
		}
		close(r.done)
	}()
	// need shutdown on these two waits?
	// what's the case for external shtudown?
	r.def = r.getDefinition()
	login, err := r.getLogin()
	if err != nil {
		err = fmt.Errorf("Error during authentication: %v", err)
		return
	}
	r.login = login
	wf := r.def.Constructor(r)
	r.wf = wf
	err = wf.Setup()
	if err != nil {
		err = fmt.Errorf("Error during setup: %v", err)
		return
	}
	err = wf.Listen()
	if err != nil {
		err = fmt.Errorf("Error during listen: %v", err)
		return
	}
	err = wf.Cleanup()
	if err != nil {
		err = fmt.Errorf("Error during cleanup: %v", err)
		return
	}
}

func (r *Runner) getDefinition() *workflow.Definition {
	arg := r.cmd.Text
	for _, def := range wfTypes {
		if def.Arg == arg {
			return def
		}
	}

	r.mu.Lock()
	r.defWait = make(chan struct{})
	r.mu.Unlock()

	r.Reply(helpMessage(r.urls.CommandClick(r.wfID)))
	<-r.defWait
	return r.def
}

func (r *Runner) getLogin() (*sharefile.Login, error) {
	creds, ok := r.authSvc.TryGet(r.cmd.User)
	if !ok {
		r.mu.Lock()
		r.loginWait = make(chan struct{})
		r.mu.Unlock()

		authURL := r.urls.AuthCallback(r.wfID)
		loginURL := r.authSvc.LoginURL(authURL)
		r.RedirectOrReply(loginURL, loginMessage(loginURL))
		<-r.loginWait
		var err error
		creds, err = r.authSvc.Add(r.cmd.User, r.loginValues)
		if err != nil {
			return nil, fmt.Errorf("Failed to add auth: %v", err)
		}
	}
	return &sharefile.Login{creds}, nil
}

func (r *Runner) SetDefinition(values url.Values) error {
	arg := values.Get(wfTypeQueryKey)
	if arg == "" {
		return fmt.Errorf("Missing wftype value")
	}
	var def *workflow.Definition
	for _, d := range wfTypes {
		if d.Arg == arg {
			def = d
			break
		}
	}
	if def == nil {
		return fmt.Errorf("Unknown workflow type '%v'", arg)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.def != nil {
		// ok if same, redir
		return fmt.Errorf("Already started workflow '%v'", r.def.Description)
	}
	r.def = def
	r.useCurrent = false
	close(r.defWait)
	return nil
}

func (r *Runner) SetLogin(cbValues url.Values) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.loginValues != nil {
		return
	}
	r.loginValues = cbValues
	r.useCurrent = false
	close(r.loginWait)
}

func (r *Runner) Name() string {
	time := r.cmd.Received.Format("2006-01-02 15:04:05")
	return r.cmd.Channel.Name + " " + time
}

func (r *Runner) User() string {
	return r.cmd.User.Name
}

func (r *Runner) Authenticate() *sharefile.Login {
	return r.login
}

func (r *Runner) EventCallbackURL() string {
	return r.urls.EventWebhook(r.wfID)
}

func (r *Runner) WaitingURL() string {
	return r.urls.Waiting(r.wfID)
}

func (r *Runner) ErrorText(err error) string {
	return errorText(err)
}

func (r *Runner) Event(event sharefile.WebhookSubscriptionEvent) {
	r.wf.Event(event)
}
