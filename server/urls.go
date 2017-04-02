package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	installPath      = "/sfslack/install"
	commandPath      = "/sfslack/command"
	commandClickPath = "/sfslack/command/click"
	sfAuthPath       = "/sfslack/sfoauth"
	eventPath        = "/sfslack/event"
	redirectPath     = "/sfslack/next"
	slackAuthPath    = "/sfslack/slackoauth"

	wfidQueryKey = "wfid"
)

type callbackUrls struct {
	host string
}

func (cb callbackUrls) CommandClick(wfID int) string {
	return wfURL(cb.host, commandClickPath, wfID)
}

func (cb callbackUrls) AuthCallback(wfID int) string {
	return wfURL(cb.host, sfAuthPath, wfID)
}

func (cb callbackUrls) EventWebhook(wfID int) string {
	return wfURL(cb.host, eventPath, wfID)
}

func (cb callbackUrls) Waiting(wfID int) string {
	return wfURL(cb.host, redirectPath, wfID)
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
		return 0, fmt.Errorf("Missing request values: %v", err)
	}
	wfidStr := values.Get(wfidQueryKey)
	if wfidStr == "" {
		return 0, fmt.Errorf("Missing '%v': %v", wfidQueryKey, req.RequestURI)
	}
	wfID, err := strconv.Atoi(wfidStr)
	if err != nil {
		return wfID, fmt.Errorf("Invalid wfID '%v' %v", wfID, err)
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

func (srv *server) publicHost(req *http.Request) string {
	host := srv.config.Host
	if host == "" {
		host = req.Host
	}
	return host
}
