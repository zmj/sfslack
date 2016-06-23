package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
)

type SfAccount struct {
	Subdomain       string `json:"subdomain,omitempty"`
	AppControlPlane string `json:"appcp,omitempty"`
	ApiControlPlane string `json:"apicp,omitempty"`
}

type SfLogin struct {
	SfAccount
	Token   SfOAuthToken
	Cookies http.CookieJar
}

type SfShare struct {
	Account   SfAccount `json:"-"`
	Id        string    `json:",omitempty"`
	Url       string    `json:"url,omitempty"`
	ShareType string    `json:",omitempty"`
	Parent    SfFolder  `json:",omitempty"`
	Items     []SfFile  `json:",omitempty"`
	Uri       string    `json:",omitempty"`
}

type SfItem struct {
	Id       string `json:",omitempty"`
	Url      string `json:"url,omitempty"`
	FileName string `json:",omitempty"`
}

type SfFile struct {
	SfItem
}

func (item SfItem) File() (SfFile, error) {
	if item.Id[0:2] != "fi" {
		return SfFile{}, errors.New("Not a file")
	}
	return SfFile{item}, nil
}

type SfFolder struct {
	SfItem
	Name string `json:",omitempty"`
}

func (item SfItem) Folder() (SfFolder, error) {
	if item.Id[0:2] != "fo" {
		return SfFolder{}, errors.New("Not a folder")
	}
	return SfFolder{SfItem: item}, nil
}

type SfItems struct {
	Items []SfItem `json:"value,omitempty"`
}

func (sf SfAccount) BaseUrl() string {
	return fmt.Sprintf("https://%v.%v/sf/v3", sf.Subdomain, sf.ApiControlPlane)
}

func (sf SfAccount) EntityUrl(entity string) string {
	return fmt.Sprintf("%v/%v", sf.BaseUrl(), entity)
}

func (sf SfAccount) ItemUrl(entity, id string) string {
	return fmt.Sprintf("%v(%v)", sf.EntityUrl(entity), id)
}

func (sf SfLogin) CreateRequestShare(parentFolderId string) (SfShare, error) {
	toCreate := SfShare{ShareType: "Request",
		Parent: SfFolder{SfItem: SfItem{Url: sf.ItemUrl("Items", parentFolderId)}}}
	return sf.CreateShare(toCreate)
}

func (sf SfLogin) CreateSendShare(files []SfFile) (SfShare, error) {
	toCreate := SfShare{ShareType: "Send"}
	for _, file := range files {
		toCreate.Items = append(toCreate.Items, SfFile{SfItem{Url: sf.ItemUrl("Items", file.Id)}})
	}
	return sf.CreateShare(toCreate)
}

func (sf SfLogin) CreateShare(toCreate SfShare) (SfShare, error) {
	toSend, err := json.Marshal(toCreate)
	if err != nil {
		return SfShare{}, err
	}
	req, err := http.NewRequest("POST",
		sf.EntityUrl("Shares"),
		bytes.NewReader(toSend))
	if err != nil {
		return SfShare{}, err
	}
	req.Header.Add("Content-Type", "application/json")

	hc := http.Client{Jar: sf.Cookies}
	resp, err := hc.Do(req)
	if err != nil {
		return SfShare{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return SfShare{}, errors.New(resp.Status)
	}

	created := SfShare{}
	err = json.NewDecoder(resp.Body).Decode(&created)
	if err != nil {
		return SfShare{}, err
	}
	created.Account = sf.SfAccount

	return created, nil
}

func (sf SfLogin) CreateFolder(name, parentFolderId string) (SfFolder, error) {
	toCreate := SfFolder{Name: name}

	toSend, err := json.Marshal(toCreate)
	if err != nil {
		return SfFolder{}, err
	}

	req, err := http.NewRequest("POST",
		sf.ItemUrl("Items", parentFolderId)+"/Folder",
		bytes.NewReader(toSend))
	if err != nil {
		return SfFolder{}, err
	}
	req.Header.Add("Content-Type", "application/json")

	hc := http.Client{Jar: sf.Cookies}
	resp, err := hc.Do(req)
	if err != nil {
		return SfFolder{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return SfFolder{}, errors.New(resp.Status)
	}

	created := SfFolder{}
	err = json.NewDecoder(resp.Body).Decode(&created)
	if err != nil {
		return SfFolder{}, err
	}

	return created, nil
}

func (sf SfLogin) GetChildren(parentFolderId string) ([]SfItem, error) {
	req, err := http.NewRequest("GET",
		sf.ItemUrl("Items", parentFolderId)+"/Children",
		nil)
	if err != nil {
		return nil, err
	}

	hc := http.Client{Jar: sf.Cookies}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	items := SfItems{}
	err = json.NewDecoder(resp.Body).Decode(&items)
	if err != nil {
		return nil, err
	}

	return items.Items, nil
}

func (sh SfShare) DownloadAllUrl() string {
	return sh.Account.ItemUrl("Shares", sh.Id) + "/Download"
}

func (sh SfShare) DownloadUrl(fileId string) string {
	return sh.Account.ItemUrl("Shares", sh.Id) + "/Download(" + fileId + ")"
}

func dbgReq(req *http.Request) {
	reqRaw, _ := httputil.DumpRequestOut(req, true)
	fmt.Println(string(reqRaw))
}
func dbgResp(resp *http.Response) {
	respRaw, _ := httputil.DumpResponse(resp, true)
	fmt.Println(string(respRaw))
}
