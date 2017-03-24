package sharefile

import "fmt"

type Share struct {
	ID        string `json:"Id,omitempty"`
	URL       string `json:"url,omitempty"`
	ShareType string `json:",omitempty"`
	Parent    Folder `json:",omitempty"`
	Items     []File `json:",omitempty"`
	URI       string `json:"Uri,omitempty"`
}

type Item struct {
	ID       string `json:"Id,omitempty"`
	URL      string `json:"url,omitempty"`
	FileName string `json:",omitempty"`
}

type File struct {
	Item
}

type Folder struct {
	Item
	Name string `json:",omitempty"`
}

type Items struct {
	Items []Item `json:"value,omitempty"`
}

func (sf Account) baseURL() string {
	return fmt.Sprintf("https://%v.%v/sf/v3", sf.Subdomain, sf.APIControlPlane)
}

func (sf Account) entityURL(entity string) string {
	return fmt.Sprintf("%v/%v", sf.baseURL(), entity)
}

func (sf Account) itemURL(entity, id string) string {
	return fmt.Sprintf("%v(%v)", sf.entityURL(entity), id)
}
