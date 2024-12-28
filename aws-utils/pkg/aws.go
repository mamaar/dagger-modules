package pkg

import (
	"encoding/json"
)

const (
	CommandRetrieveCredentials = "retrieve-credentials"
	CommandEcrGetToken         = "ecr-get-token"
)

type Credentials struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
	Region          string `json:"region"`
}

type EcrToken struct {
	Token         string `json:"token"`
	ProxyEndpoint string `json:"proxy_endpoint"`
}

type Error struct {
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

func JSONError(err error) []byte {
	b, _ := json.Marshal(Error{Message: err.Error()})
	return b
}
