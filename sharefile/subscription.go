package sharefile

import "context"

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

func (sf *Login) CreateSubscription(ctx context.Context, toCreate WebhookSubscription) (WebhookSubscription, error) {
	result := WebhookSubscription{}
	err := sf.doPost(ctx, sf.entityURL("WebhookSubscriptions"), toCreate, &result)
	return result, err
}

func (sf *Login) DeleteSubscription(ctx context.Context, id string) error {
	url := sf.itemURL("WebhookSubscriptions", id)
	return sf.doPost(ctx, url, nil, nil)
}
