package server

import (
	"sync"

	"github.com/zmj/sfslack/slack"
)

type replier struct {
	mu          *sync.Mutex
	waiting     []redirectCb
	useCurrent  bool
	currentURL  string
	firstReply  chan slack.Message
	repliesSent int
	cmd         slack.Command

	replies chan reply
}

func (r *replier) Reply(msg slack.Message) {
	r.replies <- reply{msg, "", nil}
}

func (r *replier) RedirectOrReply(url string, msg slack.Message) {
	r.replies <- reply{msg, url, nil}
}

func (r *replier) ReplyErr(err error) {
	r.replies <- reply{errorMessage(err), "", err}
}

func (r *replier) sendReplies() {
	for r.repliesSent < slack.MaxDelayedReplies {
		select {
		case re := <-r.replies:
			r.replyInner(re)
			// done
		}
	}
}

type reply struct {
	msg slack.Message
	url string
	err error
}

func (r *replier) replyInner(re reply) {
	r.mu.Lock()
	cbs := r.waiting
	r.waiting = nil
	r.currentURL = re.url
	r.useCurrent = true
	r.mu.Unlock()

	var accepted bool
	for _, cb := range cbs {
		accepted = accepted || cb(re.url)
	}
	if !accepted {
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

type redirectCb func(string) bool

func (r *replier) NextRedirect(cb redirectCb) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.useCurrent {
		cb(r.currentURL)
		return
	}
	r.waiting = append(r.waiting, cb)
}

func (r *replier) setWorking() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.useCurrent = false
}
