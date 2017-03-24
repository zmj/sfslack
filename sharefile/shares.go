package sharefile

func (sf Login) CreateRequestShare(parentFolderId string) (Share, error) {
	toCreate := Share{ShareType: "Request",
		Parent: Folder{Item: Item{URL: sf.itemURL("Items", parentFolderId)}}}
	return sf.CreateShare(toCreate)
}

func (sf Login) CreateSendShare(files []File) (Share, error) {
	toCreate := Share{ShareType: "Send"}
	for _, file := range files {
		toCreate.Items = append(toCreate.Items, File{Item{URL: sf.itemURL("Items", file.ID)}})
	}
	return sf.CreateShare(toCreate)
}

func (sf Login) CreateShare(toCreate Share) (Share, error) {
	result := Share{}
	err := sf.doPost(sf.entityURL("Shares"), toCreate, &result)
	return result, err
}

func (sf Account) DownloadAllURL(sh Share) string {
	return sf.itemURL("Shares", sh.ID) + "/Download"
}

func (sf Account) DownloadURL(sh Share, fileID string) string {
	return sf.itemURL("Shares", sh.ID) + "/Download(" + fileID + ")"
}
