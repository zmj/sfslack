package main

import (
	"context"

	"github.com/zmj/sfslack/slack"
)

type SlashCommandWorkflow interface {
	Start(context.Context, context.Context)
	SendDelayedReplies(context.Context, context.CancelFunc)
	Responses() <-chan slack.Message
}

type slashCommandWf struct {
	cmd       slack.Command
	responses chan slack.Message
}

func (wf *slashCommandWf) Responses() chan slack.Message {
	return wf.responses
}

func (wf *slashCommandWf) SendDelayedReplies(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	for sent := 0; sent < slack.MaxDelayedReplies; sent += 1 {
		select {
		case <-ctx.Done():
			// check error
			return
		case msg, ok := <-wf.responses:
			// errorrrrr
			msg.WriteTo(wf.cmd)
		}
	}
}

type sendFilesWf struct {
	slashCommandWf
}

type requestFilesWf struct {
	slashCommandWf
}

func SendFilesWorkflow(cmd slack.Command) (SlashCommandWorkflow, error) {
	return sendFilesWf{
		slashCommandWf{
			cmd:       cmd,
			responses: make(chan slack.Message),
		},
	}
}

func RequestFilesWorkflow(cmd slack.Command) (SlashCommandWorkflow, error) {
	return requestFilesWf{
		slashCommandWf{
			cmd:       cmd,
			responses: make(chan slack.Message),
		},
	}
}

func (wf *sendFilesWf) Start(replyCtx, wfCtx context.Context) {

}

func (wf *requestFilesWf) Start(replyCtx, wfCtx context.Context) {

}
