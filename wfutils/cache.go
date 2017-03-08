package wfutils

import (
	"sync"

	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

type Cache struct {
	mu        *sync.Mutex
	currentID int
	building  map[int]*Builder
	running   map[int]workflow.Workflow
}

type Builder struct {
	*workflow.Args
	WfID            int
	Definition      workflow.Definition
	CommandClickURL string
	AuthCallbackURL string
}

func NewCache() *Cache {
	return &Cache{
		mu:       &sync.Mutex{},
		building: make(map[int]*Builder),
		running:  make(map[int]workflow.Workflow),
	}
	// cleanup goroutine
}

func (c *Cache) NewBuilder(cmd slack.Command) *Builder {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentID++
	builder := &Builder{
		Args: &workflow.Args{Cmd: cmd},
		WfID: c.currentID,
	}
	c.building[c.currentID] = builder
	return builder
}

func (c *Cache) GetBuilder(wfID int) (*Builder, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	args, ok := c.building[wfID]
	return args, ok
}

func (c *Cache) delBuilder(wfID int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.building, wfID)
}

func (c *Cache) newRunning(wf workflow.Workflow) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentID++
	c.running[c.currentID] = wf
	return c.currentID
}

func (c *Cache) getRunning(wfID int) (workflow.Workflow, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	wf, ok := c.running[wfID]
	return wf, ok
}

func (c *Cache) putRunning(wfID int, wf workflow.Workflow) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.running[wfID] = wf
}

func (c *Cache) delRunning(wfID int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.running, wfID)
}
