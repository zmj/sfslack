package main

import "net/http"

type SfAccount struct {
	Subdomain       string
	AppControlPlane string
	ApiControlPlane string
}

type SfLogin struct {
	SfAccount
	Cookies http.CookieJar
}
