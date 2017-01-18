package main

import (
	"context"

	"github.com/zmj/sfslack/slack"
)

type requestFilesWf struct {
	slashCommandWf
}

func RequestFilesWorkflow(cmd slack.Command) (SlashCommandWorkflow, error) {
	return &requestFilesWf{
		slashCommandWf{
			cmd: cmd,
			// auth cache
			noMoreReplies: make(chan struct{}),
		},
	}, nil
}

func (wf *requestFilesWf) Start(ctx context.Context) (slack.Message, error) {
	return wf.start(ctx, wf.worker)
}

func (wf *requestFilesWf) worker(ctx context.Context, replies chan<- slack.Message) {

}
