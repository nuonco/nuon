package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type Helpers struct {
	issuer *oidcissuer.Issuer
}

type Params struct {
	fx.In
	Cfg *internal.Config
}

func New(params Params) (*Helpers, error) {
	if params.Cfg.TelemetryJWKS == "" {
		return &Helpers{}, nil
	}
	privateKey, keyID, _, err := oidcissuer.ParseJWKS(params.Cfg.TelemetryJWKS)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud connection issuer: %w", err)
	}
	issuer, err := oidcissuer.New(params.Cfg.PublicAPIURL, privateKey, keyID)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud connection issuer: %w", err)
	}
	return &Helpers{issuer: issuer}, nil
}

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
	output, err := sts.NewFromConfig(cfg).AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
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
