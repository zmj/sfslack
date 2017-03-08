package main

import (
	"sync"

	"time"

	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

type workflowCache struct {
	mu        *sync.Mutex
	currentID int
	building  map[int]*workflowBuilder
	running   map[int]workflow.Workflow
}

type workflowBuilder struct {
	*workflow.Args
	wfID            int
	started         time.Time
	definition      workflow.Definition
	commandClickURL string
	authCallbackURL string
}

func newWorkflowCache() *workflowCache {
	return &workflowCache{
		mu:       &sync.Mutex{},
		building: make(map[int]*workflowBuilder),
		running:  make(map[int]workflow.Workflow),
	}
	// cleanup goroutine
}

func (c *workflowCache) newBuilder(cmd slack.Command) *workflowBuilder {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentID++
	builder := &workflowBuilder{
		Args:    &workflow.Args{Cmd: cmd},
		wfID:    c.currentID,
		started: time.Now(),
	}
	c.building[c.currentID] = builder
	return builder
}

func (c *workflowCache) getBuilder(wfID int) (*workflowBuilder, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	args, ok := c.building[wfID]
	return args, ok
}

func (c *workflowCache) delBuilder(wfID int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.building, wfID)
}

func (c *workflowCache) newRunning(wf workflow.Workflow) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentID++
	c.running[c.currentID] = wf
	return c.currentID
}

func (c *workflowCache) getRunning(wfID int) (workflow.Workflow, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	wf, ok := c.running[wfID]
	return wf, ok
}

func (c *workflowCache) putRunning(wfID int, wf workflow.Workflow) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.running[wfID] = wf
}

func (c *workflowCache) delRunning(wfID int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.running, wfID)
}
