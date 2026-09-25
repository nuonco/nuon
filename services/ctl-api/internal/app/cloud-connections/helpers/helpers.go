package helpers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/smithy-go"
	"gorm.io/gorm"

	assumerole "github.com/nuonco/nuon/pkg/aws/assume-role"
	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type Helpers struct {
	db     *gorm.DB
	cfg    *internal.Config
	issuer *oidcissuer.Issuer
}

func New(db *gorm.DB, cfg *internal.Config) (*Helpers, error) {
	if cfg.TelemetryJWKS == "" {
		return &Helpers{db: db, cfg: cfg}, nil
	}
	privateKey, keyID, _, err := oidcissuer.ParseJWKS(cfg.TelemetryJWKS)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud connection issuer: %w", err)
	}
	issuer, err := oidcissuer.New(cfg.PublicAPIURL, privateKey, keyID)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud connection issuer: %w", err)
	}
	return &Helpers{db: db, cfg: cfg, issuer: issuer}, nil
}

func (h *Helpers) Credentials(ctx context.Context, connection *app.CloudConnection, sessionName string) (*credentials.Config, error) {
	if h.issuer == nil {
		return nil, fmt.Errorf("cloud connection OIDC issuer is unavailable")
	}
	token, err := h.issuer.Mint(ctx, fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID), "sts.amazonaws.com", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	return &credentials.Config{Region: connection.DefaultRegion, AssumeRole: &credentials.AssumeRoleConfig{RoleARN: connection.Principal, SessionName: sessionName, SessionDurationSeconds: 900, WebIdentityToken: token}}, nil
}

func (h *Helpers) ECRCredentials(ctx context.Context, connection *app.CloudConnection, sessionName string) (*credentials.Config, error) {
	oidc, err := h.Credentials(ctx, connection, sessionName)
	if err != nil {
		return nil, err
	}
	var legacy *credentials.Config
	mode, err := chooseAuthMode(connection.AuthMode,
		func() error { _, err := credentials.Fetch(ctx, oidc); return err },
		func() error {
			var err error
			legacy, err = h.legacyCredentials(connection, sessionName)
			if err != nil {
				return err
			}
			_, err = credentials.Fetch(ctx, legacy)
			return err
		},
	)
	if err != nil {
		return nil, err
	}
	selected := oidc
	if mode == app.CloudConnectionAuthModeLegacy {
		selected = legacy
	}
	if mode != connection.AuthMode {
		if err := h.db.WithContext(ctx).Model(&app.CloudConnection{}).Where(app.CloudConnection{ID: connection.ID, OrgID: connection.OrgID}).Update("auth_mode", mode).Error; err != nil {
			return nil, fmt.Errorf("cache cloud connection auth mode: %w", err)
		}
		connection.AuthMode = mode
	}
	return selected, nil
}

func (h *Helpers) legacyCredentials(connection *app.CloudConnection, sessionName string) (*credentials.Config, error) {
	assume := &credentials.AssumeRoleConfig{RoleARN: connection.Principal, SessionName: sessionName, SessionDurationSeconds: 900}
	switch {
	case h.cfg.IsAWS() && h.cfg.ManagementIAMRoleARN != "":
		assume.TwoStepConfig = &assumerole.TwoStepConfig{IAMRoleARN: h.cfg.ManagementIAMRoleARN}
	case h.cfg.IsGCP():
		assume.UseGCPOIDC = true
	default:
		return nil, fmt.Errorf("legacy AWS cloud connection authentication is unavailable on this control plane")
	}
	return &credentials.Config{Region: connection.DefaultRegion, AssumeRole: assume}, nil
}

func chooseAuthMode(current app.CloudConnectionAuthMode, oidc, legacy func() error) (app.CloudConnectionAuthMode, error) {
	if current == app.CloudConnectionAuthModeOIDC {
		if err := oidc(); err != nil {
			return "", err
		}
		return app.CloudConnectionAuthModeOIDC, nil
	}
	if err := oidc(); err == nil {
		return app.CloudConnectionAuthModeOIDC, nil
	} else if !accessDenied(err) {
		return "", err
	}
	if err := legacy(); err != nil {
		return "", err
	}
	return app.CloudConnectionAuthModeLegacy, nil
}

func accessDenied(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && (apiErr.ErrorCode() == "AccessDenied" || apiErr.ErrorCode() == "AccessDeniedException")
}
