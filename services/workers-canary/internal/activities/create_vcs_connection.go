package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type CreateVCSConnectionRequest struct {
	CanaryID string
	OrgID    string
}

type CreateVCSConnectionResponse struct {
	VCSConnectionID string
}

func (a *Activities) CreateVCSConnection(ctx context.Context, req *CreateVCSConnectionRequest) (*CreateVCSConnectionResponse, error) {
	apiClient, err := nuon.New(a.v,
		nuon.WithURL(a.cfg.APIURL),
		nuon.WithAuthToken(a.cfg.APIToken),
		nuon.WithOrgID(req.OrgID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create api client: %w", err)
	}

	org, err := apiClient.CreateVCSConnection(ctx, &models.ServiceCreateConnectionRequest{
		GithubInstallID: generics.ToPtr(a.cfg.GithubInstallID),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create org: %w", err)
	}

	return &CreateVCSConnectionResponse{
		VCSConnectionID: org.ID,
	}, nil
}
