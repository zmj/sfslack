package main

import (
	"fmt"
	"net/http"

	"github.com/zmj/sfslack/secrets"
)

const (
	listenport = 8080
)

func main() {
	secrets, err := secrets.Load()
	if err != nil {
		fmt.Println(err)
		return
	}
	srv := newServer(secrets)
	err = (&http.Server{
		Addr:    fmt.Sprintf(":%v", listenport),
		Handler: srv.handler(),
	}).ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}
