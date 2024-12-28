package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"

	"dagger/aws-utils/pkg"
)

func main() {
	ctx := context.Background()

	profile, hasProfile := os.LookupEnv("AWS_PROFILE")
	if !hasProfile {
		_, _ = os.Stdout.Write(pkg.JSONError(fmt.Errorf("AWS profile is not set")))
		os.Exit(1)
	}

	if len(os.Args) == 1 {
		_, _ = os.Stdout.Write(pkg.JSONError(fmt.Errorf("no command provided")))
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case pkg.CommandRetrieveCredentials:
		err = retrieveCredentials(profile)
		if err != nil {
			_, _ = os.Stdout.Write(pkg.JSONError(err))
			os.Exit(1)
		}
	case pkg.CommandEcrGetToken:
		err = getECRToken(ctx, profile)
		if err != nil {
			_, _ = os.Stdout.Write(pkg.JSONError(err))
			os.Exit(1)
		}
	default:
		_, _ = os.Stdout.Write(pkg.JSONError(fmt.Errorf("unknown command: %s", os.Args[1])))
		os.Exit(1)
	}
}

func retrieveCredentials(profile string) error {

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(profile), // Replace with your SSO profile name
	)
	if err != nil {
		return err
	}

	// Retrieve credentials
	creds, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		return err
	}

	tokens := pkg.Credentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Region:          cfg.Region,
	}

	err = json.NewEncoder(os.Stdout).Encode(tokens)
	if err != nil {
		return err
	}

	return nil
}

func getECRToken(ctx context.Context, profile string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(profile), // Replace with your SSO profile name
	)
	if err != nil {
		return err
	}

	svc := ecr.NewFromConfig(cfg)

	// Get the token
	token, err := svc.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return err
	}

	if len(token.AuthorizationData) == 0 {
		return fmt.Errorf("no authorization data found")
	}

	output := pkg.EcrToken{
		Token:         *token.AuthorizationData[0].AuthorizationToken,
		ProxyEndpoint: *token.AuthorizationData[0].ProxyEndpoint,
	}
	err = json.NewEncoder(os.Stdout).Encode(output)
	if err != nil {
		return err
	}

	return nil
}
