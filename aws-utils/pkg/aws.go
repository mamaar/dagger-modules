package pkg

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
