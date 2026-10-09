package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetrytoken"
)

// @ID EnsureInstallTelemetryCollector
// @Summary Ensure an install's telemetry collector service account
// @Description Creates or reconciles a dedicated managed identity with access only to this install's telemetry endpoints. Returns the same account on subsequent calls. Does not enable telemetry or issue credentials; use POST /v1/service-accounts/{account_id}/tokens to create a bootstrap token.
// @Tags installs
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Success 200 {object} app.ManagedServiceAccount
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/installs/{install_id}/telemetry/collector [post]
func (s *service) EnsureInstallTelemetryCollector(ctx *gin.Context) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var binding *app.ManagedServiceAccount
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		binding, err = account.New(account.Params{DB: tx}).EnsureManagedServiceAccount(ctx, account.ManagedServiceAccountRequest{
			OrgID: orgID, OwnerType: plugins.TableName(tx, app.Install{}), OwnerID: ctx.Param("install_id"),
			Purpose: app.ManagedServiceAccountPurposeTelemetryCollector, InstanceKey: "default",
			Name: "Telemetry collector",
		})
		if err != nil {
			return err
		}
		return authz.New(authz.Params{DB: tx}).EnsureManagedServiceAccountRole(ctx, binding.AccountID, &app.Role{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypeTelemetryCollector,
			Title:       "Telemetry collector",
			Description: "Scoped access to telemetry endpoints for a single install.",
			Policies: []app.Policy{{
				OrgID: generics.NewNullString(orgID),
				Name:  app.PolicyNameTelemetryCollector,
				Permissions: pgtype.Hstore{
					permissions.Object(orgID, permissions.KindTelemetry, binding.OwnerID): permissions.PermissionAll.ToStrPtr(),
				},
			}},
		})
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to ensure telemetry collector: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, binding)
}

// @ID GetInstallTelemetryCollector
// @Summary Get an install's telemetry collector service account
// @Description Returns managed identity metadata, never credentials.
// @Tags installs
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Success 200 {object} app.ManagedServiceAccount
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/installs/{install_id}/telemetry/collector [get]
func (s *service) GetInstallTelemetryCollector(ctx *gin.Context) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	binding, err := s.getInstallTelemetryCollector(ctx, orgID, ctx.Param("install_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, binding)
}

// @ID DeleteInstallTelemetryCollector
// @Summary Remove an install's telemetry collector service account
// @Description Revokes the collector's API tokens and grants, deletes its private role, and removes the identity. Does not change install telemetry settings or the runner's identity. A missing collector is success.
// @Tags installs
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Success 204
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/installs/{install_id}/telemetry/collector [delete]
func (s *service) DeleteInstallTelemetryCollector(ctx *gin.Context) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	binding, err := s.getInstallTelemetryCollector(ctx, orgID, ctx.Param("install_id"))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Error(err)
		return
	}
	if err == nil {
		if err := s.acctClient.DeleteServiceAccountByID(ctx, binding.AccountID); err != nil {
			ctx.Error(fmt.Errorf("unable to delete telemetry collector: %w", err))
			return
		}
	}
	ctx.Status(http.StatusNoContent)
}

func (s *service) getInstallTelemetryCollector(ctx context.Context, orgID, installID string) (*app.ManagedServiceAccount, error) {
	var install struct{ ID string }
	if err := s.db.WithContext(ctx).Model(&app.Install{}).Scopes(scopes.WithDisableViews).
		Select("id").Where(app.Install{ID: installID, OrgID: orgID}).Take(&install).Error; err != nil {
		return nil, err
	}
	var binding app.ManagedServiceAccount
	err := s.db.WithContext(ctx).Where(app.ManagedServiceAccount{
		OrgID: orgID, OwnerType: plugins.TableName(s.db, app.Install{}), OwnerID: installID,
		Purpose: app.ManagedServiceAccountPurposeTelemetryCollector, InstanceKey: "default",
	}).First(&binding).Error
	return &binding, err
}

type InstallTelemetryCollectorSettings struct {
	Enabled            bool              `json:"enabled"`
	RelayEndpoint      string            `json:"relay_endpoint"`
	ResourceAttributes map[string]string `json:"resource_attributes"`
}

