package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

func TestExternalImageRequestAuthentication(t *testing.T) {
	for name, tc := range map[string]struct {
		request    CreateExternalImageComponentConfigRequest
		connection bool
		wantError  bool
	}{
		"legacy ECR":         {request: CreateExternalImageComponentConfigRequest{ImageURL: "123456789012.dkr.ecr.us-east-1.amazonaws.com/acme", AWSECRImageConfig: &awsECRImageConfigRequest{IAMRoleARN: "arn:aws:iam::123456789012:role/acme", AWSRegion: "us-east-1"}}},
		"legacy GAR":         {request: CreateExternalImageComponentConfigRequest{ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", GCPGARImageConfig: &gcpGARImageConfigRequest{GCPProjectID: "acme", GCPRegion: "us-central1", ServiceAccountEmail: "acme@example.com", WorkloadIdentityProvider: "provider"}}},
		"ambient GAR":        {request: CreateExternalImageComponentConfigRequest{ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", GCPGARImageConfig: &gcpGARImageConfigRequest{GCPProjectID: "acme", GCPRegion: "us-central1"}}},
		"connection GAR":     {request: CreateExternalImageComponentConfigRequest{ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", GCPGARImageConfig: &gcpGARImageConfigRequest{UseCloudConnection: true, GCPProjectID: "acme", GCPRegion: "us-central1"}}, connection: true},
		"mixed GAR account":  {request: CreateExternalImageComponentConfigRequest{ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", GCPGARImageConfig: &gcpGARImageConfigRequest{UseCloudConnection: true, ServiceAccountEmail: "acme@example.com"}}, wantError: true},
		"mixed GAR provider": {request: CreateExternalImageComponentConfigRequest{ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", GCPGARImageConfig: &gcpGARImageConfigRequest{UseCloudConnection: true, WorkloadIdentityProvider: "provider"}}, wantError: true},
		"ambient ACR":        {request: CreateExternalImageComponentConfigRequest{ImageURL: "acme.azurecr.io/repo", AzureACRImageConfig: &azureACRImageConfigRequest{RegistryURL: "acme.azurecr.io"}}},
		"connection ACR":     {request: CreateExternalImageComponentConfigRequest{ImageURL: "acme.azurecr.io/repo", AzureACRImageConfig: &azureACRImageConfigRequest{RegistryURL: "acme.azurecr.io", UseCloudConnection: true}}, connection: true},
		"secret ACR":         {request: CreateExternalImageComponentConfigRequest{ImageURL: "acme.azurecr.io/repo", AzureACRImageConfig: &azureACRImageConfigRequest{RegistryURL: "acme.azurecr.io", TenantID: "tenant", ClientID: "client", ClientSecretName: "secret"}}},
		"certificate ACR":    {request: CreateExternalImageComponentConfigRequest{ImageURL: "acme.azurecr.io/repo", AzureACRImageConfig: &azureACRImageConfigRequest{RegistryURL: "acme.azurecr.io", TenantID: "tenant", ClientID: "client", ClientCertificateName: "cert"}}},
		"mixed ACR":          {request: CreateExternalImageComponentConfigRequest{ImageURL: "acme.azurecr.io/repo", AzureACRImageConfig: &azureACRImageConfigRequest{UseCloudConnection: true, ClientSecretName: "secret"}}, wantError: true},
		"incomplete ACR":     {request: CreateExternalImageComponentConfigRequest{ImageURL: "acme.azurecr.io/repo", AzureACRImageConfig: &azureACRImageConfigRequest{ClientID: "client"}}, wantError: true},
	} {
		t.Run(name, func(t *testing.T) {
			tc.request.Tag = "v1"
			err := tc.request.Validate(validator.New())
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			cfg, err := build.ExternalImageComponentConfig(tc.request.toConfig())
			require.NoError(t, err)
			if tc.connection {
				if cfg.GCPGARImageConfig != nil {
					require.True(t, cfg.GCPGARImageConfig.UseCloudConnection)
				} else {
					require.True(t, cfg.AzureACRImageConfig.UseCloudConnection)
				}
			} else {
				require.NoError(t, cloudconnections.ResolveImageConnection(context.Background(), nil, "org_acme", &app.ComponentConfigConnection{ExternalImageComponentConfig: cfg}))
			}
		})
	}
}

func TestExternalImageInvalidRequestHTTP(t *testing.T) {
	for name, tc := range map[string]struct {
		body        string
		description string
	}{
		"mixed GAR account": {
			body:        `{"image_url":"us-central1-docker.pkg.dev/acme/repo/image","tag":"v1","gcp_gar_image_config":{"use_cloud_connection":true,"service_account_email":"acme@example.com"}}`,
			description: "use_cloud_connection cannot be combined",
		},
		"mixed GAR provider": {
			body:        `{"image_url":"us-central1-docker.pkg.dev/acme/repo/image","tag":"v1","gcp_gar_image_config":{"use_cloud_connection":true,"workload_identity_provider":"provider"}}`,
			description: "use_cloud_connection cannot be combined",
		},
		"mixed ACR secret": {
			body:        `{"image_url":"acme.azurecr.io/repo","tag":"v1","azure_acr_image_config":{"use_cloud_connection":true,"tenant_id":"tenant","client_id":"client","client_secret_name":"secret"}}`,
			description: "use_cloud_connection cannot be combined",
		},
		"mixed ACR certificate": {
			body:        `{"image_url":"acme.azurecr.io/repo","tag":"v1","azure_acr_image_config":{"use_cloud_connection":true,"tenant_id":"tenant","client_id":"client","client_certificate_name":"cert"}}`,
			description: "use_cloud_connection cannot be combined",
		},
		"incomplete ACR": {
			body:        `{"image_url":"acme.azurecr.io/repo","tag":"v1","azure_acr_image_config":{"client_id":"client"}}`,
			description: "tenant_id",
		},
		"malformed JSON": {
			body:        `{`,
			description: "unexpected EOF",
		},
	} {
		t.Run(name, func(t *testing.T) {
			svc := &service{v: validator.New()}
			router := gin.New()
			router.Use(stderr.New(zap.NewNop(), nil).Handler())
			router.POST("/v1/components/:component_id/configs/external-image", svc.CreateExternalImageComponentConfig)
			req := httptest.NewRequest(http.MethodPost, "/v1/components/cmp_acme/configs/external-image", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)

			require.Equal(t, http.StatusBadRequest, res.Code, res.Body.String())
			var response stderr.ErrResponse
			require.NoError(t, json.Unmarshal(res.Body.Bytes(), &response))
			require.True(t, response.UserError)
			require.Equal(t, "invalid request", response.Error)
			require.Contains(t, response.Description, tc.description)
		})
	}
}
