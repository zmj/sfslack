package sharefile

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	tokenRefreshCheck = 30 * time.Minute
)

type AuthCache struct {
	mu          *sync.Mutex
	userLogins  map[interface{}]*Login
	oauthID     string
	oauthSecret string
}

func NewAuthCache(oauthID, oauthSecret string) *AuthCache {
	return &AuthCache{
		mu:          &sync.Mutex{},
		userLogins:  make(map[interface{}]*Login),
		oauthID:     oauthID,
		oauthSecret: oauthSecret,
	}
}

func (ac *AuthCache) TryGet(key interface{}) (*Login, bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	login, exists := ac.userLogins[key]
	return login, exists
}

func (ac *AuthCache) Add(key interface{}, oauthCode url.Values) (*Login, error) {
	code, err := parseOAuthCode(oauthCode)
	if err != nil {
		return nil, err
	}
	token, err := code.getToken(ac.oauthID, ac.oauthSecret)
	if err != nil {
		return nil, err
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	login := &Login{
		oauthToken: token,
		client:     &http.Client{},
	}
	ac.userLogins[key] = login
	go ac.refreshLoop(key)
	return login, nil
}

func (ac *AuthCache) LoginURL(callbackURL string) string {
	return fmt.Sprintf("https://secure.sharefile.com/oauth/authorize?response_type=code&client_id=%v&redirect_uri=%v",
		ac.oauthID,
		url.QueryEscape(callbackURL))
}

func (ac *AuthCache) refreshLoop(key interface{}) {
	t := time.NewTicker(tokenRefreshCheck)
	defer t.Stop()
	for expired := false; !expired; {
		select {
		case <-t.C:
			login, _ := ac.TryGet(key)
			if login.ExpiresAt.Before(time.Now()) {
				expired = true
				continue
			}
			if login.ExpiresAt.After(time.Now().Add(2 * time.Hour)) {
				continue
			}
			token, err := login.refresh(ac.oauthID, ac.oauthSecret)
			if err != nil {
				continue
			}
			login.oauthToken = token
		}
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	delete(ac.userLogins, key)
}
