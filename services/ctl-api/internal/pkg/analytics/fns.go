package analytics

import (
	"context"

	"github.com/pkg/errors"
	segment "github.com/segmentio/analytics-go/v3"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

func groupFn(ctx context.Context) (*segment.Group, error) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "no org found")
	}

	return &segment.Group{
		GroupId: org.ID,
		Traits: map[string]interface{}{
			"name": org.Name,
			"type": org.OrgType,
		},
	}, nil
}

func identifyFn(ctx context.Context) (*segment.Identify, error) {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "no account found")
	}

	return &segment.Identify{
		UserId: acct.ID,
		Traits: segment.NewTraits().SetEmail(acct.Email),
	}, nil
}

func userIDFn(ctx context.Context) (string, error) {
	acctID, err := cctx.AccountIDFromContext(ctx)
	if err != nil {
		return "", errors.Wrap(err, "no account id found")
	}

	return acctID, nil
}

func temporalUserIDFn(ctx workflow.Context) (string, error) {
	acctID, err := cctx.AccountIDFromContext(ctx)
	if err != nil {
		return "", errors.Wrap(err, "no account id found")
	}

	return acctID, nil
}
