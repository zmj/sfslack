package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	InitialReplyTimeout = 2 * time.Second
	MaxDelayedReplies   = 5
	DelayedReplyTimeout = 28 * time.Minute
)

type Command struct {
	Token       string
	Command     string
	Text        string
	ResponseURL string
	User        User
	Channel     Channel
	Team        Team
}

type Team struct {
	ID     string
	Domain string
}

type Channel struct {
	Team Team
	ID   string
	Name string
}

type User struct {
	Team Team
	ID   string
	Name string
}

func ParseCommand(values url.Values) (Command, error) {
	team := Team{
		ID:     values.Get("team_id"),
		Domain: values.Get("team_domain"),
	}
	channel := Channel{
		Team: team,
		ID:   values.Get("channel_id"),
		Name: values.Get("channel_name"),
	}
	user := User{
		Team: team,
		ID:   values.Get("user_id"),
		Name: values.Get("user_name"),
	}
	command := Command{
		Token:       values.Get("token"),
		Command:     values.Get("command"),
		Text:        values.Get("text"),
		ResponseURL: values.Get("response_url"),
		User:        user,
		Channel:     channel,
		Team:        team,
	}
	// validate
	return command, nil
}

type Message struct {
	Text         string       `json:"text"`
	ResponseType string       `json:"response_type,omitempty"`
	Attachments  []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Fallback string `json:"fallback,omitempty"`
	Text     string `json:"text,omitempty"`
}

func (msg Message) WriteTo(wr io.Writer) (int64, error) {
	toSend, err := json.Marshal(msg)
	if err != nil {
		return 0, err
	}
	written, err := wr.Write(toSend)
	return int64(written), err
}

func (msg Message) RespondTo(cmd Command) error {
	toSend, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST",
		cmd.ResponseURL,
		bytes.NewReader(toSend))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
