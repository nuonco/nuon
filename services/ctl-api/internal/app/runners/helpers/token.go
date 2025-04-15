package helpers

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
)

const (
	defaultRunnerTokenTimeout time.Duration = time.Hour * 24 * 90
)

func (a *Helpers) CreateToken(ctx context.Context, runnerID string) (*app.Token, error) {
	email := account.ServiceAccountEmail(runnerID)

	token, err := a.acctClient.CreateToken(ctx, email, defaultRunnerTokenTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create token")
	}

	return token, nil
}
