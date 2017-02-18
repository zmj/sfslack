package main

import (
	"context"

	"time"

	"github.com/zmj/sfslack/slack"
)

type sendFilesWf struct {
	slashCommandWf
}

func SendFilesWorkflow(cmd slack.Command) (SlashCommandWorkflow, error) {
	return &sendFilesWf{
		slashCommandWf{
			started:       time.Now(),
			cmd:           cmd,
			noMoreReplies: make(chan struct{}),
		},
	}, nil
}

func (wf *sendFilesWf) Start(ctx context.Context) (slack.Message, error) {
	return wf.start(ctx, wf.worker)
}

func (wf *sendFilesWf) worker(ctx context.Context, replies chan<- slack.Message) {
	defer close(replies)
	auth := wf.getAuth(ctx)
	if auth.StartURL != "" {
		replies <- loginMessage(auth.StartURL)
	}
	<-auth.Done
	if auth.Err != nil {
		replies <- errorMessage(auth.Err)
		return
	}
	folder, err := wf.createWorkflowFolder(ctx, auth.Login)
	if err != nil {
		replies <- errorMessage(err)
		return
	}
	folder = folder
	// start subscriber
	// make request share
	// wait for subscribe?
	// notify/redirect user with request share link
	// range over subscription
}
