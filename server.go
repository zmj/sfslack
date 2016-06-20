package main

import (
	"io"
	"net/http"
)

// "r737fd7880774cf98"
// https://jeffcombscom.sharefile.com/r-r737fd7880774cf98 (share.Uri)
// jeffcombscom
// sharefile.com

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

// /request/create
// /request/id/upload
// /request/id/download
// /send/create
// /send/id/upload
// /send/id/download

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/sfslack/request/", http.StripPrefix("/sfslack/request/", http.HandlerFunc(s.Request)))
	mux.HandleFunc("/sfslack/request/create", s.RequestCreate)
	mux.HandleFunc("/sfslack/send/create", s.SendCreate)
	return mux
}

// /request/id/upload
// /request/id/download
func (s *Server) Request(wr http.ResponseWriter, req *http.Request) {
	io.WriteString(wr, "req\n")
	login := TestLogin()
	files, err := login.GetChildren("home")
	if err != nil {
		io.WriteString(wr, "error\n")
		io.WriteString(wr, err.Error())
		return
	}
	for _, file := range files {
		io.WriteString(wr, file.Id+file.FileName+"\n")
	}
}

// /request/create
func (s *Server) RequestCreate(wr http.ResponseWriter, req *http.Request) {
	io.WriteString(wr, "req create\n")
	url, err := NewRequest()
	if err != nil {
		io.WriteString(wr, "error\n")
		io.WriteString(wr, err.Error())
		return
	}
	io.WriteString(wr, url)
}

// /send/create
func (s *Server) SendCreate(wr http.ResponseWriter, req *http.Request) {
	io.WriteString(wr, "req create\n")
	url, err := NewSend()
	if err != nil {
		io.WriteString(wr, "error\n")
		io.WriteString(wr, err.Error())
		return
	}
	io.WriteString(wr, url)
}
