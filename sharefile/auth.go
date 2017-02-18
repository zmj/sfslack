package sharefile

import (
	"context"
	"errors"
	"fmt"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"sync"
)

type AuthCache struct {
	mutex         sync.Mutex
	oauthID       string
	oauthSecret   string
	prevUserID    int
	prevRequestID int
	userIDsByKey  map[interface{}]int
	usersByID     map[int]*userAuth
}

func NewAuthCache(oauthID, oauthSecret string) *AuthCache {
	return &AuthCache{
		mutex:        sync.Mutex{},
		oauthID:      oauthID,
		oauthSecret:  oauthSecret,
		userIDsByKey: make(map[interface{}]int),
		usersByID:    make(map[int]*userAuth),
	}
}

type authStep int

const (
	stepNotStarted authStep = iota
	stepWaitingForCallback
	stepGotCallback
	stepLoggedIn
	userIDQueryKey    = "uid"
	requestIDQueryKey = "rid"
)

type userAuth struct {
	ID      int
	token   OAuthToken
	login   Login
	step    authStep
	waiting []*AuthRequest
}

type AuthRequest struct {
	ID              int
	StartURL        string
	Login           Login
	Err             error
	Done            chan struct{}
	RedirectBrowser chan string
}

func (ac *AuthCache) Get(ctx context.Context, key interface{}, callbackURL string) *AuthRequest {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	user := ac.getUser(key)
	ac.prevRequestID++
	var req = &AuthRequest{ID: ac.prevRequestID}
	switch user.step {
	case stepLoggedIn:
		req.Login = user.login
		req.Done = make(chan struct{})
		close(req.Done)
		return req
	case stepNotStarted, stepWaitingForCallback:
		req.StartURL = ac.loginURL(callbackURL, user.ID, req.ID)
		user.step = stepWaitingForCallback
	case stepGotCallback:
	}
	req.Done = make(chan struct{})
	user.waiting = append(user.waiting, req)
	go ac.cleanupRequest(ctx, user.ID, req.ID)
	return req
}

func (ac *AuthCache) cleanupRequest(ctx context.Context, userID, requestID int) {
	<-ctx.Done()
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	user, ok := ac.usersByID[userID]
	if !ok {
		return
	}
	for i, req := range user.waiting {
		if req.ID == requestID {
			req.Err = ctx.Err()
			close(req.Done)
			user.waiting = append(user.waiting[:i], user.waiting[i+1:]...)
			return
		}
	}
}

func (ac *AuthCache) Callback(ctx context.Context, values url.Values) (string, error) {
	cb, err := parseAuthCallback(values)
	if err != nil {
		return "", err
	}
	token, loginErr := cb.GetToken(ctx)
	if loginErr != nil {
		return "", err
	}
	cookies, err := cookiejar.New(nil)
	if err != nil {
		return "", err
	}
	login := Login{
		Token:   token,
		Account: token.Account,
		Cookies: cookies,
	}
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	user, ok := ac.usersByID[cb.UserID]
	if !ok {
		return "", errors.New("Unexpected authentication")
	}
	var redirect chan string
	for _, req := range user.waiting {
		req.Login = login
		if loginErr != nil {
			req.Err = loginErr
		}
		if req.ID == cb.RequestID {
			redirect = make(chan string)
			req.RedirectBrowser = redirect
		}
		close(req.Done)
	}
	if redirect != nil {
		select {
		case url := <-redirect:
			return url, nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return "", nil
}

func (ac *AuthCache) getUser(key interface{}) *userAuth {
	var user *userAuth
	id, ok := ac.userIDsByKey[key]
	if !ok {
		ac.prevUserID++
		user = &userAuth{ID: ac.prevUserID}
		ac.userIDsByKey[key] = user.ID
		ac.usersByID[user.ID] = user
	} else {
		user = ac.usersByID[id]
	}
	return user
}

func (ac *AuthCache) loginURL(callbackURL string, userID, requestID int) string {
	return fmt.Sprintf("https://secure.sharefile.com/oauth/authorize?response_type=code&client_id=%v&redirect_uri=%v",
		ac.oauthID,
		url.QueryEscape(ac.callbackURL(callbackURL, userID, requestID)))
}

func (ac *AuthCache) callbackURL(callbackURL string, userID, requestID int) string {
	return fmt.Sprintf("%v?%v=%v&%v=%v",
		callbackURL,
		userIDQueryKey, userID,
		requestIDQueryKey,
		requestID)
}

type authCallback struct {
	OAuthCode
	UserID    int
	RequestID int
}

func parseAuthCallback(values url.Values) (authCallback, error) {
	code, err := parseOAuthCode(values)
	if err != nil {
		return authCallback{}, err
	}
	userID, err := strconv.Atoi(values.Get(userIDQueryKey))
	if err != nil {
		return authCallback{}, err
	}
	requestID, err := strconv.Atoi(values.Get(requestIDQueryKey))
	if err != nil {
		return authCallback{}, err
	}
	return authCallback{
		OAuthCode: code,
		UserID:    userID,
		RequestID: requestID,
	}, nil
}

func parseOAuthCode(values url.Values) (OAuthCode, error) {
	code := OAuthCode{
		Account: Account{
			Subdomain:       values.Get("subdomain"),
			AppControlPlane: values.Get("appcp"),
			ApiControlPlane: values.Get("apicp"),
		},
		Code: values.Get("code"),
	}
	// validate
	return code, nil
}
