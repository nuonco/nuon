package helpers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/signals/verificationfailed"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type Helpers struct {
	issuer           *oidcissuer.Issuer
	db               *gorm.DB
	enqueueOrgSignal func(context.Context, orgshelpers.EnqueueOrgSignalParams) error
}

type Params struct {
	fx.In
	Cfg         *internal.Config
	DB          *gorm.DB `name:"psql"`
	OrgsHelpers *orgshelpers.Helpers
}

func New(params Params) (*Helpers, error) {
	h := &Helpers{db: params.DB, enqueueOrgSignal: params.OrgsHelpers.EnqueueOrgSignal}
	if params.Cfg.TelemetryJWKS == "" {
		return h, nil
	}
	privateKey, keyID, _, err := oidcissuer.ParseJWKS(params.Cfg.TelemetryJWKS)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud connection issuer: %w", err)
	}
	issuer, err := oidcissuer.New(params.Cfg.PublicAPIURL, privateKey, keyID)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud connection issuer: %w", err)
	}
	h.issuer = issuer
	return h, nil
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

type stsAPI interface {
	AssumeRoleWithWebIdentity(context.Context, *sts.AssumeRoleWithWebIdentityInput, ...func(*sts.Options)) (*sts.AssumeRoleWithWebIdentityOutput, error)
}

func (h *Helpers) assumeRole(ctx context.Context, connection *app.CloudConnection, client stsAPI, input *sts.AssumeRoleWithWebIdentityInput) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	output, err := client.AssumeRoleWithWebIdentity(ctx, input)
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return output, err
	}
	now := time.Now().UTC()
	message := cloudconnections.AssumeRoleErrorMessage(err)
	update := app.CloudConnection{Status: app.CloudConnectionStatusError, StatusMessage: message, LastVerifiedAt: &now}
	result := h.db.WithContext(ctx).Model(&app.CloudConnection{}).
		Where(app.CloudConnection{OrgID: connection.OrgID, ID: connection.ID, Principal: connection.Principal}).
		Select("status", "status_message", "last_verified_at").Updates(update)
	if result.Error != nil {
		return nil, errors.Join(err, fmt.Errorf("save cloud connection verification: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return nil, err
	}
	signalCtx := context.WithValue(ctx, keys.AccountIDCtxKey, connection.CreatedByID)
	if signalErr := h.enqueueOrgSignal(signalCtx, orgshelpers.EnqueueOrgSignalParams{
		OrgID: connection.OrgID,
		Signal: &verificationfailed.Signal{
			ConnectionID: connection.ID, ConnectionName: connection.Name, OrgID: connection.OrgID,
			Platform: connection.Platform, Message: message,
		},
		IdempotencyKey: fmt.Sprintf("cloud-connection-verification-failed-%s-%s", connection.ID, now.Format(time.RFC3339Nano)),
	}); signalErr != nil {
		return nil, errors.Join(err, fmt.Errorf("enqueue cloud connection verification failure: %w", signalErr))
	}
	return nil, err
}
