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

func TestLogin() SfLogin {
	account := SfAccount{"jeffcombscom", "sharefile.com", "sf-api.com"}
	authCookie := http.Cookie{
		Name:  "SFAPI_AuthID",
		Value: "a7622b87-3fff-4caf-97dd-dd7ddb78057d"}

	cookieUrl, _ := url.Parse(account.BaseUrl())
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(cookieUrl, []*http.Cookie{&authCookie})
	return SfLogin{account, SfOAuthToken{}, jar}
}

type AuthCache struct {
	mutex         sync.Mutex
	currentUserId int
	userIds       map[SlackUser]int
	logins        map[int]SfLogin
	pending       map[int][]chan SfLogin
}

func NewAuthCache() *AuthCache {
	ac := &AuthCache{
		userIds: make(map[SlackUser]int),
		logins:  make(map[int]SfLogin),
		pending: make(map[int][]chan SfLogin),
	}
	go ac.refreshLoop()
	return ac
}

func (ac *AuthCache) Authenticate(wf SlackWorkflow) chan SfLogin {
	sf, found := ac.getLogin(wf.User)
	if found {
		c := make(chan SfLogin, 1)
		c <- sf
		close(c)
		return c
	}
	url, c := ac.startLogin(wf.User)
	wf.Responses <- BuildLoginNotification(url)
	return c
}

func (ac *AuthCache) FinishAuth(userId int, authCode SfOAuthCode) {
	token, err := authCode.GetToken()
	if err != nil {
		fmt.Println("Auth finish failed", err.Error())
		return
	}
	sf := SfLogin{token.SfAccount, token, nil}

	ac.mutex.Lock()
	ac.logins[userId] = sf
	pending := ac.pending[userId]
	ac.pending[userId] = nil
	ac.mutex.Unlock()
	for _, c := range pending {
		c <- sf
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

func (ac *AuthCache) startLogin(su SlackUser) (string, chan SfLogin) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	id := ac.getId(su)
	c := make(chan SfLogin)
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
			var toRefresh []SfLogin
			ac.mutex.Lock()
			for _, sf := range ac.logins {
				if sf.Token.ShouldRefresh() {
					toRefresh = append(toRefresh, sf)
				}
			}
			ac.mutex.Unlock()
			for _, sf := range toRefresh {
				newToken, err := sf.Token.Refresh()
				if err != nil {
					fmt.Println("Failed token refresh", err.Error())
					// remove login or no? if actually expired?
					continue
				}
				ac.mutex.Lock()
				sf.Token = newToken
				sf.Cookies = nil
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
		fmt.Sprintf("https://%v.%v/oauth/token", sf.Subdomain, sf.AppControlPlane),
		strings.NewReader(toSend))
	if err != nil {
		return SfOAuthToken{}, err
	}
	req.Header.Add("Content-Type", "application/json")

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
	if sf.Cookies == nil {
		req.Header.Add("Authorization", "Bearer "+sf.Token.AccessToken)
	}
	sf.Cookies, _ = cookiejar.New(nil)
}
