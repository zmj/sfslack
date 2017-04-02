package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/zmj/sfslack/log"

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
	log     *log.Logger
}

func NewServer(cfg Config, log *log.Logger) (*http.Server, error) {
	err := cfg.validate()
	if err != nil {
		return nil, fmt.Errorf("Bad config: %v", err)
	}
	srv := &server{
		config: cfg,
		wfSvc:  wfhost.New(sfauth.New(cfg.SfOAuthID, cfg.SfOAuthSecret), log),
		log:    log,
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
			err = fmt.Errorf("Couldn't parse wfID: %v", err)
			http.Error(wr, err.Error(), http.StatusBadRequest)
			srv.logErr(err)
			return
		}
		wf, ok := srv.wfSvc.Get(wfID)
		if !ok {
			err = fmt.Errorf("Workflow not found '%v'", wfID)
			http.Error(wr, err.Error(), http.StatusInternalServerError)
			srv.logErr(err)
			return
		}
		h(wf, wr, req)
	}
}

func (srv *server) logErr(err error) {
	if err == nil {
		return
	}
	srv.log.Err(err)
}

func printReq(wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))
	http.Error(wr, "", http.StatusNotFound)
}
