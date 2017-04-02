package main

import (
	"flag"
	"fmt"

	"github.com/zmj/sfslack/log"

	"os"

	"time"

	"github.com/zmj/sfslack/server"
)

func main() {
	cfg := server.Config{}
	flag.StringVar(&cfg.Host, "host", "", "Public host address")
	flag.IntVar(&cfg.Port, "port", 8080, "Listen Port")
	flag.StringVar(&cfg.SfOAuthID, "sfoauthid", "", "ShareFile OAuth Client ID")
	flag.StringVar(&cfg.SfOAuthSecret, "sfoauthsecret", "", "ShareFile OAuth Client Secret")
	flag.StringVar(&cfg.SlackVerificationToken, "slackverifytoken", "", "Slack command verification token")
	flag.Parse()
	// validate before open log, not in server

	logFileName := fmt.Sprintf("logs/%v.log", time.Now().Format(time.Stamp))
	logfile, err := os.Create(logFileName)
	if err != nil {
		err = fmt.Errorf("Failed to create log file\n%v", err)
		fmt.Println(err.Error())
		return
	}
	srv, err := server.NewServer(cfg, log.New(logfile, true))
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
