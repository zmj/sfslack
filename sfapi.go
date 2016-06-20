package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

	fmt.Println(created)
	return created, nil
}
