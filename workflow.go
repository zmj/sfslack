package main

import (
	"context"
	"fmt"

	"time"

	sf "github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

type SlashCommandWorkflow interface {
	Start(context.Context) (slack.Message, error)
}

type authGetter func(context.Context) *sf.AuthRequest

type slashCommandWf struct {
	started       time.Time
	getAuth       authGetter
	cmd           slack.Command
	noMoreReplies chan struct{}
	err           error
}

type workflowWorker func(context.Context, chan<- slack.Message)

func (wf *slashCommandWf) start(ctx context.Context, worker workflowWorker) (slack.Message, error) {
	c := make(chan slack.Message)
	go func() {
		wfCtx, wfCancel := context.WithTimeout(context.Background(), slack.DelayedReplyTimeout)
		// copy context values
		defer wfCancel()
		worker(wfCtx, c)
	}()
	var firstReply slack.Message
	select {
	case <-ctx.Done():
		firstReply = timeoutMessage
	case msg, ok := <-c:
		if ok {
			firstReply = msg
		}
	}
	go wf.sendDelayedReplies(c)
	return firstReply, wf.err
}

func (wf *slashCommandWf) sendDelayedReplies(replies <-chan slack.Message) {
	remaining := slack.MaxDelayedReplies
	for msg := range replies {
		remaining--
		if remaining >= 0 {
			go msg.RespondTo(wf.cmd)
		}
		if remaining == 0 {
			close(wf.noMoreReplies)
		}
	}
}

func errorMessage(err error) slack.Message {
	return slack.Message{Text: "Error:" + err.Error()}
}

func loginMessage(url string) slack.Message {
	return slack.Message{Text: fmt.Sprintf("Please log in: %v", url)}
}

const (
	slackFolderName = ".slack"
	nowFormat       = "2006-01-02 15:04:05"
)

func findOrCreateSlackFolder(ctx context.Context, sfapi sf.Login) (sf.Folder, error) {
	home, err := sfapi.GetChildren(ctx, "home")
	if err != nil {
		return sf.Folder{}, err
	}
	for _, item := range home {
		if item.FileName == slackFolderName {
			folder, err := item.Folder()
			if err != nil {
				return sf.Folder{}, err
			}
			return folder, nil
		}
	}
	return sfapi.CreateFolder(slackFolderName, "home")
}

func (wf *slashCommandWf) folderName() string {
	return wf.started.Format(nowFormat)
}

func (wf *slashCommandWf) createWorkflowFolder(ctx context.Context, sfapi sf.Login) (sf.Folder, error) {
	slackFolder, err := findOrCreateSlackFolder(ctx, sfapi)
	if err != nil {
		return sf.Folder{}, err
	}
	return sfapi.CreateFolder(ctx, wf.folderName(), slackFolder.Id)
}

var timeoutMessage = slack.Message{Text: "Logging you in..."}
