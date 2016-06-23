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
	fmt.Println("Notifying ", wf.User.Name, " error: ", err.Error())
	wf.Responses <- SlackMessage{Text: "Error: " + err.Error()}
}

func (wf SlackWorkflow) Authenticate() SfLogin {
	// send request, wait on it
	return TestLogin()
}

func (wf SlackWorkflow) Request() {
	defer close(wf.Responses)
	sf := wf.Authenticate()
	// what if quit while in auth? need to select quit/authreq
	folder, share, err := SetupRequestShare(sf)
	if err != nil {
		wf.SendError(err)
		return
	}
	msg := SlackMessage{
		Text:         wf.User.Name + " is requesting files: " + share.Uri,
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
		msg.Text = "Received " + files[0].FileName + ": " + share.DownloadAllUrl()
	} else {
		msg.Text = "Received " + string(len(files)) + " files: " + share.DownloadAllUrl()
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

func (wf SlackWorkflow) Send() {
	defer close(wf.Responses)
	sf := wf.Authenticate()
	// wrong, need to select here to exit cleanly? or signal quit?
	folder, share, err := SetupRequestShare(sf)
	if err != nil {
		wf.SendError(err)
		return
	}
	wf.Responses <- SlackMessage{Text: "Upload your files: " + share.Uri}
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
			files := make([]SfFile, 0, 0)
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
	var msg SlackMessage
	if len(files) == 1 {
		// download all url doesn't do zip for single file, looks better
		msg.Text = slackUser.Name + " has shared " + files[0].FileName + ": " + share.DownloadAllUrl()
	} else {
		msg.Text = slackUser.Name + " has shared " + string(len(files)) + " files: " + share.DownloadAllUrl()
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
	msg.ResponseType = "in_channel"
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
