package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"sync"

	"github.com/zmj/sfslack/sharefile"
)

const (
	publicHostHeader = "X-PUBLIC-HOST"
)

type server struct {
	authCache *sharefile.AuthCache
	workflows map[int]*runner
	mu        *sync.Mutex
	wfID      int
}

func NewServer(cfg Config) (*http.Server, error) {
	err := cfg.validate()
	if err != nil {
		return nil, err
	}
	srv := &server{
		authCache: sharefile.NewAuthCache(cfg.OAuthID, cfg.OAuthSecret),
		workflows: make(map[int]*runner),
		mu:        &sync.Mutex{},
	}
	return &http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.Port),
		Handler: srv.handler(),
	}, nil
}

func (srv *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", printReq)
	mux.HandleFunc(appAddPath, srv.AppAdded)
	mux.HandleFunc(commandPath, srv.newCommand)
	mux.HandleFunc(commandClickPath, srv.wfHandler(srv.newCommandClick))
	mux.HandleFunc(authPath, srv.wfHandler(srv.authCallback))
	mux.HandleFunc(eventPath, srv.wfHandler(srv.eventCallback))
	mux.HandleFunc(redirectPath, srv.wfHandler(srv.redirect))
	return mux
}

func (srv *server) wfHandler(h func(*runner, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		wfID, err := wfID(req)
		if err != nil {
			http.Error(wr, err.Error(), http.StatusBadRequest)
			return
		}
		wf, ok := srv.get(wfID)
		if !ok {
			http.Error(wr, "Workflow not found", http.StatusInternalServerError)
			return
		}
		h(wf, wr, req)
	}
}

func (srv *server) get(wfID int) (*runner, bool) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	r, ok := srv.workflows[wfID]
	return r, ok
}

func (srv *server) put(r *runner) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.wfID++
	r.wfID = srv.wfID
	srv.workflows[srv.wfID] = r
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
