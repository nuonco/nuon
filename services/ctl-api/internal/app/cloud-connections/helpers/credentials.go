package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) Credentials(ctx context.Context, connection *app.CloudConnection, sessionName string) (*credentials.Config, error) {
	if h.issuer == nil {
		return nil, fmt.Errorf("cloud connection OIDC issuer is unavailable")
	}
	if connection.Platform != app.CloudPlatformAWS || connection.Status != app.CloudConnectionStatusVerified {
		return nil, fmt.Errorf("cloud connection must be a verified AWS connection")
	}
	token, err := h.issuer.Mint(ctx, fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID), "sts.amazonaws.com", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(connection.DefaultRegion))
	if err != nil {
		return nil, err
	}
	output, err := h.assumeRole(ctx, connection, sts.NewFromConfig(cfg), &sts.AssumeRoleWithWebIdentityInput{
		RoleArn: &connection.Principal, RoleSessionName: &sessionName, WebIdentityToken: &token, DurationSeconds: aws.Int32(900),
	})
	if err != nil {
		return nil, fmt.Errorf("assume cloud connection role: %w", err)
	}
	value := output.Credentials
	if value == nil || value.AccessKeyId == nil || value.SecretAccessKey == nil || value.SessionToken == nil {
		return nil, fmt.Errorf("assume role response did not include credentials")
	}
	return &credentials.Config{Region: connection.DefaultRegion, Static: &credentials.StaticCredentials{
		AccessKeyID: *value.AccessKeyId, SecretAccessKey: *value.SecretAccessKey, SessionToken: *value.SessionToken,
	}}, nil
}
