package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api"
)

type GetInstallsByOrgIDRequest struct {
	OrgID string `validate:"required"`
}

func (a *Activities) GetInstallsByOrgID(ctx context.Context, req *GetInstallsByOrgIDRequest) ([]api.Install, error) {
	installs := []api.Install{}
	internalAPIClient, err := api.New(a.v,
		api.WithURL(a.cfg.InternalAPIURL),
		api.WithAdminEmail("canary@serviceaccount.nuon.co"),
	)
	if err != nil {
		return installs, fmt.Errorf("unable to create internal api client: %w", err)
	}

	installs, err = internalAPIClient.OrgInstalls(ctx, req.OrgID)
	if err != nil {
		return installs, fmt.Errorf("unable to get installs: %w", err)
	}

	return installs, nil
}
