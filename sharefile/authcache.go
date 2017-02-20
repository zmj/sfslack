package sharefile

import (
	"net/http"
)

type AuthCache struct {
}

func NewAuthCache() *AuthCache {
	return &AuthCache{}
}

type Login struct {
}

type AuthCallback struct {
}

func (ac *AuthCache) TryGet(key interface{}) (Login, bool) {

}

func (ac *AuthCache) Add(key interface{}, cb AuthCallback) err {

}

func ParseCallback(req *http.Request) (AuthCallback, err) {

}

func (ac *AuthCache) LoginURL(wfID int) string {

}