// @ID GetInstallTelemetryCollectorSettings
// @Summary Get effective forwarding settings for an install's telemetry collector
// @Description Requires read permission on the install's telemetry resource. Returns enabled=false when telemetry, relay configuration, or token issuance is unavailable; polling remains accessible while disabled. Does not expose runner settings or backend credentials.
// @Tags installs/runner
// @Produce json
// @Security APIKey
// @Param install_id path string true "Install ID"
// @Success 200 {object} InstallTelemetryCollectorSettings
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/installs/{install_id}/telemetry/collector-settings [get]
func (s *service) GetInstallTelemetryCollectorSettings(ctx *gin.Context) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	config, err := s.helpers.GetInstallTelemetryConfig(ctx, orgID, ctx.Param("install_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get collector settings: %w", err))
		return
	}

	settings := InstallTelemetryCollectorSettings{}
	installConfig := app.InstallConfig{TelemetryEnabled: config.TelemetryEnabled}
	orgTelemetry := app.OrgTelemetrySettings{RelayEndpoint: config.OrgRelayEndpoint}
	endpoint := orgTelemetry.ResolveRelayEndpoint(s.cfg.TelemetryRelayEndpoint)
	if installConfig.IsTelemetryEnabled(config.OrgTelemetryEnabled) && s.telemetryTokenIssuer != nil && app.ValidateTelemetryRelayEndpoint(endpoint) == nil {
		settings.Enabled = true
		settings.RelayEndpoint = endpoint
		settings.ResourceAttributes = config.ResourceAttributes()
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(http.StatusOK, settings)
}

type CreateInstallTelemetryCollectorAccessTokenResponse = telemetrytoken.AccessTokenResponse

// @ID CreateInstallTelemetryCollectorAccessToken
// @Summary Create a relay access token for an install's telemetry collector
// @Description Requires create permission on the install's telemetry resource and authentication as its managed collector account. Returns a ten-minute telemetry:write JWT bound to the current relay endpoint, with no runner identity. Disabled telemetry refuses issuance; existing JWTs remain valid until expiry.
// @Tags installs/runner
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Param relay_endpoint query string true "Exact relay endpoint from collector settings"
// @Success 200 {object} CreateInstallTelemetryCollectorAccessTokenResponse
// @Failure 400 {object} stderr.ErrResponse
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Failure 503 {object} stderr.ErrResponse
// @Router /v1/installs/{install_id}/telemetry/collector-access-token [post]
func (s *service) CreateInstallTelemetryCollectorAccessToken(ctx *gin.Context) {
	if s.telemetryTokenIssuer == nil {
		ctx.JSON(http.StatusServiceUnavailable, stderr.ErrResponse{
			Error: "telemetry token issuance is unavailable", Description: "telemetry token issuance is unavailable",
		})
		return
	}
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	installID := ctx.Param("install_id")
	var binding app.ManagedServiceAccount
	err = s.db.WithContext(ctx).Where(app.ManagedServiceAccount{
		AccountID: acct.ID, OrgID: orgID,
		OwnerType: plugins.TableName(s.db, app.Install{}), OwnerID: installID,
		Purpose: app.ManagedServiceAccountPurposeTelemetryCollector, InstanceKey: "default",
	}).First(&binding).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Error(stderr.ErrAuthorization{
			Err: errors.New("account is not this install's telemetry collector"), Description: "account is not this install's telemetry collector",
		})
		return
	}
	if err != nil {
		ctx.Error(fmt.Errorf("resolve telemetry collector account: %w", err))
		return
	}
	config, err := s.helpers.GetInstallTelemetryConfig(ctx, orgID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("get telemetry collector install: %w", err))
		return
	}
	installConfig := app.InstallConfig{TelemetryEnabled: config.TelemetryEnabled}
	if !installConfig.IsTelemetryEnabled(config.OrgTelemetryEnabled) {
		ctx.Error(stderr.ErrAuthorization{
			Err: errors.New("install telemetry is disabled"), Description: "install telemetry is disabled",
		})
		return
	}
	orgTelemetry := app.OrgTelemetrySettings{RelayEndpoint: config.OrgRelayEndpoint}
	endpoint := orgTelemetry.ResolveRelayEndpoint(s.cfg.TelemetryRelayEndpoint)
	if err := app.ValidateTelemetryRelayEndpoint(endpoint); err != nil {
		ctx.JSON(http.StatusServiceUnavailable, stderr.ErrResponse{
			Error: "telemetry relay is unavailable", Description: "telemetry relay is unavailable",
		})
		return
	}
	requestedEndpoint := ctx.Query("relay_endpoint")
	if requestedEndpoint == "" {
		ctx.Error(stderr.ErrUser{
			Err: errors.New("relay_endpoint is required"), Description: "fetch collector settings and supply relay_endpoint",
		})
		return
	}
	if requestedEndpoint != endpoint {
		ctx.JSON(http.StatusConflict, stderr.ErrResponse{
			Error: "telemetry relay endpoint does not match current settings", Description: "refresh collector settings before requesting a telemetry token",
		})
		return
	}
	accessToken, err := s.telemetryTokenIssuer.Issue(telemetrytoken.Principal{
		OrgID: orgID, AppID: config.AppID, InstallID: installID,
		CollectorID: binding.AccountID, RelayEndpoint: endpoint,
	}, true)
	if err != nil {
		ctx.Error(fmt.Errorf("create collector telemetry access token: %w", err))
		return
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.Header("Pragma", "no-cache")
	ctx.JSON(http.StatusOK, CreateInstallTelemetryCollectorAccessTokenResponse{
		AccessToken: accessToken, TokenType: "Bearer", ExpiresIn: int64(telemetrytoken.Lifetime.Seconds()),
	})
}
