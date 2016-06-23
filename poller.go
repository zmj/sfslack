package main

import "time"

type FolderPoller struct {
	Sf       SfLogin
	FolderId string
	NewItems chan []SfItem
	Quit     chan struct{}
}

func (sf SfLogin) FolderPoller(folderId string) *FolderPoller {
	fp := &FolderPoller{
		sf,
		folderId,
		make(chan []SfItem),
		make(chan struct{})}
	return fp
}

func (fp *FolderPoller) PollForSend() {
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

func (fp *FolderPoller) PollForRequest() {
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
