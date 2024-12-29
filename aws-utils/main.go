// A set of utilities for working with AWS services.
//
// Retrieve AWS credentials for a specific profile. Helpful for authenticating with AWS when there are multiple SSO profiles.
// Get ECR credentials for authenticating of simply push container images to ECR.

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/lambda"

	"dagger/aws-utils/internal/dagger"
)

const (
	SecretAccessKeyID     = "aws-utils:access-key-id"
	SecretSecretAccessKey = "aws-utils:secret-access-key"
	SecretSessionToken    = "aws-utils:session-token"
	SecretEcrPassword     = "aws-utils:ecr-password"
)

type Credentials struct {
	AccessKeyID     *dagger.Secret `json:"access_key_id"`
	SecretAccessKey *dagger.Secret `json:"secret_access_key"`
	SessionToken    *dagger.Secret `json:"session_token"`
	Region          string         `json:"region"`
}

type EcrToken struct {
	Username string         `json:"username"`
	Password *dagger.Secret `json:"password"`
	Endpoint string         `json:"endpoint"`
}
type AwsUtils struct{}

func (m *AwsUtils) setupConfig(ctx context.Context, awsDir *dagger.Directory, awsProfile string) (aws.Config, error) {
	_, err := awsDir.Export(ctx, "/root/.aws")
	if err != nil {
		return aws.Config{}, err
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(awsProfile),
	)
	if err != nil {
		return aws.Config{}, err
	}

	return cfg, nil
}

// RetrieveCredentials retrieves AWS credentials for the given profile
func (m *AwsUtils) RetrieveCredentials(ctx context.Context, awsDir *dagger.Directory, awsProfile string) (Credentials, error) {
	cfg, err := m.setupConfig(ctx, awsDir, awsProfile)
	if err != nil {
		return Credentials{}, err
	}

	// Retrieve credentials
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return Credentials{}, err
	}

	accessKeyIDSecret := dag.SetSecret(SecretAccessKeyID, creds.AccessKeyID)
	secretAccessKeySecret := dag.SetSecret(SecretSecretAccessKey, creds.SecretAccessKey)
	sessionTokenSecret := dag.SetSecret(SecretSessionToken, creds.SessionToken)

	tokens := Credentials{
		AccessKeyID:     accessKeyIDSecret,
		SecretAccessKey: secretAccessKeySecret,
		SessionToken:    sessionTokenSecret,
		Region:          cfg.Region,
	}

	err = json.NewEncoder(os.Stdout).Encode(tokens)
	if err != nil {
		return Credentials{}, err
	}

	return tokens, nil
}

// GetEcrToken retrieves an ECR token for the given profile. The token consists of username, password and endpoint.
func (m *AwsUtils) GetEcrToken(ctx context.Context, awsDir *dagger.Directory, awsProfile string) (EcrToken, error) {
	cfg, err := m.setupConfig(ctx, awsDir, awsProfile)
	if err != nil {
		return EcrToken{}, err
	}

	svc := ecr.NewFromConfig(cfg)

	// Get the tokenResponse
	tokenResponse, err := svc.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return EcrToken{}, err
	}

	if len(tokenResponse.AuthorizationData) == 0 {
		return EcrToken{}, fmt.Errorf("no authorization data found")
	}

	token := *tokenResponse.AuthorizationData[0].AuthorizationToken
	proxyEndpoint := *tokenResponse.AuthorizationData[0].ProxyEndpoint

	tokenDecoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return EcrToken{}, fmt.Errorf("failed to decode ECR tokenResponse: %w", err)
	}

	username, password, _ := strings.Cut(string(tokenDecoded), ":")
	passwordSecret := dag.SetSecret(SecretEcrPassword, password)

	endpointUrl, err := url.Parse(proxyEndpoint)
	if err != nil {
		return EcrToken{}, fmt.Errorf("failed to parse ECR endpoint: %w", err)
	}

	return EcrToken{
		Username: username,
		Password: passwordSecret,
		Endpoint: endpointUrl.Host,
	}, nil
}

// PushToEcr pushes a container image to ECR. It returns a list of references for the pushed images.
func (m *AwsUtils) PushToEcr(ctx context.Context, container *dagger.Container, awsDir *dagger.Directory, awsProfile string, imageName string, tags []string) ([]string, error) {
	token, err := m.GetEcrToken(ctx, awsDir, awsProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to get ECR token: %w", err)
	}

	var refs []string
	for _, tag := range tags {
		uri := fmt.Sprintf("%s/%s:%s", token.Endpoint, imageName, tag)
		ref, err := container.
			WithRegistryAuth(token.Endpoint, token.Username, token.Password).
			Publish(ctx, uri)
		if err != nil {
			return nil, fmt.Errorf("failed to push image: %w", err)
		}
		ref, _, _ = strings.Cut(ref, "@")
		refs = append(refs, ref)
	}
	return refs, nil
}

// UpdateLambdaImage updates the image of a Lambda function.
func (m *AwsUtils) UpdateLambdaImage(ctx context.Context, awsDir *dagger.Directory, awsProfile string, functionName string, imageRef string) error {
	cfg, err := m.setupConfig(ctx, awsDir, awsProfile)
	if err != nil {
		return err
	}
	svc := lambda.NewFromConfig(cfg)

	_, err = svc.UpdateFunctionCode(ctx, &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(functionName),
		ImageUri:     aws.String(imageRef),
	})
	if err != nil {
		return err
	}

	return nil
}
