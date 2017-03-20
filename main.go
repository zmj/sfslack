package main

import (
	"fmt"

	"flag"

	"github.com/zmj/sfslack/server"
)

func main() {
	cfg := server.Config{}
	flag.IntVar(&cfg.Port, "port", 8080, "Listen Port")
	flag.StringVar(&cfg.SfOAuthID, "oauthid", "", "OAuth Client ID")
	flag.StringVar(&cfg.SfOAuthSecret, "oauthsecret", "", "OAuth Client Secret")
	flag.Parse()

	srv, err := server.NewServer(cfg)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = srv.ListenAndServe()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
