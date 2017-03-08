package wfutils

import "github.com/zmj/sfslack/workflow"

type Runner interface {
	Run(ResponseCallback) error
}

type ResponseCallback func(workflow.Response) (accepted bool)

func NewRunner(builder *Builder, cache *Cache) Runner {
	return &runner{
		builder:   builder,
		cache:     cache,
		responses: make(chan workflow.Response),
	}
}

type runner struct {
	builder       *Builder
	cache         *Cache
	responses     chan workflow.Response
	firstResponse ResponseCallback
}

func (r *runner) Run(cb ResponseCallback) error {
	r.firstResponse = cb
	r.builder.Reply = r.receive
	wf := r.builder.Definition.Constructor(*r.builder.Args)
	err := wf.Setup()
	if err != nil {
		return err
	}
	defer wf.Shutdown()
	for {
		select {
		case resp := <-r.responses:
			err := r.reply(resp)
			if err != nil {
				return err
			}
		case <-wf.Done():
			return wf.Err()
		}
	}
}

func (r *runner) reply(resp workflow.Response) error {
	if r.firstResponse != nil {
		sent := r.firstResponse(resp)
		r.firstResponse = nil
		if sent {
			return nil
		}
	}
	return resp.Msg.RespondTo(r.builder.Cmd)
}

func (r *runner) receive(resp workflow.Response) {
	go func() {
		r.responses <- resp
	}()
}
