package helpers

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
)

func (h *Helpers) GARAuthorization(ctx context.Context, orgID string, cfg *app.GCPGARImageConfig, imageURL string) (*oauth2.Token, error) {
	connection, err := cloudconnections.GARConnection(ctx, h.db, orgID, cfg, imageURL)
	if err != nil {
		return nil, err
	}
	source, err := cloudconnections.GARTokenSource(ctx, h.issuer, connection, fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID))
	if err != nil {
		return nil, err
	}
	return source.Token()
}
