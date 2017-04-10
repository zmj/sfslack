package sfauth

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/zmj/sfslack/sharefile"
)

const (
	tokenRefreshCheck = 30 * time.Minute
)

type Cache struct {
	mu          *sync.Mutex
	userLogins  map[interface{}]*login
	oauthID     string
	oauthSecret string
}

func New(oauthID, oauthSecret string) *Cache {
	return &Cache{
		mu:          &sync.Mutex{},
		userLogins:  make(map[interface{}]*login),
		oauthID:     oauthID,
		oauthSecret: oauthSecret,
	}
}

func (c *Cache) TryGet(key interface{}) (sharefile.Credentials, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	login, exists := c.userLogins[key]
	return login, exists
}

func (c *Cache) Add(key interface{}, oauthCode url.Values) (sharefile.Credentials, error) {
	code, err := parseOAuthCode(oauthCode)
	if err != nil {
		return nil, err
	}
	token, err := code.getToken(c.oauthID, c.oauthSecret)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v login %v\n", time.Now().Format(time.Stamp), key)
	c.mu.Lock()
	defer c.mu.Unlock()
	login := &login{
		token:   token,
		account: token.account,
		client:  &http.Client{},
	}
	c.userLogins[key] = login
	go c.refreshLoop(key)
	return login, nil
}

func (c *Cache) LoginURL(callbackURL string) string {
	return fmt.Sprintf("https://secure.sharefile.com/oauth/authorize?response_type=code&client_id=%v&redirect_uri=%v",
		c.oauthID,
		url.QueryEscape(callbackURL))
}

func (c *Cache) refreshLoop(key interface{}) {
	t := time.NewTicker(tokenRefreshCheck)
	defer t.Stop()
	for expired := false; !expired; {
		select {
		case <-t.C:
			cred, _ := c.TryGet(key)
			login := cred.(*login)
			if login.token.expiresAt.Before(time.Now()) {
				expired = true
				continue
			}
			if login.token.expiresAt.After(time.Now().Add(2 * time.Hour)) {
				continue
			}
			token, err := login.token.refresh(c.oauthID, c.oauthSecret)
			if err != nil {
				continue
			}
			login.token = token
			login.client.Jar = nil // clear authid
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.userLogins, key)
}
