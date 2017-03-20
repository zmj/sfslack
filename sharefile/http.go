package sharefile

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

func (sf *Login) doPost(entity string, body, response interface{}) error {
	toSend, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST",
		sf.entityURL(entity),
		bytes.NewReader(toSend))
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

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}
	// need this for building urls?
	// created.Account = sf.Account

	return nil
}
