package helpers

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// logStreamTokenDuration matches the activity default. Long enough to outlast
// a typical EKS provision; short enough to limit blast radius of the token.
const logStreamTokenDuration = 4 * time.Hour

// CreateLogStream inserts a LogStream row, mints a service-account-scoped
// write token, and grants it the runner role on the org. Returns the row
// with transient WriteToken + RunnerAPIURL populated for the response.
//
// Mirrors the activity at worker/activities/create_log_stream.go so the
// service handler and the activity both use one path.
func (h *Helpers) CreateLogStream(ctx context.Context, ownerType, ownerID, parentLogStreamID string) (*app.LogStream, error) {
	ls := app.LogStream{
		OwnerType:         ownerType,
		OwnerID:           ownerID,
		Open:              true,
		ParentLogStreamID: generics.NewNullString(parentLogStreamID),
	}

	if res := h.db.WithContext(ctx).Create(&ls); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create log stream")
	}

	svcAcct, err := h.accountsClient.CreateServiceAccount(ctx, ls.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create service account")
	}

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org id from context")
	}
	if err := h.authzClient.AddAccountOrgRole(ctx, app.RoleTypeRunner, orgID, svcAcct.ID); err != nil {
		return nil, errors.Wrap(err, "unable to add runner role to log stream service account")
	}

	token, err := h.accountsClient.CreateToken(ctx, svcAcct.Email, logStreamTokenDuration)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create log stream write token")
	}

	ls.WriteToken = token.Token
	ls.RunnerAPIURL = h.cfg.RunnerAPIURL
	return &ls, nil
}
