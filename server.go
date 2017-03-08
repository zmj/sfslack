package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"strconv"

	"github.com/zmj/sfslack/secrets"
	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/wfutils"
)

const (
	publicHostHeader = "X-PUBLIC-HOST"
	wfidQueryKey     = "wfid"
	wfTypeQueryKey   = "wftype"
)

type server struct {
	authCache *sharefile.AuthCache
	wfCache   *wfutils.Cache
}

func newServer(secrets secrets.Secrets) *server {
	return &server{
		authCache: sharefile.NewAuthCache(secrets.OAuthID, secrets.OAuthSecret),
		wfCache:   wfutils.NewCache(),
	}
}

func (srv *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", printReq)
	mux.HandleFunc(commandPath, srv.newCommand)
	mux.HandleFunc(commandClickPath, srv.newCommandClick)
	mux.HandleFunc(authPath, srv.authCallback)
	mux.HandleFunc(eventPath, srv.eventCallback)
	return mux
}

func workflowID(req *http.Request) (int, error) {
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

func printReq(wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))
	http.Error(wr, "", http.StatusNotFound)
}

func (srv *server) logErr(err error) {
	if err == nil {
		return
	}
	// todo
}
