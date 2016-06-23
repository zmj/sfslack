package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
		// leaked channels here - change flow
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
