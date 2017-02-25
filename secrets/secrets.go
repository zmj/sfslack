package secrets

import (
	"encoding/json"
	"io/ioutil"
)

const (
	secretsFile = "secrets.json"
)

type Secrets struct {
	OAuthID     string
	OAuthSecret string
	// slack api key
}

func Load() (Secrets, error) {
	bytes, err := ioutil.ReadFile(secretsFile)
	if err != nil {
		return Secrets{}, err
	}
	secrets := Secrets{}
	err = json.Unmarshal(bytes, &secrets)
	// validate
	return secrets, err
}
