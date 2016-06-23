package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	slackFolderName = ".slack"
	nowFormat       = "2006-01-02 03:04:05PM" // change to 24hr for folder sort
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
	poller, share, err := SetupRequestShare(sf)
	if err != nil {
		wf.SendError(err)
		return
	}
	defer close(poller.Quit)
	msg := SlackMessage{
		Text:         wf.User.Name + " is requesting files: " + share.Uri,
		ResponseType: "in_channel"}
	wf.Responses <- msg
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
			wf.Responses <- share.BuildRequestNotification(files)
		}
	}
}

func (requestShare SfShare) BuildRequestNotification(files []SfFile) SlackMessage {
	var msg SlackMessage
	if len(files) == 1 {
		msg.Text = "Received " + files[0].FileName + ": " + requestShare.DownloadUrl(files[0].Id)
	} else {
		msg.Text = "Received " + string(len(files)) + " files: " + requestShare.DownloadAllUrl()
		for _, file := range files {
			msg.Attachments = append(msg.Attachments, SlackAttachment{
				Text:     file.FileName,
				Fallback: file.FileName,
			})
		}
	}
	return msg
}

func (wf SlackWorkflow) Send() {
	defer close(wf.Responses)
	sf := wf.Authenticate()
	// wrong, need to select here to exit cleanly? or signal quit?
	poller, share, err := SetupRequestShare(sf)
	if err != nil {
		wf.SendError(err)
		return
	}
	defer close(poller.Quit)
	wf.Responses <- SlackMessage{Text: "Upload your files: " + share.Uri}
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

func (sendShare SfShare) BuildSendNotification(files []SfFile, slackUser SlackUser) SlackMessage {
	var msg SlackMessage
	if len(files) == 1 {
		// download all url doesn't do zip for single file, looks better
		msg.Text = slackUser.Name + " has shared " + files[0].FileName + ": " + sendShare.DownloadAllUrl()
	} else {
		msg.Text = slackUser.Name + " has shared " + string(len(files)) + " files: " + sendShare.DownloadAllUrl()
		msg.ResponseType = "in_channel"
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

func SetupRequestShare(sf SfLogin) (*FolderPoller, SfShare, error) {
	channelName := "channel"
	requestTime := time.Now().Format(nowFormat)
	slackFolder, err := sf.FindOrCreateSlackFolder()
	if err != nil {
		return nil, SfShare{}, err
	}
	folderName := channelName + " " + requestTime
	shareFolder, err := sf.CreateFolder(folderName, slackFolder.Id)
	if err != nil {
		return nil, SfShare{}, err
	}
	share, err := sf.CreateRequestShare(shareFolder.Id)
	if err != nil {
		// cleanup folder?
		return nil, SfShare{}, err
	}
	return sf.PollFolder(shareFolder.Id), share, nil
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

type FolderPoller struct {
	Sf       SfLogin
	FolderId string
	NewItems chan []SfItem
	Quit     chan struct{}
}

func (sf SfLogin) PollFolder(folderId string) *FolderPoller {
	fp := &FolderPoller{
		sf,
		folderId,
		make(chan []SfItem),
		make(chan struct{})}
	go fp.Poll()
	return fp
}

func (fp *FolderPoller) Poll() {
	// probably want Timer for Reset(newPollTime) ?
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	defer close(fp.NewItems)
	known := make(map[string]bool)
	for {
		select {
		case <-ticker.C:
			items, err := fp.Sf.GetChildren(fp.FolderId)
			if err != nil {
				continue
			}
			var newItems []SfItem
			for _, item := range items {
				if !known[item.Id] {
					known[item.Id] = true
					newItems = append(newItems, item)
				}
			}
			if len(newItems) > 0 {
				fp.NewItems <- newItems
			}
		case _, ok := <-fp.Quit:
			if !ok {
				return
			}
		}
	}
}
