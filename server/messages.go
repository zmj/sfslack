package server

import (
	"fmt"
	"strings"

	"github.com/zmj/sfslack/slack"
)

func loginMessage(loginURL string) slack.Message {
	return slack.Message{
		Text: fmt.Sprintf("Please %v", slack.FormatURL(loginURL, "log in")),
	}
}

func helpMessage(wfClickURL string) slack.Message {
	var links []string
	for arg, def := range wfTypes {
		link := fmt.Sprintf("%v&%v=%v", wfClickURL, wfTypeQueryKey, arg)
		links = append(links, slack.FormatURL(link, def.Description))
	}
	return slack.Message{
		Text: strings.Join(links, " | "),
	}
}

func errorMessage(err error) slack.Message {
	return slack.Message{
		Text: errorText(err),
	}
}

func errorText(err error) string {
	return fmt.Sprintf("Oops, something went wrong: %v", err.Error())
}
