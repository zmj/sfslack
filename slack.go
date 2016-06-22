package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"fmt"
)

type SlackCommand struct {
	Token       string
	Command     string
	Text        string
	ResponseUrl string
	User        SlackUser
	Channel     SlackChannel
	Team        SlackTeam
}

type SlackTeam struct {
	Id     string
	Domain string
}

type SlackChannel struct {
	Team SlackTeam
	Id   string
	Name string
}

type SlackUser struct {
	Team SlackTeam
	Id   string
	Name string
}

func ParseCommand(values url.Values) (SlackCommand, error) {
	fmt.Println(values)
	team := SlackTeam{
		Id:     values.Get("team_id"),
		Domain: values.Get("team_domain"),
	}
	channel := SlackChannel{
		Team: team,
		Id:   values.Get("channel_id"),
		Name: values.Get("channel_name"),
	}
	user := SlackUser{
		Team: team,
		Id:   values.Get("user_id"),
		Name: values.Get("user_name"),
	}
	command := SlackCommand{
		Token:       values.Get("token"),
		Command:     values.Get("command"),
		Text:        values.Get("text"),
		ResponseUrl: values.Get("response_url"),
		User:        user,
		Channel:     channel,
		Team:        team,
	}
	// validate
	return command, nil
}

type SlackResponse struct {
	Text         string `json:"text"`
	ResponseType string `json:"response_type,omitempty"`
}

func (cmd SlackCommand) Respond(message string, toChannel bool) error {
	response := SlackResponse{}
	response.Text = message
	if toChannel {
		response.ResponseType = "in_channel"
	}

	toSend, err := json.Marshal(response)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST",
		cmd.ResponseUrl,
		bytes.NewReader(toSend))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	dbgReq(req)
	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dbgResp(resp)
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
