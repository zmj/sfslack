package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SfShare struct {
	Id        string
	url       string
	ShareType string
	Parent    SfFolder
	Items     []SfFile
	Uri       string
}

type SfFile struct {
	Id  string
	url string
}

type SfFolder struct {
	Id  string
	url string
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
	toCreate := SfShare{Parent: SfFolder{url: sf.Account.ItemUrl("Items", "box")}}
	toSend, err := json.Marshal(toCreate)
	if err != nil {
		return SfShare{}, err
	}

	hc := http.Client{Jar: sf.Cookies}
	resp, err := hc.Post(sf.Account.EntityUrl("Share"), "application/json", bytes.NewReader(toSend))
	if err != nil {
		return SfShare{}, err
	}
	defer resp.Body.Close()

	received := make([]byte, resp.ContentLength, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, received)
	if err != nil {
		return SfShare{}, err
	}

	created := SfShare{}
	err = json.Unmarshal(received, &created)
	if err != nil {
		return SfShare{}, err
	}

	return created, nil
}
