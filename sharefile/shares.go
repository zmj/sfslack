package sharefile

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

func (sf Login) CreateRequestShare(parentFolderId string) (Share, error) {
	toCreate := Share{ShareType: "Request",
		Parent: Folder{Item: Item{URL: sf.itemURL("Items", parentFolderId)}}}
	return sf.CreateShare(toCreate)
}

func (sf Login) CreateSendShare(files []File) (Share, error) {
	toCreate := Share{ShareType: "Send"}
	for _, file := range files {
		toCreate.Items = append(toCreate.Items, File{Item{URL: sf.itemURL("Items", file.ID)}})
	}
	return sf.CreateShare(toCreate)
}

func (sf Login) CreateShare(toCreate Share) (Share, error) {
	toSend, err := json.Marshal(toCreate)
	if err != nil {
		return Share{}, err
	}
	req, err := http.NewRequest("POST",
		sf.entityURL("Shares"),
		bytes.NewReader(toSend))
	if err != nil {
		return Share{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = sf.withCredentials(req)

	hc := http.Client{Jar: sf.cookies}
	resp, err := hc.Do(req)
	if err != nil {
		return Share{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Share{}, errors.New(resp.Status)
	}

	created := Share{}
	err = json.NewDecoder(resp.Body).Decode(&created)
	if err != nil {
		return Share{}, err
	}
	created.Account = sf.Account

	return created, nil
}

func (sh Share) DownloadAllURL() string {
	return sh.Account.itemURL("Shares", sh.ID) + "/Download"
}

func (sh Share) DownloadURL(fileID string) string {
	return sh.Account.itemURL("Shares", sh.ID) + "/Download(" + fileID + ")"
}
