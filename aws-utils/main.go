// A generated module for AwsUtils functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"

	"dagger/aws-utils/internal/dagger"
	"dagger/aws-utils/pkg"
)

const (
	GoImage     = "golang:1.23"
	CacheVolume = "aws"
)

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
func (m *AwsUtils) RetrieveCredentials(ctx context.Context, awsDir *dagger.Directory, awsProfile string) (string, error) {
	return m.util(ctx, awsDir, awsProfile, []string{pkg.CommandRetrieveCredentials}).Stdout(ctx)
}

func (m *AwsUtils) GetEcrToken(ctx context.Context, awsDir *dagger.Directory, awsProfile string) (string, error) {
	return m.util(ctx, awsDir, awsProfile, []string{pkg.CommandEcrGetToken}).Stdout(ctx)
}
