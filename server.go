package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/zmj/sfslack/slack"
)

func main() {
	s := &Server{NewAuthCache()}
	(&http.Server{
		Addr:    ":8619",
		Handler: s.Handler(),
	}).ListenAndServe()
	select {}
}

type Server struct {
	Auth *AuthCache
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

func ParseCommand(req *http.Request) (slack.Command, error) {
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

func mapToWorkflow(req *http.Request) (SlashCommandWorkflow, error) {
	cmd, err := ParseCommand(req)
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

func (srv *Server) Command(wr http.ResponseWriter, req *http.Request) {
	replyCtx, replied := context.WithTimeout(context.Background(), slack.InitialReplyTimeout)
	defer replied()

	workflow, err := mapToWorkflow(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	wfCtx, wfDone := context.WithTimeout(context.Background(), slack.DelayedReplyTimeout)
	go workflow.Start(replyCtx, wfCtx)
	first := <-workflow.Responses()
	go workflow.SendDelayedReplies(wfCtx, wfDone)
	first.WriteTo(wr)
}

func (s *Server) AuthCallback(wr http.ResponseWriter, req *http.Request) {
	userId, err := strconv.Atoi(req.URL.Query().Get("userid"))
	if err != nil {
		http.Error(wr, "Unable to parse auth id", http.StatusBadRequest)
		return
	}
	authCode, err := ParseOAuthCode(req.URL.Query())
	if err != nil {
		http.Error(wr, "Unable to parse OAuth token", http.StatusBadRequest)
		return
	}
	redirect := make(chan string)
	go s.Auth.FinishAuth(userId, authCode, redirect)
	select {
	case url := <-redirect:
		if len(url) > 0 {
			http.Redirect(wr, req, url, http.StatusFound)
			return
		}
	case <-time.After(10 * time.Second):
	}
	wr.Write([]byte("Logged in! You may close this page."))
}
