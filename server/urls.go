package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	commandPath      = "/sfslack/command"
	commandClickPath = "/sfslack/command/click"
	authPath         = "/sfslack/auth"
	eventPath        = "/sfslack/event"
	redirectPath     = "/sfslack/wait"

	wfidQueryKey   = "wfid"
	wfTypeQueryKey = "wftype"
)

type callbackURLs struct {
	CommandClick string
	AuthCallback string
	EventWebhook string
	Waiting      string
}

func (s *server) callbackURLs(host string, wfID int) callbackURLs {
	return callbackURLs{
		CommandClick: wfURL(host, commandClickPath, wfID),
		AuthCallback: wfURL(host, authPath, wfID),
		EventWebhook: wfURL(host, eventPath, wfID),
		Waiting:      wfURL(host, redirectPath, wfID),
	}
}

func wfURL(host, path string, wfID int) string {
	return fmt.Sprintf("https://%v%v?%v=%v",
		host,
		path,
		wfidQueryKey,
		wfID)
}

func wfID(req *http.Request) (int, error) {
	values, err := httpValues(req)
	if err != nil {
		return 0, errors.New("Missing wfID")
	}
	wfidStr := values.Get(wfidQueryKey)
	if wfidStr == "" {
		return 0, errors.New("Missing wfID")
	}
	wfID, err := strconv.Atoi(wfidStr)
	if err != nil {
		return wfID, errors.New("Invalid wfID")
	}
	return wfID, nil
}

func httpValues(req *http.Request) (url.Values, error) {
	if req.Method == "GET" {
		return req.URL.Query(), nil
	} else if req.Method == "POST" {
		err := req.ParseForm()
		if err != nil {
			return url.Values{}, err
		}
		return req.PostForm, nil
	} else {
		return url.Values{}, fmt.Errorf("Unsupported HTTP method '%v'", req.Method)
	}
}

func publicHost(req *http.Request) string {
	host := req.Header.Get(publicHostHeader)
	if host == "" {
		host = req.URL.Host
	}
	return host
}
