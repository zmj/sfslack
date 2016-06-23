package main

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

func TestLogin() SfLogin {
	account := SfAccount{"jeffcombscom", "sharefile.com", "sf-api.com"}
	authCookie := http.Cookie{
		Name:  "SFAPI_AuthID",
		Value: "a7622b87-3fff-4caf-97dd-dd7ddb78057d"}

	cookieUrl, _ := url.Parse(account.BaseUrl())
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(cookieUrl, []*http.Cookie{&authCookie})
	return SfLogin{account, jar}
}

func AuthExists(slack SlackUser) (SfLogin, bool) {
	return TestLogin(), true
}
