package credentials

import (
	"context"
	"fmt"
	"os"
	"time"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/go-playground/validator/v10"

	assumerole "github.com/nuonco/nuon/pkg/aws/assume-role"
)

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
	return fmt.Sprintf("unable to assume role: %s: %s", e.RoleARN, e.Err.Error())
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

	if c.Profile != "" {
		awsCfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithSharedConfigProfile(c.Profile),
			config.WithRegion(c.Region))
		if err != nil {
			return aws.Config{}, fmt.Errorf("unable to load profile %s: %w", c.Profile, err)
		}

		return awsCfg, nil
	}
	if c.UseDefault {
		if os.Getenv("AWS_REGION") == "" && c.Region == "" {
			return aws.Config{}, fmt.Errorf("must set AWS_REGION in the environment or on the credentials config")
		}

		awsCfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(c.Region))
		if err != nil {
			return aws.Config{}, ErrUnableToFetchStatic{err}
		}

		return awsCfg, nil
	}

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

	if c.AssumeRole == nil {
		return aws.Config{}, fmt.Errorf("invalid config, must set either default, static or assume role")
	}

	assumer, err := assumerole.New(v, assumerole.WithSettings(assumerole.Settings{
		RoleARN:             c.AssumeRole.RoleARN,
		RoleSessionName:     c.AssumeRole.SessionName,
		RoleSessionDuration: time.Second * time.Duration(c.AssumeRole.SessionDurationSeconds),
		ExternalID:          c.AssumeRole.ExternalID,

		UseGithubOIDC: c.AssumeRole.UseGithubOIDC,
		UseGCPOIDC:    c.AssumeRole.UseGCPOIDC,
		TwoStepConfig: c.AssumeRole.TwoStepConfig,
		Region:        c.Region,
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
