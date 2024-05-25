package db

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
)

// FetchIamTokenPassword fetches an iam token which can be used as a password using the default aws credentials provider
func FetchIamTokenPassword(ctx context.Context, cfg database) (string, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}

	dbEndpoint := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	authToken, err := auth.BuildAuthToken(ctx, dbEndpoint, cfg.Region, cfg.User, awsCfg.Credentials)
	if err != nil {
		return "", err
	}

	return authToken, nil
}
