package server

import (
	"sync"

	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

type replier struct {
	mu          *sync.Mutex
	waiting     []redirectCb
	useCurrent  bool
	current     reply
	firstReply  chan slack.Message
	repliesSent int
	cmd         slack.Command
	wf          workflow.Workflow

	done    chan struct{}
	replies chan reply
}

func (r *replier) Reply(msg slack.Message) bool {
	c := make(chan bool, 1)
	r.replies <- reply{msg, "", nil, c}
	return <-c
}

func (r *replier) RedirectOrReply(url string, msg slack.Message) bool {
	c := make(chan bool, 1)
	r.replies <- reply{msg, url, nil, c}
	return <-c
}

func (r *replier) ReplyErr(err error) bool {
	c := make(chan bool, 1)
	r.replies <- reply{errorMessage(err), "", err, c}
	return <-c
}

func (r *replier) sendReplies() {
	defer func() {
		nope := <-r.replies
		nope.accepted <- false
	}()
	for r.repliesSent < slack.MaxDelayedReplies {
		select {
		case re := <-r.replies:
			re.accepted <- true
			r.replyInner(re)
		case <-r.done:
			return
		}
	}
}

type reply struct {
	msg      slack.Message
	url      string
	err      error
	accepted chan bool
}

func (r *replier) replyInner(re reply) {
	r.mu.Lock()
	cbs := r.waiting
	r.waiting = nil
	r.current = re
	r.useCurrent = true
	r.mu.Unlock()

	var accepted bool
	for _, cb := range cbs {
		accepted = cb(re.url, re.err) || accepted
	}
	if re.url == "" || !accepted {
		r.sendMsg(re.msg)
	}
}

func (r *replier) sendMsg(msg slack.Message) {
	if r.firstReply != nil {
		r.firstReply <- msg
		r.firstReply = nil
		return
	}

	err := msg.RespondTo(r.cmd)
	if err != nil {
		// log
		return
	}
	r.repliesSent++
}

type redirectCb func(string, error) bool

func (r *replier) NextRedirect(cb redirectCb) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.useCurrent {
		go cb(r.current.url, r.current.err)
	} else {
		r.waiting = append(r.waiting, cb)
	}
}
