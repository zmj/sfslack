package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	OAuthId       = "n9UV16i2BeawbSsR8426cjHezF3cwX7o"
	OAuthSecret   = "gbPfE206XMZvkfFU26WJhJQMI3wW3itXVBaM0Fo0nv3lVhhH"
	OAuthRedirect = "www.empirimancy.net"
)

type AuthCache struct {
	mutex         sync.Mutex
	currentUserId int
	userIds       map[SlackUser]int
	logins        map[int]SfLogin
	pending       map[int][]chan Auth
}

func NewAuthCache() *AuthCache {
	ac := &AuthCache{
		userIds: make(map[SlackUser]int),
		logins:  make(map[int]SfLogin),
		pending: make(map[int][]chan Auth),
	}
	go ac.refreshLoop()
	return ac
}

func (ac *AuthCache) Authenticate(wf SlackWorkflow) chan Auth {
	sf, found := ac.getLogin(wf.User)
	if found {
		c := make(chan Auth, 1)
		c <- Auth{sf, nil}
		close(c)
		return c
	}
	url, c := ac.startLogin(wf.User)
	wf.Responses <- BuildLoginNotification(url)
	return c
}

type Auth struct {
	Login    SfLogin
	Redirect chan string // maybe nil
}

func (ac *AuthCache) FinishAuth(userId int, authCode SfOAuthCode, redirect chan string) {
	token, err := authCode.GetToken()
	if err != nil {
		fmt.Println("Auth finish failed", err.Error())
		return
	}
	jar, _ := cookiejar.New(nil)
	sf := SfLogin{token.SfAccount, token, jar}

	ac.mutex.Lock()
	ac.logins[userId] = sf
	pending := ac.pending[userId]
	ac.pending[userId] = nil
	ac.mutex.Unlock()
	for i, c := range pending {
		auth := Auth{sf, nil}
		mostRecent := i == len(pending)-1
		if mostRecent {
			auth.Redirect = redirect
		}
		c <- auth
	}
}

func BuildLoginNotification(url string) SlackMessage {
	return SlackMessage{Text: fmt.Sprint("Please log in: ", url)}
}

func (ac *AuthCache) getId(su SlackUser) int {
	if id, found := ac.userIds[su]; found {
		return id
	}
	ac.currentUserId++
	ac.userIds[su] = ac.currentUserId
	return ac.currentUserId
}

func (ac *AuthCache) getLogin(su SlackUser) (SfLogin, bool) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	id := ac.getId(su)
	sf, found := ac.logins[id]
	return sf, found
}

func (ac *AuthCache) startLogin(su SlackUser) (string, chan Auth) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	id := ac.getId(su)
	c := make(chan Auth)
	ac.pending[id] = append(ac.pending[id], c)
	return ac.loginUrl(id), c
}

func (ac *AuthCache) loginUrl(id int) string {
	return fmt.Sprintf("https://secure.sharefile.com/oauth/authorize?response_type=code&client_id=%v&redirect_uri=%v",
		OAuthId,
		url.QueryEscape(ac.callbackUrl(id)))
}

func (ac *AuthCache) callbackUrl(id int) string {
	return fmt.Sprintf("http://%v/sfslack/auth?userid=%v", OAuthRedirect, id)
}

func (ac *AuthCache) refreshLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			toRefresh := make(map[int]SfLogin)
			ac.mutex.Lock()
			for id, sf := range ac.logins {
				if sf.Token.ShouldRefresh() {
					toRefresh[id] = sf
				}
			}
			ac.mutex.Unlock()
			for id, sf := range toRefresh {
				newToken, err := sf.Token.Refresh()
				if err != nil {
					fmt.Println("Failed token refresh", err.Error())
					// remove login or no? if actually expired?
					continue
				}
				ac.mutex.Lock()
				sf.Token = newToken
				sf.Cookies, _ = cookiejar.New(nil)
				ac.logins[id] = sf
				ac.mutex.Unlock()
			}
		}
	}
}

type SfOAuthCode struct {
	Code string
	SfAccount
}

type SfOAuthToken struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	SfAccount
	ExpiresAt time.Time `json:"-"`
}

func ParseOAuthCode(values url.Values) (SfOAuthCode, error) {
	account := SfAccount{
		Subdomain:       values.Get("subdomain"),
		AppControlPlane: values.Get("appcp"),
		ApiControlPlane: values.Get("apicp"),
	}
	code := SfOAuthCode{
		SfAccount: account,
		Code:      values.Get("code"),
	}
	// validate
	return code, nil
}

func (sf SfOAuthCode) GetToken() (SfOAuthToken, error) {
	values := map[string]string{
		"client_id":     OAuthId,
		"client_secret": OAuthSecret,
		"code":          sf.Code,
		"grant_type":    "authorization_code",
	}
	return sf.TokenPost(values)
}

func (sf SfOAuthToken) Refresh() (SfOAuthToken, error) {
	values := map[string]string{
		"client_id":     OAuthId,
		"client_secret": OAuthSecret,
		"refresh_token": sf.RefreshToken,
		"grant_type":    "refresh_token",
	}
	return sf.TokenPost(values)
}

func (sf SfAccount) TokenPost(values map[string]string) (SfOAuthToken, error) {
	var valuePairs []string
	for k, v := range values {
		valuePairs = append(valuePairs, fmt.Sprintf("%v=%v", k, v))
	}
	toSend := strings.Join(valuePairs, "&")
	req, err := http.NewRequest("POST",
		fmt.Sprintf("https://%v.%v/oauth/token?requirev3=true", sf.Subdomain, sf.AppControlPlane),
		strings.NewReader(toSend))
	if err != nil {
		return SfOAuthToken{}, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return SfOAuthToken{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return SfOAuthToken{}, errors.New(resp.Status)
	}

	created := SfOAuthToken{}
	err = json.NewDecoder(resp.Body).Decode(&created)
	if err != nil {
		return SfOAuthToken{}, err
	}
	created.SetExpiresAt()

	return created, nil
}

func (sf SfOAuthToken) ShouldRefresh() bool {
	return sf.ExpiresAt.Sub(time.Now()).Hours() < 2
}

func (sf SfOAuthToken) SetExpiresAt() {
	d := time.Duration(sf.ExpiresIn) * time.Second
	sf.ExpiresAt = time.Now().Add(d)
}

func (sf SfLogin) AddHeaders(req *http.Request) {
	url, _ := url.Parse(fmt.Sprintf("https://%v.%v", sf.Subdomain, sf.ApiControlPlane))
	cookies := sf.Cookies.Cookies(url)
	if len(cookies) == 0 {
		req.Header.Add("Authorization", "Bearer "+sf.Token.AccessToken)
	}
}
