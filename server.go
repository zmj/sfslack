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
	mux.HandleFunc("/sfslack/request/", s.Request)
	mux.HandleFunc("/sfslack/request/create", s.RequestCreate)
	return mux
}

// /request/id/upload
// /request/id/download
func (s *Server) Request(wr http.ResponseWriter, req *http.Request) {
	io.WriteString(wr, "req\n")
	io.WriteString(wr, req.URL.String())
}

// /request/create
func (s *Server) RequestCreate(wr http.ResponseWriter, req *http.Request) {
	io.WriteString(wr, "req create\n")
	io.WriteString(wr, req.URL.String())
}
