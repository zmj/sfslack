package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	sf "github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

func main() {
	s := &Server{sf.NewAuthCache("", "")}
	(&http.Server{
		Addr:    ":8080",
		Handler: s.Handler(),
	}).ListenAndServe()
}

type Server struct {
	Auth *sf.AuthCache
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/sfslack/", s.Print)
	mux.HandleFunc("/sfslack/send", s.Command)
	mux.HandleFunc("/sfslack/request", s.Command)
	mux.HandleFunc("/sfslack/auth", s.AuthCallback)
	return mux
}

func (s *Server) Print(wr http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.String())
	wr.Write([]byte("hello"))
}

func (srv *Server) Command(wr http.ResponseWriter, req *http.Request) {
	workflow, err := mapToWorkflow(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), slack.InitialReplyTimeout)
	defer cancel()
	reply, err := workflow.Start(ctx)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	reply.WriteTo(wr)
}

func mapToWorkflow(req *http.Request) (SlashCommandWorkflow, error) {
	cmd, err := parseCommand(req)
	if err != nil {
		return nil, err
	}
	var workflow SlashCommandWorkflow
	switch cmd.Command {
	case "/sfsend":
		workflow, err = SendFilesWorkflow(cmd)
	case "/sfrequest":
		workflow, err = RequestFilesWorkflow(cmd)
	default:
		err = errors.New("Unknown command")
	}
	return workflow, err
}

func parseCommand(req *http.Request) (slack.Command, error) {
	var values url.Values
	var err error
	if req.Method == "GET" {
		values = req.URL.Query()
	} else if req.Method == "POST" {
		err = req.ParseForm()
		if err != nil {
			return slack.Command{}, err
		}
		values = req.PostForm
	} else {
		return slack.Command{}, errors.New("Unsupported HTTP method " + req.Method)
	}
	return slack.NewCommand(values)
}

func (srv *Server) AuthCallback(wr http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	redirect, err := srv.Auth.Callback(ctx, req.URL.Query())
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}
	if redirect != "" {
		http.Redirect(wr, req, redirect, http.StatusFound)
		return
	}
	wr.Write([]byte("Logged in! You may close this page."))
}
