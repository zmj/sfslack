package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"strconv"

	"time"

	"github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
	"sync"
)

const (
	publicHostHeader = "X-PUBLIC-HOST"
	wfidQueryKey     = "wfid"
	wfTypeQueryKey   = "wftype"
	redirectTimeout  = 3 * time.Second
)

type server struct {
	authCache *sharefile.AuthCache
	workflows *wfCache
}

type wfCache struct {
	mu &sync.Mutex
	wfID int
	runners map[int] *runner
}

func NewServer(cfg config) (*http.Server, error) {
	err := cfg.validate()
	if err != nil {
		return nil, err
	}
	srv := &Server{
		authCache: sharefile.NewAuthCache(cfg.OAuthID, cfg.OAuthSecret),
		wfCache:   &wfCache{
			mu: &sync.Mutex{},
			runners: make(map[int]*runner),
		},
	}
	return &http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.Port),
		Handler: srv.handler(),
	}, nil
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

func (c *wfCache) new(cmd slack.Command) *runner {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.wfID++
	r := newRunner(cmd, c.wfID)
	c.runners[c.wfID] = r
	return r
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
