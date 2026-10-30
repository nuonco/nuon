package helpers

import (
	"context"

	"github.com/go-playground/validator/v10"

	ecr "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
)

func (h *Helpers) ECRAuthorization(ctx context.Context, orgID string, cfg *app.AWSECRImageConfig, imageURL string) (*ecr.Authorization, error) {
	connection, err := cloudconnections.ECRConnection(ctx, h.db, orgID, cfg, imageURL)
	if err != nil {
		return nil, err
	}
	creds, err := h.Credentials(ctx, connection, "nuon-ecr-pull")
	if err != nil {
		return nil, err
	}
	creds.Region = cfg.AWSRegion
	client, err := ecr.New(validator.New(), ecr.WithCredentials(creds))
	if err != nil {
		return nil, err
	}
	return client.GetAuthorization(ctx)
}
