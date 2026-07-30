package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// validateAWSAccountConnection returns the validated connection so callers can read
// the account it names without loading the row a second time.
func (s *Helpers) validateAWSAccountConnection(ctx context.Context, connectionID string) (*app.AWSAccountConnection, error) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureAWSAccountConnections)
	if err != nil {
		return nil, fmt.Errorf("check aws account connections feature: %w", err)
	}
	if !enabled {
		return nil, stderr.ErrAuthorization{
			Err:         fmt.Errorf("aws account connections feature is not enabled"),
			Description: "AWS account connections are not enabled for this organization",
		}
	}

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}
	var connection app.AWSAccountConnection
	result := s.db.WithContext(ctx).Where("id = ? AND org_id = ?", connectionID, orgID).First(&connection)
	if result.Error != nil {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("aws account connection not found: %w", result.Error),
			Description: "AWS account connection was not found for this organization",
		}
	}
	if connection.VerificationStatus != app.AWSAccountConnectionVerificationVerified {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("aws account connection %s is not verified", connectionID),
			Description: "AWS account connection must be verified before it can be used",
		}
	}
	if connection.RoleARN == "" {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("aws account connection %s has no role ARN", connectionID),
			Description: "AWS account connection must have a role ARN before it can be used",
		}
	}
	return &connection, nil
}
