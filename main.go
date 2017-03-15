package main

import (
	"fmt"

	"flag"

	"github.com/zmj/sfslack/server"
)

func main() {
	cfg := server.Config{}
	flag.IntVar(&cfg.Port, "port", 8080, "Listen Port")
	flag.StringVar(&cfg.OAuthID, "oauthid", "", "OAuth Client ID")
	flag.StringVar(&cfg.OAuthSecret, "oauthsecret", "" "OAuth Client Secret")
	flag.Parse()

	srv := server.NewServer(cfg)
	err = srv.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}
