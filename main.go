package main

import (
	"fmt"

	"flag"

	"github.com/zmj/sfslack/server"
)

func main() {
	cfg := server.Config{}
	flag.IntVar(&cfg.Port, "port", 8080, "Listen Port")
	flag.StringVar(&cfg.SfOAuthID, "sfoauthid", "", "ShareFile OAuth Client ID")
	flag.StringVar(&cfg.SfOAuthSecret, "sfoauthsecret", "", "ShareFile OAuth Client Secret")
	flag.StringVar(&cfg.SlackVerificationToken, "slackverifytoken", "", "Slack command verification token")
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
