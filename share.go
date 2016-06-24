package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	slackFolderName = ".slack"
	nowFormat       = "2006-01-02 15:04:05"
)

type SlackWorkflow struct {
	User      SlackUser
	Responses chan SlackMessage
	Quit      chan struct{}
}

func (wf SlackWorkflow) SendError(err error) {
	fmt.Println("Notifying", wf.User.Name, "error:", err.Error())
	wf.Responses <- SlackMessage{Text: "Error:" + err.Error()}
}

func (wf SlackWorkflow) Request(authReq chan Auth) {
	defer close(wf.Responses)
	var auth Auth
	select {
	case <-wf.Quit:
		return
	case auth = <-authReq:
	}
	sf := auth.Login
	if auth.Redirect != nil {
		close(auth.Redirect)
	}
	folder, share, err := SetupRequestShare(sf)
	if err != nil {
		wf.SendError(err)
		return
	}
	msg := SlackMessage{
		Text:         fmt.Sprintf("%v is requesting files: %v", wf.User.Name, share.Uri),
		ResponseType: "in_channel"}
	wf.Responses <- msg
	poller := sf.FolderPoller(folder.Id)
	go poller.PollForRequest()
	defer close(poller.Quit)
	for {
		select {
		case <-wf.Quit:
			return
		case newItems, ok := <-poller.NewItems:
			if !ok {
				return
			}
			var files []SfFile
			for _, item := range newItems {
				if file, err := item.File(); err == nil {
					files = append(files, file)
				}
			}
			if len(files) == 0 {
				continue
			}
			sendShare, err := sf.CreateSendShare(files)
			if err != nil {
				wf.SendError(err)
				continue
			}
			wf.Responses <- sendShare.BuildRequestNotification(files)
		}
	}
}

func (share SfShare) BuildRequestNotification(files []SfFile) SlackMessage {
	var msg SlackMessage
	if len(files) == 1 {
		msg.Text = fmt.Sprintf("Received %v: %v", files[0].FileName, share.DownloadAllUrl())
	} else {
		msg.Text = fmt.Sprintf("Received %v files: %v", len(files), share.DownloadAllUrl())
		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, file.FileName)
		}
		msg.Attachments = []SlackAttachment{
			SlackAttachment{
				Text:     strings.Join(fileNames, "\n"),
				Fallback: strings.Join(fileNames, " "),
			},
		}
	}
	return msg
}

func (wf SlackWorkflow) Send(authReq chan Auth) {
	defer close(wf.Responses)
	var auth Auth
	select {
	case <-wf.Quit:
		return
	case auth = <-authReq:
	}
	sf := auth.Login
	folder, share, err := SetupRequestShare(sf)
	if err != nil {
		wf.SendError(err)
		return
	}
	if auth.Redirect == nil {
		wf.Responses <- SlackMessage{Text: fmt.Sprintf("Upload your files: %v", share.Uri)}
	} else {
		auth.Redirect <- share.Uri
		close(auth.Redirect)
	}
	poller := sf.FolderPoller(folder.Id)
	go poller.PollForSend()
	defer close(poller.Quit)
	for {
		select {
		case <-wf.Quit:
			return
		case newItems, ok := <-poller.NewItems:
			if !ok {
				return
			}
			var files []SfFile
			for _, item := range newItems {
				if file, err := item.File(); err == nil {
					files = append(files, file)
				}
			}
			if len(files) == 0 {
				continue
			}
			sendShare, err := sf.CreateSendShare(files)
			if err != nil {
				wf.SendError(err)
				return
			}
			wf.Responses <- sendShare.BuildSendNotification(files, wf.User)
			return
		}
	}
}

func (share SfShare) BuildSendNotification(files []SfFile, slackUser SlackUser) SlackMessage {
	msg := SlackMessage{ResponseType: "in_channel"}
	if len(files) == 1 {
		msg.Text = fmt.Sprintf("%v has shared %v: %v", slackUser.Name, files[0].FileName, share.DownloadAllUrl())
	} else {
		msg.Text = fmt.Sprintf("%v has shared %v files: %v", slackUser.Name, len(files), share.DownloadAllUrl())
		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, file.FileName)
		}
		msg.Attachments = []SlackAttachment{
			SlackAttachment{
				Text:     strings.Join(fileNames, "\n"),
				Fallback: strings.Join(fileNames, " "),
			},
		}
	}
	return msg
}

func SetupRequestShare(sf SfLogin) (SfFolder, SfShare, error) {
	channelName := "channel"
	requestTime := time.Now().Format(nowFormat)
	slackFolder, err := sf.FindOrCreateSlackFolder()
	if err != nil {
		return SfFolder{}, SfShare{}, err
	}
	folderName := channelName + " " + requestTime
	shareFolder, err := sf.CreateFolder(folderName, slackFolder.Id)
	if err != nil {
		return SfFolder{}, SfShare{}, err
	}
	share, err := sf.CreateRequestShare(shareFolder.Id)
	if err != nil {
		// cleanup folder?
		return SfFolder{}, SfShare{}, err
	}

	return shareFolder, share, nil
}

func (sf SfLogin) FindOrCreateSlackFolder() (SfFolder, error) {
	home, err := sf.GetChildren("home")
	if err != nil {
		return SfFolder{}, err
	}
	for _, item := range home {
		if item.FileName == slackFolderName {
			folder, err := item.Folder()
			if err != nil {
				return SfFolder{}, err
			}
			return folder, nil
		}
	}
	return sf.CreateFolder(slackFolderName, "home")
}
