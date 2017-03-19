package server

import (
	"sync"

	"github.com/zmj/sfslack/slack"
)

type replier struct {
	mu          *sync.Mutex
	redirects   []func(string) bool
	useCurrent  bool
	currentURL  string
	firstReply  chan slack.Message
	repliesSent int
	cmd         slack.Command

	replies chan reply
}

func (r *replier) Reply(msg slack.Message) {
	r.replies <- reply{msg, ""}
}

func (r *replier) RedirectOrReply(url string, msg slack.Message) {
	r.replies <- reply{msg, url}
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
}

func (r *replier) replyInner(re reply) {
	r.mu.Lock()
	cbs := r.redirects
	r.redirects = nil
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

func (r *replier) GetRedirect(cb func(string) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.useCurrent {
		cb(r.currentURL)
		return
	}
	r.redirects = append(r.redirects, cb)
}

func (r *replier) Working() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.useCurrent = false
}
