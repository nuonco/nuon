package helpers

import (
	"context"

	"github.com/nuonco/nuon/pkg/azure/acr"
	"github.com/nuonco/nuon/pkg/oci/metadata"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
)

func (h *Helpers) ACRAuthorization(ctx context.Context, orgID string, cfg *app.AzureACRImageConfig, imageURL string) (*metadata.RegistryAuth, error) {
	connection, err := cloudconnections.ACRConnection(ctx, h.db, orgID, cfg, imageURL, h.cfg.LegacyImageAuthEnabled)
	if err != nil || connection == nil {
		return nil, err
	}
	credential, err := cloudconnections.ACRCredential(h.issuer, connection, cloudconnections.Subject(connection))
	if err != nil {
		return nil, err
	}
	token, err := acr.GetRepositoryTokenWithCredential(ctx, credential, connection.AzureTenantID, connection.AzureRegistryLoginServer)
	if err != nil {
		return nil, err
	}
	return &metadata.RegistryAuth{ServerAddress: "https://" + connection.AzureRegistryLoginServer, Username: acr.DefaultACRUsername, Password: token}, nil
}
