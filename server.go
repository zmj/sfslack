package main

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
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

func TestLogin() SfLogin {
	account := SfAccount{"jeffcombscom", "sharefile.com", "sf-api.com"}
	authCookie := http.Cookie{
		Name:  "SFAPI_AuthID",
		Value: "2718f716-aee5-4e86-9c57-41e10f6be1ae"}

	cookieUrl, _ := url.Parse(account.BaseUrl())
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(cookieUrl, []*http.Cookie{&authCookie})
	return SfLogin{account, jar}
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
	login := TestLogin()
	share, err := login.CreateRequestShare()
	if err != nil {
		io.WriteString(wr, "error\n")
		io.WriteString(wr, err.Error())
		return
	}
	io.WriteString(wr, share.Uri)
}

// /send/create
func (s *Server) SendCreate(wr http.ResponseWriter, req *http.Request) {
	io.WriteString(wr, "send create\n")
	login := TestLogin()
	fileId := "fi9f7e97-9ac6-8093-32f5-ebb5530009cf"
	share, err := login.CreateSendShare([]string{fileId})
	if err != nil {
		io.WriteString(wr, "error\n")
		io.WriteString(wr, err.Error())
		return
	}
	io.WriteString(wr, share.Uri)
}
