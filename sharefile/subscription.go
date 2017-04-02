package sharefile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

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

type WebhookSubscriptionEvent struct {
	WebhookSubscriptionID string            `json:"WebhookSubscriptionId,omitempty"`
	AccountID             string            `json:"AccountId,omitempty"`
	Event                 SubscriptionEvent `json:",omitempty"`
}

type SubscriptionEvent struct {
	Timestamp     time.Time          `json:",omitempty"`
	OperationName string             `json:",omitempty"`
	ResourceType  string             `json:",omitempty"`
	Resource      SubscribedResource `json:",omitempty"`
}

type SubscribedResource struct {
	ID     string              `json:"Id,omitempty"`
	Parent *SubscribedResource `json:",omitempty"`
}

func (sf *Login) CreateSubscription(ctx context.Context, toCreate WebhookSubscription) (WebhookSubscription, error) {
	result := WebhookSubscription{}
	err := sf.doPost(ctx, sf.Account().entityURL("WebhookSubscriptions"), toCreate, &result)
	return result, err
}

func (sf *Login) DeleteSubscription(ctx context.Context, id string) error {
	url := sf.Account().itemURL("WebhookSubscriptions", id)
	return sf.doPost(ctx, url, nil, nil)
}

func (sf *Login) Subscribe(ctx context.Context, folder Folder, callbackURL string, eventTypes ...string) (WebhookSubscription, error) {
	toCreate := WebhookSubscription{
		SubscriptionContext: SubscriptionContext{
			ResourceType: ResourceTypeFolder,
			ResourceId:   folder.ID,
		},
		WebhookURL: callbackURL,
	}
	for _, et := range eventTypes {
		toCreate.Events = append(toCreate.Events, SubscribedResourceEvent{
			ResourceType:  ResourceTypeFile,
			OperationName: et,
		})
		toCreate.Events = append(toCreate.Events, SubscribedResourceEvent{
			ResourceType:  ResourceTypeFolder,
			OperationName: et,
		})
	}
	return sf.CreateSubscription(ctx, toCreate)
}

func ParseEvent(rdr io.Reader) (WebhookSubscriptionEvent, error) {
	event := WebhookSubscriptionEvent{}
	err := json.NewDecoder(rdr).Decode(&event)
	if err != nil {
		err = fmt.Errorf("Event parsing failed: %v", err)
	}
	return event, err
}
