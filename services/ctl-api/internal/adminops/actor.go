package adminops

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	actorServiceAccountID   = "ctl-api-admin-job"
	actorServiceAccountName = "ctl-api admin job"
)

func (o *Ops) WithActor(ctx context.Context) (context.Context, error) {
	acct, err := o.acctClient.EnsureServiceAccount(ctx, actorServiceAccountID, actorServiceAccountName)
	if err != nil {
		return nil, fmt.Errorf("unable to ensure admin job service account: %w", err)
	}

	return cctx.SetAccountContext(ctx, acct), nil
}
