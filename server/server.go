package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/zmj/sfslack/server/wfhost"
	"github.com/zmj/sfslack/sharefile/sfauth"
)

const (
	slashCommand = "/sharefile"
)

type server struct {
	config  Config
	authSvc *sfauth.Cache // only used in wfsvc
	wfSvc   *wfhost.Cache
}

func NewServer(cfg Config) (*http.Server, error) {
	err := cfg.validate()
	if err != nil {
		return nil, fmt.Errorf("Bad config: %v", err)
	}
	srv := &server{
		config:  cfg,
		authSvc: sfauth.New(cfg.SfOAuthID, cfg.SfOAuthSecret),
		wfSvc:   wfhost.New(),
	}
	return &http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.Port),
		Handler: srv.handler(),
	}, nil
}

func (srv *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", printReq)
	mux.HandleFunc(installPath, srv.install)
	mux.HandleFunc(slackAuthPath, srv.slackAuth)
	mux.HandleFunc(commandPath, srv.newCommand)
	mux.HandleFunc(commandClickPath, srv.wfHandler(srv.newCommandClick))
	mux.HandleFunc(sfAuthPath, srv.wfHandler(srv.sfAuth))
	mux.HandleFunc(eventPath, srv.wfHandler(srv.eventCallback))
	mux.HandleFunc(redirectPath, srv.wfHandler(srv.redirect))
	return mux
}

type wfHandler func(*wfhost.Runner, http.ResponseWriter, *http.Request)

func (srv *server) wfHandler(h wfHandler) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		wfID, err := wfID(req)
		if err != nil {
			http.Error(wr, err.Error(), http.StatusBadRequest)
			return
		}
		wf, ok := srv.wfSvc.Get(wfID)
		if !ok {
			http.Error(wr, "Workflow not found", http.StatusInternalServerError)
			return
		}
		h(wf, wr, req)
	}
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
