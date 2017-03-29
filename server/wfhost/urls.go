package wfhost

type CallbackURLs interface {
	CommandClick(int) string
	AuthCallback(int) string
	EventWebhook(int) string
	Waiting(int) string
}
