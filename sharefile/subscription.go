package sharefile

const (
	ResourceTypeFile    = "File"
	ResourceTypeFolder  = "Folder"
	OperationNameUpload = "Upload"
)

type WebhookSubscription struct {
	ID                  string                    `json:"Id,omitempty"`
	URL                 string                    `json:"url,omitempty"`
	SubscriptionContext SubscriptionContext       `json:",omitempty"`
	Events              []SubscribedResourceEvent `json:",omitempty"`
	WebhookURL          string                    `json:"WebhookUrl,omitempty"`
}

type SubscriptionContext struct {
	ResourceType string `json:",omitempty"`
	ResourceId   string `json:",omitempty"`
}

type SubscribedResourceEvent struct {
	ResourceType  string `json:",omitempty"`
	OperationName string `json:",omitempty"`
}

func (sf *Login) CreateSubscription(toCreate WebhookSubscription) (WebhookSubscription, error) {
	result := WebhookSubscription{}
	err := sf.doPost(sf.entityURL("WebhookSubscriptions"), toCreate, &result)
	return result, err
}

func (sf *Login) DeleteSubscription(id string) error {
	url := sf.itemURL("WebhookSubscriptions", id)
	return sf.doPost(url, nil, nil)
}
