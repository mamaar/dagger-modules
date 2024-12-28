// A set of utilities for working with AWS services.
//
// Retrieve AWS credentials for a specific profile. Helpful for authenticating with AWS when there are multiple SSO profiles.
// Get ECR credentials for authenticating of simply push container images to ECR.

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"dagger/aws-utils/internal/dagger"
	"dagger/aws-utils/pkg"
)

const (
	GoImage     = "golang:1.23"
	CacheVolume = "aws"
)

type Credentials struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
	Region          string `json:"region"`
}

type EcrToken struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Endpoint string `json:"endpoint"`
}

type Error struct {
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

type AwsUtils struct{}

func (m *AwsUtils) util(ctx context.Context, awsDir *dagger.Directory, awsProfile string, command []string) *dagger.Container {
	cacheVolume := dag.CacheVolume(CacheVolume)
	return dag.Container().
		From(GoImage).
		WithMountedCache("/go/pkg", cacheVolume).
		WithDirectory("/root/.aws", awsDir).
		WithEnvVariable("AWS_PROFILE", awsProfile).
		WithDirectory("/app", dag.CurrentModule().Source()).
		WithWorkdir("/app/dagger").
		WithExec(append([]string{"go", "run", "dagger/aws-utils/cmd"}, command...))
}

// RetrieveCredentials retrieves AWS credentials for the given profile
func (m *AwsUtils) RetrieveCredentials(ctx context.Context, awsDir *dagger.Directory, awsProfile string) (Credentials, error) {
	out, err := m.util(ctx, awsDir, awsProfile, []string{pkg.CommandRetrieveCredentials}).Stdout(ctx)
	if err != nil {
		unErr := Error{}
		_ = json.Unmarshal([]byte(out), &unErr)
		return Credentials{}, unErr
	}
	cred := Credentials{}
	_ = json.Unmarshal([]byte(out), &cred)
	return cred, nil
}

// GetEcrToken retrieves an ECR token for the given profile. The token consists of username, password and endpoint.
func (m *AwsUtils) GetEcrToken(ctx context.Context, awsDir *dagger.Directory, awsProfile string) (EcrToken, error) {
	out, err := m.util(ctx, awsDir, awsProfile, []string{pkg.CommandEcrGetToken}).Stdout(ctx)
	if err != nil {
		unErr := Error{}
		_ = json.Unmarshal([]byte(out), &unErr)
		return EcrToken{}, unErr
	}
	tok := EcrToken{}
	_ = json.Unmarshal([]byte(out), &tok)
	return tok, nil
}

// PushToEcr pushes a container image to ECR. It returns a list of references for the pushed images.
func (m *AwsUtils) PushToEcr(ctx context.Context, container *dagger.Container, awsDir *dagger.Directory, awsProfile string, imageName string, tags []string) ([]string, error) {
	token, err := m.GetEcrToken(ctx, awsDir, awsProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to get ECR token: %w", err)
	}

	secret := dag.SetSecret("ecr-token", token.Password)

	var refs []string
	for _, tag := range tags {
		uri := fmt.Sprintf("%s/%s:%s", token.Endpoint, imageName, tag)
		ref, err := container.
			WithRegistryAuth(token.Endpoint, token.Username, secret).
			Publish(ctx, uri)
		if err != nil {
			return nil, fmt.Errorf("failed to push image: %w", err)
		}
		refs = append(refs, ref)
	}
	return refs, nil
}
