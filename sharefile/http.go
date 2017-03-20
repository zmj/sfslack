package sharefile

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (sf *Login) doPost(url string, send, recv interface{}) error {
	var body io.Reader
	if send != nil {
		b, err := json.Marshal(body)
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
	hc := &http.Client{}
	hc, req = sf.withCredentials(hc, req)

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
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
	// need this for building urls?
	// created.Account = sf.Account

	return nil
}
