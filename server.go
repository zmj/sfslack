package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
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
	mux.HandleFunc("/sfslack/send", s.SlackCommand)
	mux.HandleFunc("/sfslack/request", s.SlackCommand)
	mux.HandleFunc("/sfslack/auth", s.AuthCallback)
	return mux
}

func (s *Server) Print(wr http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.String())
	dbgReq(req)
	wr.Write([]byte("hello"))
}

func ParseCommand(req *http.Request) (SlackCommand, error) {
	var values url.Values
	var err error
	if req.Method == "GET" {
		values = req.URL.Query()
	} else if req.Method == "POST" {
		err = req.ParseForm()
		if err != nil {
			return SlackCommand{}, err
		}
		values = req.PostForm
	} else {
		return SlackCommand{}, errors.New("Unsupported HTTP method " + req.Method)
	}
	return NewCommand(values)
}

func (s *Server) SlackCommand(wr http.ResponseWriter, req *http.Request) {
	cmd, err := ParseCommand(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	wf := SlackWorkflow{cmd.User, make(chan SlackMessage), make(chan struct{})}
	switch cmd.Command {
	case "/sfsend":
		go func() { wf.Send(s.Auth.Authenticate(wf)) }()
	case "/sfrequest":
		go func() { wf.Request(s.Auth.Authenticate(wf)) }()
	default:
		http.Error(wr, "Unknown command", http.StatusBadRequest)
		close(wf.Responses)
		close(wf.Quit)
		return
	}
	firstResponse := <-wf.Responses
	go func() {
		defer close(wf.Quit)
		for sent := 0; sent < maxSlackResponses; sent++ {
			select {
			case <-time.After(maxSlackMessageTime):
				return
			case msg, ok := <-wf.Responses:
				if !ok {
					return
				}
				err = msg.RespondTo(cmd)
				if err != nil {
					fmt.Println("Failed to send ", msg, err)
					return
				}
			}
		}
	}()
	firstResponse.WriteTo(wr)
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
		}
	case <-time.After(10 * time.Second):
	}
	wr.Write([]byte("Logged in! You may close this page."))
}
