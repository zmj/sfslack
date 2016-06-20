package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
)

type SfShare struct {
	Id        string   `json:",omitempty"`
	Url       string   `json:"url,omitempty"`
	ShareType string   `json:",omitempty"`
	Parent    SfFolder `json:",omitempty"`
	Items     []SfFile `json:",omitempty"`
	Uri       string   `json:",omitempty"`
}

type SfFile struct {
	Id  string `json:",omitempty"`
	Url string `json:"url,omitempty"`
}

type SfFolder struct {
	Id  string `json:",omitempty"`
	Url string `json:"url,omitempty"`
}

type SfFiles struct {
	Items []SfFile `json:"value,omitempty"`
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

func (sf SfLogin) CreateRequestShare() (SfShare, error) {
	toCreate := SfShare{ShareType: "Request",
		Parent: SfFolder{Url: sf.Account.ItemUrl("Items", "box")}}
	return sf.CreateShare(toCreate)
}

func (sf SfLogin) CreateSendShare(fileIds []string) (SfShare, error) {
	toCreate := SfShare{ShareType: "Send"}
	for _, id := range fileIds {
		toCreate.Items = append(toCreate.Items, SfFile{Url: sf.Account.ItemUrl("Items", id)})
	}
	return sf.CreateShare(toCreate)
}

func (sf SfLogin) CreateShare(toCreate SfShare) (SfShare, error) {
	toSend, err := json.Marshal(toCreate)
	if err != nil {
		return SfShare{}, err
	}
	req, err := http.NewRequest(http.MethodPost,
		sf.Account.EntityUrl("Shares"),
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

	created := SfShare{}
	err = json.NewDecoder(resp.Body).Decode(&created)
	if err != nil {
		return SfShare{}, err
	}

	return created, nil
}

func (sf SfLogin) GetShareFiles(shareId string) ([]SfFile, error) {
	req, err := http.NewRequest(http.MethodGet,
		sf.Account.ItemUrl("Shares", shareId)+"/Items",
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

	items := SfFiles{}
	err = json.NewDecoder(resp.Body).Decode(&items)
	if err != nil {
		return nil, err
	}

	return items.Items, nil
}

func dbgReq(req *http.Request) {
	reqRaw, _ := httputil.DumpRequestOut(req, true)
	fmt.Println(string(reqRaw))
}
func dbgResp(resp *http.Response) {
	respRaw, _ := httputil.DumpResponse(resp, true)
	fmt.Println(string(respRaw))
}
