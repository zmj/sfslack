package wfhost

import (
	"sync"

	"github.com/zmj/sfslack/sharefile/sfauth"
	"github.com/zmj/sfslack/slack"
)

type Cache struct {
	authSvc   *sfauth.Cache
	workflows map[int]*Runner
	mu        *sync.Mutex
	wfID      int
}

func New(authSvc *sfauth.Cache) *Cache {
	return &Cache{
		authSvc:   authSvc,
		workflows: make(map[int]*Runner),
		mu:        &sync.Mutex{},
	}
}

func (c *Cache) Get(wfID int) (*Runner, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	r, ok := c.workflows[wfID]
	return r, ok
}

func (c *Cache) New(cmd slack.Command, urls CallbackURLs) (*Runner, slack.Message) {
	first := make(chan slack.Message, 1)
	r := &Runner{
		replier: &replier{
			mu:         &sync.Mutex{},
			firstReply: first,
			cmd:        cmd,
			replies:    make(chan reply),
			done:       make(chan struct{}),
		},
		urls:    urls,
		authSvc: c.authSvc,
	}
	c.put(r)
	go r.run()
	go r.sendReplies()
	// needs timeout? inside runner
	return r, <-first
}

func (c *Cache) put(r *Runner) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.wfID++
	r.wfID = c.wfID
	c.workflows[c.wfID] = r
}
