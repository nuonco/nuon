package credentials

import (
	"context"
	"fmt"

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

func (c *Config) fetchCredentials(ctx context.Context) (aws.Config, error) {
	v := validator.New()

	// if default credentials are set, just use the machine's credentials
	if c.Default {
		awsCfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return aws.Config{}, fmt.Errorf("unable to load static credentials: %w", err)
		}

		return awsCfg, nil
	}

	// if static credentials are set, prefer those
	if c.StaticCredentials != (StaticCredentials{}) {
		provider := credentials.NewStaticCredentialsProvider(
			c.AccessKeyID,
			c.SecretAccessKey,
			c.SessionToken)

		awsCfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(provider))
		if err != nil {
			return aws.Config{}, fmt.Errorf("unable to load static credentials: %w", err)
		}
		return awsCfg, nil
	}

	assumer, err := assumerole.New(v, assumerole.WithSettings(assumerole.Settings{
		RoleARN:             c.RoleARN,
		RoleSessionName:     c.SessionName,
		RoleSessionDuration: c.SessionDuration,
	}))
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to create role assumer: %w", err)
	}

	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to assume role: %w", err)
	}

	return cfg, nil
}
