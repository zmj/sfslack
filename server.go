package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func main() {
	s := &Server{}
	(&http.Server{
		Addr:    ":8619",
		Handler: s.Handler(),
	}).ListenAndServe()
	select {}
}

type Server struct {
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/sfslack/", s.Print)
	mux.HandleFunc("/sfslack/send", s.Send)
	mux.HandleFunc("/sfslack/request", s.Request)
	return mux
}

func (s *Server) Print(wr http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.String())
	dbgReq(req)
	wr.Write([]byte("hello"))
}

func ParseRequest(req *http.Request) (SlackCommand, error) {
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
	return ParseCommand(values)
}

func (s *Server) Send(wr http.ResponseWriter, req *http.Request) {
	cmd, err := ParseRequest(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	url, err := NewSend(cmd)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}

	message := "Upload your files: " + url
	io.WriteString(wr, message)
}

func (s *Server) Request(wr http.ResponseWriter, req *http.Request) {
	cmd, err := ParseRequest(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	url, err := NewRequest(cmd)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}

	message := cmd.User.Name + " is requesting files: " + url
	content := SlackResponse{Text: message, ResponseType: "in_channel"}
	toSend, err := json.Marshal(content)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}

	wr.Header().Add("Content-Type", "application/json")
	wr.Write(toSend)
}
