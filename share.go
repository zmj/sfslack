package main

import (
	"fmt"
	"time"
)

const (
	slackFolderName = ".slack"
	nowFormat       = "2006-01-02 03:04:05PM"
)

type SlackWorkflow struct {
	User      SlackUser
	Responses chan SlackResponse
	Quit      chan struct{}
}

func (sw SlackWorkflow) SendError(err error) {
	fmt.Println("Notifying ", sw.User.Name, " error: ", err.Error())
	sw.Responses <- SlackResponse{Text: "Error: " + err.Error()}
}

func (sw SlackWorkflow) Authenticate() SfLogin {
	// send request, wait on it
	return TestLogin()
}

func (sw SlackWorkflow) Request() {
	defer close(sw.Responses)
	sf := sw.Authenticate()
	// what if quit while in auth? need to select quit/authreq
	poller, url, err := SetupRequestShare(sf)
	if err != nil {
		sw.SendError(err)
		return
	}
	defer close(poller.Quit)
	msg := SlackResponse{
		Text:         sw.User.Name + " is requesting files: " + url,
		ResponseType: "in_channel"}
	sw.Responses <- msg
	for {
		select {
		case <-sw.Quit:
			return
		case newItems, ok := <-poller.NewItems:
			if !ok {
				return
			}
			// consolidate to single msg
			// download link?
			for _, item := range newItems {
				sw.Responses <- SlackResponse{Text: item.FileName + " was sent to you."}
			}
		}
	}
}

func (sw SlackWorkflow) Send() {
	defer close(sw.Responses)
	sf := sw.Authenticate()
	// wrong, need to select here to exit cleanly? or signal quit?
	poller, url, err := SetupRequestShare(sf)
	if err != nil {
		sw.SendError(err)
		return
	}
	defer close(poller.Quit)
	sw.Responses <- SlackResponse{Text: "Upload your files: " + url}
	for {
		select {
		case <-sw.Quit:
			return
		case newItems, ok := <-poller.NewItems:
			if !ok {
				return
			}
			fileIds := make([]string, 0, 0)
			for _, item := range newItems {
				if file, err := item.File(); err == nil {
					fileIds = append(fileIds, file.Id)
				}
			}
			// 0 files?
			sendShare, err := sf.CreateSendShare(fileIds)
			if err != nil {
				sw.SendError(err)
				return
			}
			// list files
			msg := SlackResponse{
				Text:         sw.User.Name + " has shared files: " + sendShare.Uri,
				ResponseType: "in_channel"}
			sw.Responses <- msg
			return
		}
	}
}

func SetupRequestShare(sf SfLogin) (*FolderPoller, string, error) {
	channelName := "channel"
	requestTime := time.Now().Format(nowFormat)
	slackFolder, err := sf.FindOrCreateSlackFolder()
	if err != nil {
		return nil, "", err
	}
	folderName := channelName + " " + requestTime
	shareFolder, err := sf.CreateFolder(folderName, slackFolder.Id)
	if err != nil {
		return nil, "", err
	}
	share, err := sf.CreateRequestShare(shareFolder.Id)
	if err != nil {
		// cleanup folder?
		return nil, "", err
	}
	return sf.PollFolder(shareFolder.Id), share.Uri, nil
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
			newItems := make([]SfItem, 0, 0)
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
