package slack

import (
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
	Received    time.Time
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
		Received:    time.Now(),
	}
	// validate
	return command, nil
}
