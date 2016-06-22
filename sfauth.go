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
		Value: "421e60ff-7721-4002-a492-3060a3c594a4"}

	cookieUrl, _ := url.Parse(account.BaseUrl())
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(cookieUrl, []*http.Cookie{&authCookie})
	return SfLogin{account, jar}
}

func AuthExists(slack SlackUser) (SfLogin, bool) {
	return TestLogin(), true
}
