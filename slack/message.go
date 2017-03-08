package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Message struct {
	Text         string       `json:"text"`
	ResponseType string       `json:"response_type,omitempty"`
	Attachments  []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Fallback string `json:"fallback,omitempty"`
	Text     string `json:"text,omitempty"`
}

func (msg Message) WriteTo(wr io.Writer) (int64, error) {
	toSend, err := json.Marshal(msg)
	if err != nil {
		return 0, err
	}
	written, err := wr.Write(toSend)
	return int64(written), err
}

func (msg Message) RespondTo(cmd Command) error {
	toSend, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST",
		cmd.ResponseURL,
		bytes.NewReader(toSend))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

func FormatURL(url, display string) string {
	return fmt.Sprintf("<%v|%v>", url, display)
}
