package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

const (
	slackFolderName = ".slack"
	nowFormat       = "2006-01-02 03:04:05PM"
)

func TestLogin() SfLogin {
	account := SfAccount{"jeffcombscom", "sharefile.com", "sf-api.com"}
	authCookie := http.Cookie{
		Name:  "SFAPI_AuthID",
		Value: "2718f716-aee5-4e86-9c57-41e10f6be1ae"}

	cookieUrl, _ := url.Parse(account.BaseUrl())
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(cookieUrl, []*http.Cookie{&authCookie})
	return SfLogin{account, jar}
}

func NewRequest() (string, error) {
	poller, url, err := SetupRequestShare()
	if err != nil {
		return "", err
	}
	go func() {
		defer close(poller.Quit)
	loop:
		for {
			select {
			case <-time.After(30 * time.Minute):
				break loop
			case items, ok := <-poller.NewItems:
				if !ok {
					break loop
				}
				for _, item := range items {
					// send uploaded items as command responses
					fmt.Println(item.Id + " " + item.FileName)
				}
			}
		}
	}()
	return url, nil
}

func NewSend() (string, error) {
	sf := TestLogin()
	poller, url, err := SetupRequestShare()
	if err != nil {
		return "", err
	}
	// create send share for uploaded items
	go func() {
		defer close(poller.Quit)
	loop:
		for {
			select {
			case <-time.After(10 * time.Minute):
				break loop
			case items, ok := <-poller.NewItems:
				if ok {
					fileIds := make([]string, 0, 0)
					for _, item := range items {
						if file, err := item.File(); err == nil {
							fileIds = append(fileIds, file.Id)
						}
					}
					sendShare, err := sf.CreateSendShare(fileIds)
					if err != nil {
						fmt.Println("failed to create send share")
						fmt.Println(err.Error())
					}
					// broadcast send share url
					fmt.Println(sendShare.Uri)
				}
				break loop
			}
		}
	}()
	return url, nil
}

func SetupRequestShare() (*FolderPoller, string, error) {
	channelName := "channel"
	requestTime := time.Now().Format(nowFormat)
	sf := TestLogin()
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
