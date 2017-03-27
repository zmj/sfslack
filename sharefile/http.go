package sharefile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Login struct {
	Credentials
}

type Credentials interface {
	Account() Account
	Do(*http.Request) (*http.Response, error)
}

func (login Login) doPost(ctx context.Context, url string, send, recv interface{}) error {
	var body io.Reader
	if send != nil {
		b, err := json.Marshal(send)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	return login.do(ctx, req, recv)
}

func (login *Login) doGet(ctx context.Context, url string, recv interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	return login.do(ctx, req, recv)
}

func (login *Login) do(ctx context.Context, req *http.Request, recv interface{}) error {
	resp, err := login.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	if recv != nil {
		err = json.NewDecoder(resp.Body).Decode(recv)
		if err != nil {
			return err
		}
	}
	return nil
}
