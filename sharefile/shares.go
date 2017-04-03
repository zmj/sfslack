package sharefile

import "context"

func (sf Login) CreateRequestShare(ctx context.Context, parentFolderId string) (Share, error) {
	toCreate := Share{ShareType: "Request",
		Parent: Folder{Item: Item{URL: sf.Account().itemURL("Items", parentFolderId)}}}
	return sf.CreateShare(ctx, toCreate)
}

func (sf Login) CreateSendShare(ctx context.Context, files []File) (Share, error) {
	toCreate := Share{ShareType: "Send"}
	for _, file := range files {
		file := File{Item{URL: sf.Account().itemURL("Items", file.ID)}}
		toCreate.Items = append(toCreate.Items, file)
	}
	return sf.CreateShare(ctx, toCreate)
}

func (sf Login) UpdateSendShare(ctx context.Context, share Share, files []File) (Share, error) {
	toUpdate := Share{
		URL:   share.URL,
		Items: files,
	}
	return sf.UpdateShare(ctx, toUpdate)
}

func (sf Login) CreateShare(ctx context.Context, toCreate Share) (Share, error) {
	result := Share{}
	err := sf.doPost(ctx, sf.Account().entityURL("Shares"), toCreate, &result)
	return result, err
}

func (sf Login) UpdateShare(ctx context.Context, toUpdate Share) (Share, error) {
	result := Share{}
	err := sf.doPatch(ctx, sf.Account().entityURL("Shares"), toUpdate, &result)
	return result, err
}

func (sf Account) DownloadAllURL(sh Share) string {
	return sf.itemURL("Shares", sh.ID) + "/Download"
}

func (sf Account) DownloadURL(sh Share, fileID string) string {
	return sf.itemURL("Shares", sh.ID) + "/Download(" + fileID + ")"
}
