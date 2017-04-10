package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

const (
	oauthAccessURL = "https://slack.com/api/oauth.access"
)

type AppOAuthCode struct {
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AppOAuthToken struct {
	AccessToken string        `json:"access_token"`
	Scope       string        `json:"scope"`
	TeamName    string        `json:"team_name"`
	TeamID      string        `json:"team_id"`
	Bot         BotOAuthToken `json:"bot"`
}

type BotOAuthToken struct {
	UserID      string `json:"bot_user_id"`
	AccessToken string `json:"bot_access_token"`
}

func (c AppOAuthCode) GetToken() (*AppOAuthToken, error) {
	body, err := toBody(c)
	if err != nil {
		return nil, fmt.Errorf("Failed to serialize code: %v", err)
	}
	req, err := http.NewRequest("POST", oauthAccessURL, body)
	if err != nil {
		return nil, fmt.Errorf("Failed to create token request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	b, _ := httputil.DumpRequestOut(req, true)
	fmt.Println(string(b))
	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Token request failed: %v", err)
	}
	b, _ = httputil.DumpResponse(resp, true)
	fmt.Println(string(b))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Token request failed: %v", err)
	}
	recv := &AppOAuthToken{}
	if recv != nil {
		err = json.NewDecoder(resp.Body).Decode(recv)
		if err != nil {
			return nil, fmt.Errorf("Token response deserialized failed: %v", err)
		}
	}
	return recv, nil
}
func toBody(send interface{}) (io.Reader, error) {
	var body io.Reader
	if send != nil {
		b, err := json.Marshal(send)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
	}
	return body, nil
}
