package main

import "time"
import sf "github.com/zmj/sfslack/sharefile"

type FolderPoller struct {
	Sf       sf.Login
	FolderId string
	NewItems chan []sf.Item
	Quit     chan struct{}
}

func (sf sf.Login) FolderPoller(folderId string) *FolderPoller {
	fp := &FolderPoller{
		sf,
		folderId,
		make(chan []sf.Item),
		make(chan struct{})}
	return fp
}

func (fp *FolderPoller) PollForSend() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	defer close(fp.NewItems)
	known := make(map[string]bool)
	var uploaded []sf.Item
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
			if len(newItems) == 0 && len(uploaded) > 0 {
				fp.NewItems <- uploaded
				uploaded = nil
			} else if len(newItems) > 0 {
				for _, item := range newItems {
					uploaded = append(uploaded, item)
				}
			}
		case _, ok := <-fp.Quit:
			if !ok {
				return
			}
		}
	}
}

func (fp *FolderPoller) PollForRequest() {
	ticker := time.NewTicker(8 * time.Second)
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
			var newItems []sf.Item
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
