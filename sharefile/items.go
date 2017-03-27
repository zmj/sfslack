package sharefile

import (
	"context"
	"errors"
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

func (sf *Login) CreateFolder(ctx context.Context, name, parentFolderId string) (Folder, error) {
	toCreate := Folder{Name: name}
	url := sf.Account().itemURL("Items", parentFolderId) + "/Folder"
	result := Folder{}
	return result, sf.doPost(ctx, url, toCreate, &result)
}

func (sf *Login) GetChildren(ctx context.Context, parentFolderId string) ([]Item, error) {
	url := sf.Account().itemURL("Items", parentFolderId) + "/Children"

	result := Items{}
	return result.Items, sf.doGet(ctx, url, &result)
}
