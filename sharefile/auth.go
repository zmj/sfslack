package sharefile

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type Account struct {
	Subdomain       string `json:"subdomain,omitempty"`
	AppControlPlane string `json:"appcp,omitempty"`
	APIControlPlane string `json:"apicp,omitempty"`
}

type oauthToken struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	Account
	ExpiresAt time.Time `json:"-"`
}

type oauthCode struct {
	Code string
	Account
}

type Login struct {
	oauthToken
	client *http.Client
}

func (login Login) withCredentials(req *http.Request) *http.Request {
	if login.client.Jar == nil {
		jar, _ := cookiejar.New(nil)
		login.client.Jar = jar
	}
	url, _ := url.Parse(fmt.Sprintf("https://%v.%v", login.Subdomain, login.APIControlPlane))
	cookies := login.client.Jar.Cookies(url)
	if len(cookies) == 0 {
		req.Header.Add("Authorization", "Bearer "+login.oauthToken.AccessToken)
	}
	return req
}

func parseOAuthCode(values url.Values) (oauthCode, error) {
	code := oauthCode{
		Account: Account{
			Subdomain:       values.Get("subdomain"),
			AppControlPlane: values.Get("appcp"),
			APIControlPlane: values.Get("apicp"),
		},
		Code: values.Get("code"),
	}
	// validate
	return code, nil
}

func (code oauthCode) getToken(oauthID, oauthSecret string) (oauthToken, error) {
	values := map[string]string{
		"client_id":     oauthID,
		"client_secret": oauthSecret,
		"code":          code.Code,
		"grant_type":    "authorization_code",
	}
	return code.tokenPost(values)
}

func (token oauthToken) refresh(oauthID, oauthSecret string) (oauthToken, error) {
	values := map[string]string{
		"client_id":     oauthID,
		"client_secret": oauthSecret,
		"refresh_token": token.RefreshToken,
		"grant_type":    "refresh_token",
	}
	return token.tokenPost(values)
}

func (sf Account) tokenPost(values map[string]string) (oauthToken, error) {
	var valuePairs []string
	for k, v := range values {
		valuePairs = append(valuePairs, fmt.Sprintf("%v=%v", k, v))
	}
	toSend := strings.Join(valuePairs, "&")
	req, err := http.NewRequest("POST",
		fmt.Sprintf("https://%v.%v/oauth/token?requirev3=true", sf.Subdomain, sf.AppControlPlane),
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
	token = token.withExpiresAt()
	return token, nil
}

func (token oauthToken) withExpiresAt() oauthToken {
	d := time.Duration(token.ExpiresIn) * time.Second
	token.ExpiresAt = time.Now().Add(d)
	return token
}
