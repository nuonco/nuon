package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/orgregistry"
)

type GetOrgRegistryRepositoryRequest struct {
	AppID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetOrgRegistryRepository(ctx context.Context, req *GetOrgRegistryRepositoryRequest) (*configs.OCIRegistryRepository, error) {
	if req == nil || req.AppID == "" {
		return nil, fmt.Errorf("app id is required")
	}
	return orgregistry.New(a.v, a.db, a.cfg).Repository(ctx, req.AppID)
}
