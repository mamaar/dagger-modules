# AWS Utils

This module provides a set of utilities to interact with AWS services.
Currently, it supports the following helpers:

## retrieve-credentials

Returns the AWS credentials for a given profile.

### Parameters
- `aws-dir` (required): The path to the AWS config directory. Example: `~/.aws`
- `aws-profile` (required): The profile to retrieve the credentials from. Example: `default`


## get-ecr-token

Returns a ECR token for a given profile.
The token consists of the `username` and `password` and `endpoint` to authenticate with the ECR registry.

### Parameters
- `aws-dir` (required): The path to the AWS config directory. Example: `~/.aws`
- `aws-profile` (required): The profile to retrieve the credentials from. Example: `default`


## push-to-ecr

Takes a `dagger.Container` and pushes the image to the ECR registry.

### Parameters
- `aws-dir` (required): The path to the AWS config directory. Example: `~/.aws`
- `aws-profile` (required): The profile to retrieve the credentials from. Example: `default`
- `container` (required): The container to push to the ECR registry.
- `image-name` (required): The name of the image to push to the ECR registry.
- `tags` (required): The tags to apply to the image. Example: `latest`