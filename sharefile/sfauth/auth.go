package sfauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/zmj/sfslack/sharefile"
)

type oauthToken struct {
	AccessToken  string            `json:"access_token,omitempty"`
	RefreshToken string            `json:"refresh_token,omitempty"`
	ExpiresIn    int               `json:"expires_in,omitempty"`
	account      sharefile.Account `json:"-"`
	expiresAt    time.Time         `json:"-"`
}

type oauthCode struct {
	code    string
	account sharefile.Account
}

type login struct {
	token   oauthToken
	account sharefile.Account
	client  *http.Client
}

func (login *login) withCredentials(req *http.Request) *http.Request {
	if login.client.Jar == nil {
		jar, _ := cookiejar.New(nil)
		login.client.Jar = jar
	}
	url, _ := url.Parse(fmt.Sprintf("https://%v.%v", login.account.Subdomain, login.account.APIControlPlane))
	cookies := login.client.Jar.Cookies(url)
	if len(cookies) == 0 { // && len(req.Header.Get("Authorization")) == 0 {
		fmt.Printf("%v Authorization\n", time.Now().Format(time.Stamp))
		req.Header.Add("Authorization", "Bearer "+login.token.AccessToken)
	}
	return req
}

func parseOAuthCode(values url.Values) (oauthCode, error) {
	code := oauthCode{
		account: sharefile.Account{
			Subdomain:       values.Get("subdomain"),
			AppControlPlane: values.Get("appcp"),
			APIControlPlane: values.Get("apicp"),
		},
		code: values.Get("code"),
	}
	// validate
	return code, nil
}

func (code oauthCode) getToken(oauthID, oauthSecret string) (oauthToken, error) {
	values := map[string]string{
		"client_id":     oauthID,
		"client_secret": oauthSecret,
		"code":          code.code,
		"grant_type":    "authorization_code",
	}
	return tokenPost(code.account, values)
}

func (token oauthToken) refresh(oauthID, oauthSecret string) (oauthToken, error) {
	values := map[string]string{
		"client_id":     oauthID,
		"client_secret": oauthSecret,
		"refresh_token": token.RefreshToken,
		"grant_type":    "refresh_token",
	}
	return tokenPost(token.account, values)
}

func tokenPost(acct sharefile.Account, values map[string]string) (oauthToken, error) {
	var valuePairs []string
	for k, v := range values {
		valuePairs = append(valuePairs, fmt.Sprintf("%v=%v", k, v))
	}
	toSend := strings.Join(valuePairs, "&")
	req, err := http.NewRequest("POST",
		fmt.Sprintf("https://%v.%v/oauth/token?requirev3=true", acct.Subdomain, acct.AppControlPlane),
		strings.NewReader(toSend))
	if err != nil {
		return oauthToken{}, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return oauthToken{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return oauthToken{}, errors.New(resp.Status)
	}

	token := oauthToken{}
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return oauthToken{}, err
	}
	token.account = acct
	token = token.withExpiresAt()
	return token, nil
}

func (token oauthToken) withExpiresAt() oauthToken {
	d := time.Duration(token.ExpiresIn) * time.Second
	token.expiresAt = time.Now().Add(d)
	return token
}

func (login *login) Account() sharefile.Account {
	return login.token.account
}

func (login *login) Do(req *http.Request) (*http.Response, error) {
	req = login.withCredentials(req)
	resp, err := login.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}
	resp.Body.Close()
	login.client.Jar = nil
	req = login.withCredentials(req)
	return login.client.Do(req)
}
