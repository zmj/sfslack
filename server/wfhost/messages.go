package wfhost

import (
	"fmt"
	"strings"

	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

const (
	wfTypeQueryKey = "wftype"
)

func loginMessage(loginURL string) slack.Message {
	return slack.Message{
		Text: fmt.Sprintf("Please %v", slack.FormatURL(loginURL, "log in")),
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

// when is init necessary?
var wfTypes = []*workflow.Definition{
	workflow.Definitions.Send,
	workflow.Definitions.Request,
}

func helpMessage(wfClickURL string) slack.Message {
	var links []string
	for _, def := range wfTypes {
		link := fmt.Sprintf("%v&%v=%v", wfClickURL, wfTypeQueryKey, def.Arg)
		links = append(links, slack.FormatURL(link, def.Description))
	}
	return slack.Message{
		Text: strings.Join(links, " | "),
	}
}
