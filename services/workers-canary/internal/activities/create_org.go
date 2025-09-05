package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/generics"
)

type CreateOrgRequest struct {
	CanaryID    string `validate:"required"`
	SandboxMode bool   `validate:"required"`
	APIToken    string `validate:"required"`
}

type CreateOrgResponse struct {
	OrgID string
}

func (a *Activities) CreateOrg(ctx context.Context, req *CreateOrgRequest) (*CreateOrgResponse, error) {
	apiClient, err := nuon.New(
		nuon.WithValidator(a.v),
		nuon.WithURL(a.cfg.APIURL),
		nuon.WithAuthToken(req.APIToken),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create api client: %w", err)
	}

	org, err := apiClient.CreateOrg(ctx, &models.ServiceCreateOrgRequest{
		Name:           generics.ToPtr(req.CanaryID),
		UseSandboxMode: req.SandboxMode,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create org: %w", err)
	}

	return &CreateOrgResponse{
		OrgID: org.ID,
	}, nil
}
