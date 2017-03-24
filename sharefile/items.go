package sharefile

import (
	"encoding/json"
	"errors"
	"net/http"
)

func (item Item) File() (File, error) {
	if item.ID[0:2] != "fi" {
		return File{}, errors.New("Not a file")
	}
	return File{item}, nil
}

func (item Item) Folder() (Folder, error) {
	if item.ID[0:2] != "fo" {
		return Folder{}, errors.New("Not a folder")
	}
	return Folder{Item: item}, nil
}

func (sf Login) CreateFolder(name, parentFolderId string) (Folder, error) {
	toCreate := Folder{Name: name}
	url := sf.itemURL("Items", parentFolderId) + "/Folder"
	result := Folder{}
	return result, sf.doPost(url, toCreate, &result)
}

func (sf Login) GetChildren(parentFolderId string) ([]Item, error) {
	req, err := http.NewRequest("GET",
		sf.itemURL("Items", parentFolderId)+"/Children",
		nil)
	if err != nil {
		return nil, err
	}
	req = sf.withCredentials(req)

	resp, err := sf.client.Do(req)
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
