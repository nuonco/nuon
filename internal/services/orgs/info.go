package services

import (
	"context"

	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
)

func (s *service) GetInfo(ctx context.Context, orgID string) (*orgsv1.GetInfoResponse, error) {
	return nil, nil
}
