package sharefile

import (
	"fmt"
	"net/http/cookiejar"
	"net/url"
	"sync"
)

type AuthCache struct {
	mu          *sync.Mutex
	userLogins  map[interface{}]Login
	oauthID     string
	oauthSecret string
}

func NewAuthCache(oauthID, oauthSecret string) *AuthCache {
	return &AuthCache{
		mu:          &sync.Mutex{},
		userLogins:  make(map[interface{}]Login),
		oauthID:     oauthID,
		oauthSecret: oauthSecret,
	}
}

func (ac *AuthCache) TryGet(key interface{}) (Login, bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	login, exists := ac.userLogins[key]
	return login, exists
}

func (ac *AuthCache) Add(key interface{}, oauthCode url.Values) (Login, error) {
	code, err := parseOAuthCode(oauthCode)
	if err != nil {
		return Login{}, err
	}
	token, err := code.getToken(ac.oauthID, ac.oauthSecret)
	if err != nil {
		return Login{}, err
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	cj, _ := cookiejar.New(nil)
	login := Login{
		token:   token,
		cookies: cj,
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
	// todo
}
