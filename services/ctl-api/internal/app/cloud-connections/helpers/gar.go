package helpers

import (
	"context"

	"golang.org/x/oauth2"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
)

func (h *Helpers) GARAuthorization(ctx context.Context, orgID string, cfg *app.GCPGARImageConfig, imageURL string) (*oauth2.Token, error) {
	connection, err := cloudconnections.GARConnection(ctx, h.db, orgID, cfg, imageURL, h.cfg.LegacyImageAuthEnabled)
	if err != nil {
		return nil, err
	}
	if connection == nil {
		return h.legacyGARAuthorization(ctx, cfg)
	}
	source, err := cloudconnections.GARTokenSource(ctx, h.issuer, connection, cloudconnections.Subject(connection))
	if err != nil {
		return nil, err
	}
	return source.Token()
}
