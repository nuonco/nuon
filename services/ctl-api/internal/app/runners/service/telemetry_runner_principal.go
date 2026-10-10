package service

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetrytoken"
)

var errTelemetryRunnerUnauthorized = errors.New("runner is not authorized to export telemetry")

func (s *service) resolveTelemetryRunnerPrincipal(ctx context.Context, acct *app.Account) (telemetrytoken.Principal, error) {
	var principal telemetrytoken.Principal

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return principal, fmt.Errorf("get telemetry runner org: %w", err)
	}
	if acct == nil || acct.AccountType != app.AccountTypeService || acct.Subject == "" || !slices.Contains(acct.OrgIDs, orgID) || !hasTelemetryRunnerRole(acct, orgID) {
		return principal, telemetryRunnerAuthorizationError()
	}

	var runner app.Runner
	err = s.db.WithContext(ctx).
		Preload("RunnerGroup").
		Where(app.Runner{ID: acct.Subject, OrgID: orgID}).
		First(&runner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return principal, telemetryRunnerAuthorizationError()
		}
		return principal, fmt.Errorf("get telemetry runner: %w", err)
	}

	if runner.Status == app.RunnerStatusDisabled || runner.Status == app.RunnerStatusDeprovisioned {
		return principal, telemetryRunnerAuthorizationError()
	}

	group := runner.RunnerGroup
	if group.OrgID != orgID || group.Type != app.RunnerGroupTypeInstall || group.OwnerType != plugins.TableName(s.db, app.Install{}) {
		return principal, telemetryRunnerAuthorizationError()
	}

	principal, err = s.installsHelpers.GetInstallTelemetryTokenPrincipal(ctx, orgID, group.OwnerID, acct.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return principal, telemetryRunnerAuthorizationError()
		}
		return principal, fmt.Errorf("get telemetry runner install: %w", err)
	}
	if principal.RelayEndpoint == "" {
		return principal, stderr.ErrUser{Err: fmt.Errorf("telemetry relay is not configured"), Description: "telemetry relay is not configured"}
	}
	if err := app.ValidateTelemetryRelayEndpoint(principal.RelayEndpoint); err != nil {
		return principal, err
	}
	return principal, nil
}

func hasTelemetryRunnerRole(acct *app.Account, orgID string) bool {
	for _, role := range acct.Roles {
		if role.RoleType == app.RoleTypeRunner && role.OrgID.ValueString() == orgID {
			return true
		}
	}
	return false
}

func telemetryRunnerAuthorizationError() error {
	return stderr.ErrAuthorization{
		Err:         errTelemetryRunnerUnauthorized,
		Description: errTelemetryRunnerUnauthorized.Error(),
	}
}
