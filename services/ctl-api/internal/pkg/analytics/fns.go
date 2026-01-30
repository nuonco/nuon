package analytics

import (
	"context"

	"github.com/pkg/errors"
	segment "github.com/segmentio/analytics-go/v3"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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

	traits := segment.NewTraits().SetEmail(acct.Email)

	// Extract attribution from user journey metadata and add as traits
	// This enables GA4 user_id linking for marketing ROI analysis
	if len(acct.UserJourneys) > 0 {
		for _, journey := range acct.UserJourneys {
			for _, step := range journey.Steps {
				if step.Name == "account_created" && step.Metadata != nil {
					if attr, ok := step.Metadata["attribution"].(map[string]interface{}); ok {
						for k, v := range attr {
							if s, ok := v.(string); ok && s != "" {
								traits.Set(k, s)
							}
						}
					}
					break
				}
			}
		}
	}

	return &segment.Identify{
		UserId: acct.ID,
		Traits: traits,
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
