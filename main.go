package main

import (
	"fmt"

	"github.com/zmj/sfslack/secrets"
	"github.com/zmj/sfslack/server"
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
	srv := server.NewServer(secrets)
	err = srv.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}
