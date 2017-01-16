package sharefile

import (
	"context"
	"fmt"
	"net/url"
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

type authStep int

const (
	stepNotStarted authStep = iota
	stepWaitingForCallback
	stepGotCallback
	stepLoggedIn
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
	RedirectBrowser func(string) bool
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
	return fmt.Sprintf("%v?uid=%v&rid=%v", callbackURL, userID, requestID)
}
