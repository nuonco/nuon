package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	defaultLogStreamTokenDuration time.Duration = time.Hour
)

type CreateLogStreamRequest struct {
	OperationID string `validate:"required"`
}

// @temporal-gen activity
// @by-id OperationID
func (a *Activities) CreateLogStream(ctx context.Context, req CreateLogStreamRequest) (*app.LogStream, error) {
	typ := "runner_operations"

	ls := app.LogStream{
		OwnerType: typ,
		OwnerID:   req.OperationID,
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

	token, err := a.acctClient.CreateToken(ctx, svcAcct.Email, defaultLogStreamTokenDuration)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create token")
	}

	// this token is only available on the temporal response, and is not persisted to the log stream object
	ls.WriteToken = token.Token
	ls.RunnerAPIURL = a.cfg.RunnerAPIURL
	return &ls, nil
}
