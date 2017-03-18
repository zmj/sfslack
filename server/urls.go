package server

import "fmt"

const (
	commandPath      = "/sfslack/command"
	commandClickPath = "/sfslack/command/click"
	authPath         = "/sfslack/auth"
	eventPath        = "/sfslack/event"

	wfidQueryKey   = "wfid"
	wfTypeQueryKey = "wftype"
)

func commandClickURL(host string, wfID int) string {
	return fmt.Sprintf("https://%v%v?%v=%v",
		host,
		commandClickPath,
		wfidQueryKey,
		wfID)
}

func authCallbackURL(host string, wfID int) string {
	return fmt.Sprintf("https://%v%v?%v=%v",
		host,
		authPath,
		wfidQueryKey,
		wfID)
}

func eventCallbackURL(host string, wfID int) string {
	return fmt.Sprintf("https://%v%v?%v=%v",
		host,
		eventPath,
		wfidQueryKey,
		wfID)
}
