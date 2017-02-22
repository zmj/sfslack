package sharefile

import (
	"fmt"
	"net/url"
	"sync"
)

type AuthCache struct {
	mu          *sync.Mutex
	userLogins  map[interface{}]*userLogin
	oauthID     string
	oauthSecret string
}

func NewAuthCache(oauthID, oauthSecret string) *AuthCache {
	return &AuthCache{
		mu:          &sync.Mutex{},
		userLogins:  make(map[interface{}]*userLogin),
		oauthID:     oauthID,
		oauthSecret: oauthSecret,
	}
}

type userLogin struct {
	login Login
	token oauthToken
}

func (ac *AuthCache) TryGet(key interface{}) (Login, bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	userLogin, exists := ac.userLogins[key]
	return userLogin.login, exists
}

func (ac *AuthCache) Add(key interface{}, oauthCode url.Values) error {
	code, err := parseOAuthCode(oauthCode)
	if err != nil {
		return err
	}
	token, err := code.getToken(ac.oauthID, ac.oauthSecret)
	if err != nil {
		return err
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	userLogin := &userLogin{
		token: token,
		login: Login{},
	}
	ac.userLogins[key] = userLogin
	go ac.refreshLoop(key)
	return nil
}

func (ac *AuthCache) LoginURL(callbackURL string) string {
	return fmt.Sprintf("https://secure.sharefile.com/oauth/authorize?response_type=code&client_id=%v&redirect_uri=%v",
		ac.oauthID,
		url.QueryEscape(callbackURL))
}

func (ac *AuthCache) refreshLoop(key interface{}) {
	// todo
}
