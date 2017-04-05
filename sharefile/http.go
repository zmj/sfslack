package sharefile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	body, err := toBody(send)
	if err != nil {
		return fmt.Errorf("Failed to serialize body: %v", err)
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("Failed to create request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	return login.do(ctx, req, recv)
}

func (login Login) doPatch(ctx context.Context, url string, send, recv interface{}) error {
	body, err := toBody(send)
	if err != nil {
		return fmt.Errorf("Failed to serialize body: %v", err)
	}
	req, err := http.NewRequest("PATCH", url, body)
	if err != nil {
		return fmt.Errorf("Failed to create request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	return login.do(ctx, req, recv)
}

func toBody(send interface{}) (io.Reader, error) {
	var body io.Reader
	if send != nil {
		b, err := json.Marshal(send)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
	}
	return body, nil
}

func (login *Login) doGet(ctx context.Context, url string, recv interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	return login.do(ctx, req, recv)
}

func (login *Login) do(ctx context.Context, req *http.Request, recv interface{}) error {
	// log.dbg this through context
	// b, _ := httputil.DumpRequestOut(req, req.Method != "GET")
	// fmt.Println(string(b))

	resp, err := login.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// b, _ = httputil.DumpResponse(resp, true)
	// fmt.Println(string(b))

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
