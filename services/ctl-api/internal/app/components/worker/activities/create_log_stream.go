package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

const (
	defaultLogStreamTokenDuration time.Duration = time.Hour
)

type CreateLogStreamRequest struct {
	BuildID string `validate:"required"`
}

// @temporal-gen activity
// @by-id BuildID
func (a *Activities) CreateLogStream(ctx context.Context, req CreateLogStreamRequest) (*app.LogStream, error) {
	typ := "component_builds"
	id := req.BuildID

	ls := app.LogStream{
		OwnerType: typ,
		OwnerID:   id,
		Open:      true,
	}

	res := a.db.WithContext(ctx).Create(&ls)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create log stream")
	}

	// create a service account to write to the log stream for up to 1 hour.
	svcAcct, err := a.acctClient.CreateServiceAccount(ctx, ls.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create service account")
	}

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org id from context")
	}

	a.authzClient.AddAccountOrgRole(ctx, app.RoleTypeRunner, orgID, svcAcct.ID)

	token, err := a.acctClient.CreateToken(ctx, svcAcct.Email, defaultLogStreamTokenDuration)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create token")
	}

	// this token is only available on the temporal response, and is not persisted to the log stream object
	ls.WriteToken = token.Token
	ls.RunnerAPIURL = a.cfg.RunnerAPIURL
	return &ls, nil
}
