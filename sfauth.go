package main

import (
	"fmt"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	sf "github.com/zmj/sfslack/sharefile"
	"github.com/zmj/sfslack/slack"
)

const (
	OAuthRedirect = "www.empirimancy.net"
)

type AuthCache struct {
	mutex         sync.Mutex
	currentUserId int
	userIds       map[slack.User]int
	logins        map[int]sf.Login
	pending       map[int][]chan Auth
}

func NewAuthCache() *AuthCache {
	ac := &AuthCache{
		userIds: make(map[slack.User]int),
		logins:  make(map[int]sf.Login),
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
	Login    sf.Login
	Redirect chan string // maybe nil
}

func (ac *AuthCache) FinishAuth(userId int, authCode sf.OAuthCode, redirect chan string) {
	token, err := authCode.GetToken()
	if err != nil {
		fmt.Println("Auth finish failed", err.Error())
		return
	}
	jar, _ := cookiejar.New(nil)
	sf := sf.Login{token.Account, token, jar}

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

func BuildLoginNotification(url string) slack.Message {
	return slack.Message{Text: fmt.Sprintf("Please log in: %v", url)}
}

func (ac *AuthCache) getId(su slack.User) int {
	if id, found := ac.userIds[su]; found {
		return id
	}
	ac.currentUserId++
	ac.userIds[su] = ac.currentUserId
	return ac.currentUserId
}

func (ac *AuthCache) getLogin(su slack.User) (sf.Login, bool) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	id := ac.getId(su)
	sf, found := ac.logins[id]
	return sf, found
}

func (ac *AuthCache) startLogin(su slack.User) (string, chan Auth) {
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
			toRefresh := make(map[int]sf.Login)
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

func ParseOAuthCode(values url.Values) (sf.OAuthCode, error) {
	account := sf.Account{
		Subdomain:       values.Get("subdomain"),
		AppControlPlane: values.Get("appcp"),
		ApiControlPlane: values.Get("apicp"),
	}
	code := sf.OAuthCode{
		Account: account,
		Code:    values.Get("code"),
	}
	// validate
	return code, nil
}
