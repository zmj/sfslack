package main

import (
	"encoding/json"
	"io"
	"net/http"
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
	mux.HandleFunc("/sfslack/send", s.Send)
	mux.HandleFunc("/sfslack/request", s.Request)
	return mux
}

func (s *Server) Send(wr http.ResponseWriter, req *http.Request) {
	dbgReq(req)

	err := req.ParseForm()
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	cmd, err := ParseCommand(req.Form)
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
	dbgReq(req)

	err := req.ParseForm()
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	cmd, err := ParseCommand(req.Form)
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
	response := SlackResponse{Text: message, ResponseType: "in_channel"}

	encoder := json.NewEncoder(wr)
	err = encoder.Encode(response)
}
