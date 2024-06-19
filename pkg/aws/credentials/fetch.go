package credentials

import (
	"context"
	"fmt"
	"time"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/go-playground/validator/v10"

	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
)

// Fetch is used to get credentials, regardless of whether they are in the context, or not. Compared to FromContext,
// this will _always_ attempt to return credentials, where as if creds are not in a context, they will not be fetched in
// FromContext
func Fetch(ctx context.Context, cfg *Config) (aws.Config, error) {
	if cfg.CacheID != "" {
		creds, err := FromContext(ctx, cfg)
		if err == nil {
			return creds, nil
		}
	}

	awsCfg, err := cfg.fetchCredentials(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to fetch creds: %w", err)
	}

	return awsCfg, nil
}

type ErrUnableToAssumeRole struct {
	RoleARN string
	Err     error
}

func (e ErrUnableToAssumeRole) Error() string {
	return fmt.Sprintf("unable to assume role: %s", e.RoleARN)
}

func (e ErrUnableToAssumeRole) Unwrap() error {
	return e.Err
}

type ErrUnableToFetchStatic struct {
	Err error
}

func (e ErrUnableToFetchStatic) Unwrap() error {
	return e.Err
}

func (e ErrUnableToFetchStatic) Error() string {
	return "unable to fetch static"
}

func (c *Config) fetchCredentials(ctx context.Context) (aws.Config, error) {
	v := validator.New()

	// if default credentials are set, just use the machine's credentials
	if c.UseDefault {
		awsCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(c.Region))
		if err != nil {
			return aws.Config{}, ErrUnableToFetchStatic{err}
		}

		return awsCfg, nil
	}

	// if static credentials are set, prefer those
	if c.Static != nil {
		provider := credentials.NewStaticCredentialsProvider(
			c.Static.AccessKeyID,
			c.Static.SecretAccessKey,
			c.Static.SessionToken)

		awsCfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(provider), config.WithRegion(c.Region))
		if err != nil {
			return aws.Config{}, ErrUnableToFetchStatic{err}
		}
		return awsCfg, nil
	}

	assumer, err := assumerole.New(v, assumerole.WithSettings(assumerole.Settings{
		RoleARN:             c.AssumeRole.RoleARN,
		RoleSessionName:     c.AssumeRole.SessionName,
		RoleSessionDuration: time.Second * time.Duration(c.AssumeRole.SessionDurationSeconds),
		TwoStepConfig:       c.AssumeRole.TwoStepConfig,
		Region:              c.Region,
	}))
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to create role assumer: %w", err)
	}

	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return aws.Config{}, ErrUnableToAssumeRole{c.AssumeRole.RoleARN, err}
	}

	return cfg, nil
}
