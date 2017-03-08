package main

import (
	"net/http"
	"net/http/httputil"

	"fmt"

	"strings"

	"github.com/zmj/sfslack/slack"
	"github.com/zmj/sfslack/workflow"
)

const (
	commandPath      = "/sfslack/command"
	commandClickPath = "/sfslack/command/click"
)

var wfTypes = map[string]workflow.Definition{
	"send":    workflow.Definitions.Send,
	"request": workflow.Definitions.Request,
}

func (srv *server) newCommand(wr http.ResponseWriter, req *http.Request) {
	var respondErr error
	defer srv.logErr(respondErr)

	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	cmd, err := parseCommand(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	wr.Header().Add("Content-Type", "application/json")

	builder := srv.wfCache.newBuilder(cmd)
	builder = withCallbackURLs(builder, publicHost(req))

	def, ok := wfTypes[cmd.Text]
	if !ok {
		_, respondErr = helpMessage(builder.commandClickURL).WriteTo(wr)
		return
	}
	builder.definition = def

	login, ok := srv.authCache.TryGet(cmd.User)
	if !ok {
		loginURL := srv.authCache.LoginURL(builder.authCallbackURL)
		_, respondErr = loginMessage(loginURL).WriteTo(wr)
		return
	}
	builder.Sf = login

	msg := srv.startWorkflowForResponse()
	_, respondErr = msg.WriteTo(wr)
}

func (srv *server) newCommandClick(wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	wfID, err := workflowID(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	wfType := req.URL.Query().Get(wfTypeQueryKey)
	def, ok := wfTypes[wfType]
	if !ok {
		http.Error(wr, "Unknown workflow type", http.StatusBadRequest)
		return
	}

	builder, ok := srv.wfCache.getBuilder(wfID)
	if !ok {
		http.Error(wr, "Unknown workflow ID", http.StatusInternalServerError)
		return
	}
	builder.definition = def

	login, ok := srv.authCache.TryGet(builder.Cmd.User)
	if !ok {
		loginURL := srv.authCache.LoginURL(builder.authCallbackURL)
		http.Redirect(wr, req, loginURL, http.StatusFound)
		return
	}
	builder.Sf = login

	redirectURL := srv.startWorkflowForRedirect()
	if redirectURL == "" {
		wr.Write([]byte("Logged in! You may close this page."))
		return
	}
	http.Redirect(wr, req, redirectURL, http.StatusFound)
}

func parseCommand(req *http.Request) (slack.Command, error) {
	values, err := httpValues(req)
	if err != nil {
		return slack.Command{}, err
	}
	return slack.ParseCommand(values)
}

func loginMessage(loginURL string) slack.Message {
	return slack.Message{
		Text: fmt.Sprintf("Please log in: %v", loginURL),
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

func withCallbackURLs(builder *workflowBuilder, host string) *workflowBuilder {
	builder.authCallbackURL = authCallbackURL(host, builder.wfID)
	builder.commandClickURL = commandClickURL(host, builder.wfID)
	builder.EventURL = eventCallbackURL(host, builder.wfID)
	return builder
}

func commandClickURL(host string, wfID int) string {
	return fmt.Sprintf("https://%v%v?%v=%v",
		host,
		commandClickPath,
		wfidQueryKey,
		wfID)
}

func workingMessage() slack.Message {
	return slack.Message{
		Text: "Logging you in...",
	}
}
