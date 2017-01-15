package sharefile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const (
	OAuthId     = "n9UV16i2BeawbSsR8426cjHezF3cwX7o"
	OAuthSecret = "gbPfE206XMZvkfFU26WJhJQMI3wW3itXVBaM0Fo0nv3lVhhH"
)

type OAuthCode struct {
	Code string
	Account
}

type OAuthToken struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	Account
	ExpiresAt time.Time `json:"-"`
}

type Account struct {
	Subdomain       string `json:"subdomain,omitempty"`
	AppControlPlane string `json:"appcp,omitempty"`
	ApiControlPlane string `json:"apicp,omitempty"`
}

type Login struct {
	Account
	Token   OAuthToken
	Cookies http.CookieJar
}

type Share struct {
	Account   Account `json:"-"`
	Id        string  `json:",omitempty"`
	Url       string  `json:"url,omitempty"`
	ShareType string  `json:",omitempty"`
	Parent    Folder  `json:",omitempty"`
	Items     []File  `json:",omitempty"`
	Uri       string  `json:",omitempty"`
}

type Item struct {
	Id       string `json:",omitempty"`
	Url      string `json:"url,omitempty"`
	FileName string `json:",omitempty"`
}

type File struct {
	Item
}

func (item Item) File() (File, error) {
	if item.Id[0:2] != "fi" {
		return File{}, errors.New("Not a file")
	}
	return File{item}, nil
}

type Folder struct {
	Item
	Name string `json:",omitempty"`
}

func (item Item) Folder() (Folder, error) {
	if item.Id[0:2] != "fo" {
		return Folder{}, errors.New("Not a folder")
	}
	return Folder{Item: item}, nil
}

type Items struct {
	Items []Item `json:"value,omitempty"`
}

func (sf Account) BaseUrl() string {
	return fmt.Sprintf("https://%v.%v/sf/v3", sf.Subdomain, sf.ApiControlPlane)
}

func (sf Account) EntityUrl(entity string) string {
	return fmt.Sprintf("%v/%v", sf.BaseUrl(), entity)
}

func (sf Account) ItemUrl(entity, id string) string {
	return fmt.Sprintf("%v(%v)", sf.EntityUrl(entity), id)
}

func (sf Login) CreateRequestShare(parentFolderId string) (Share, error) {
	toCreate := Share{ShareType: "Request",
		Parent: Folder{Item: Item{Url: sf.ItemUrl("Items", parentFolderId)}}}
	return sf.CreateShare(toCreate)
}

func (sf Login) CreateSendShare(files []File) (Share, error) {
	toCreate := Share{ShareType: "Send"}
	for _, file := range files {
		toCreate.Items = append(toCreate.Items, File{Item{Url: sf.ItemUrl("Items", file.Id)}})
	}
	return sf.CreateShare(toCreate)
}

func (sf Login) CreateShare(toCreate Share) (Share, error) {
	toSend, err := json.Marshal(toCreate)
	if err != nil {
		return Share{}, err
	}
	req, err := http.NewRequest("POST",
		sf.EntityUrl("Shares"),
		bytes.NewReader(toSend))
	if err != nil {
		return Share{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	sf.AddHeaders(req)

	hc := http.Client{Jar: sf.Cookies}
	resp, err := hc.Do(req)
	if err != nil {
		return Share{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Share{}, errors.New(resp.Status)
	}

	created := Share{}
	err = json.NewDecoder(resp.Body).Decode(&created)
	if err != nil {
		return Share{}, err
	}
	created.Account = sf.Account

	return created, nil
}

func (sf Login) CreateFolder(name, parentFolderId string) (Folder, error) {
	toCreate := Folder{Name: name}

	toSend, err := json.Marshal(toCreate)
	if err != nil {
		return Folder{}, err
	}

	req, err := http.NewRequest("POST",
		sf.ItemUrl("Items", parentFolderId)+"/Folder",
		bytes.NewReader(toSend))
	if err != nil {
		return Folder{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	sf.AddHeaders(req)

	hc := http.Client{Jar: sf.Cookies}
	resp, err := hc.Do(req)
	if err != nil {
		return Folder{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Folder{}, errors.New(resp.Status)
	}

	created := Folder{}
	err = json.NewDecoder(resp.Body).Decode(&created)
	if err != nil {
		return Folder{}, err
	}

	return created, nil
}

func (sf Login) GetChildren(parentFolderId string) ([]Item, error) {
	req, err := http.NewRequest("GET",
		sf.ItemUrl("Items", parentFolderId)+"/Children",
		nil)
	if err != nil {
		return nil, err
	}
	sf.AddHeaders(req)

	hc := http.Client{Jar: sf.Cookies}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	items := Items{}
	err = json.NewDecoder(resp.Body).Decode(&items)
	if err != nil {
		return nil, err
	}

	return items.Items, nil
}

func (sh Share) DownloadAllUrl() string {
	return sh.Account.ItemUrl("Shares", sh.Id) + "/Download"
}

func (sh Share) DownloadUrl(fileId string) string {
	return sh.Account.ItemUrl("Shares", sh.Id) + "/Download(" + fileId + ")"
}

func dbgReq(req *http.Request) {
	reqRaw, _ := httputil.DumpRequestOut(req, true)
	fmt.Println(string(reqRaw))
}
func dbgResp(resp *http.Response) {
	respRaw, _ := httputil.DumpResponse(resp, true)
	fmt.Println(string(respRaw))
}

func (sf Login) AddHeaders(req *http.Request) {
	url, _ := url.Parse(fmt.Sprintf("https://%v.%v", sf.Subdomain, sf.ApiControlPlane))
	cookies := sf.Cookies.Cookies(url)
	if len(cookies) == 0 {
		req.Header.Add("Authorization", "Bearer "+sf.Token.AccessToken)
	}
}
func (sf OAuthToken) ShouldRefresh() bool {
	return sf.ExpiresAt.Sub(time.Now()).Hours() < 2
}

func (sf OAuthToken) SetExpiresAt() {
	d := time.Duration(sf.ExpiresIn) * time.Second
	sf.ExpiresAt = time.Now().Add(d)
}

func (sf OAuthCode) GetToken() (OAuthToken, error) {
	values := map[string]string{
		"client_id":     OAuthId,
		"client_secret": OAuthSecret,
		"code":          sf.Code,
		"grant_type":    "authorization_code",
	}
	return sf.TokenPost(values)
}

func (sf OAuthToken) Refresh() (OAuthToken, error) {
	values := map[string]string{
		"client_id":     OAuthId,
		"client_secret": OAuthSecret,
		"refresh_token": sf.RefreshToken,
		"grant_type":    "refresh_token",
	}
	return sf.TokenPost(values)
}

func (sf Account) TokenPost(values map[string]string) (OAuthToken, error) {
	var valuePairs []string
	for k, v := range values {
		valuePairs = append(valuePairs, fmt.Sprintf("%v=%v", k, v))
	}
	toSend := strings.Join(valuePairs, "&")
	req, err := http.NewRequest("POST",
		fmt.Sprintf("https://%v.%v/oauth/token?requirev3=true", sf.Subdomain, sf.AppControlPlane),
		strings.NewReader(toSend))
	if err != nil {
		return OAuthToken{}, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return OAuthToken{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return OAuthToken{}, errors.New(resp.Status)
	}

	created := OAuthToken{}
	err = json.NewDecoder(resp.Body).Decode(&created)
	if err != nil {
		return OAuthToken{}, err
	}
	created.SetExpiresAt()

	return created, nil
}
