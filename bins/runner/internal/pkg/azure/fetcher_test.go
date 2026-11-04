package azure

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/pkg/azure"
)

const testAudience = "api://11111111-1111-1111-1111-111111111111"

type fakeClient struct {
	nuonrunner.Client
	settings    *models.ServiceRunnerPublicSettings
	settingsErr error
	gotToken    string
}

func (c *fakeClient) GetPublicSettings(context.Context, string) (*models.ServiceRunnerPublicSettings, error) {
	return c.settings, c.settingsErr
}

func (c *fakeClient) RunnerAuthAzure(_ context.Context, req *models.ServiceRunnerAuthAzureRequest) (*models.ServiceRunnerAuthAzureResponse, error) {
	c.gotToken = *req.Token
	return &models.ServiceRunnerAuthAzureResponse{Authenticated: true, RunnerID: req.RunnerID, Token: "runner-token"}, nil
}

// fakeIMDS issues "token-for:<resource>" for allowed resources. Everything else
// gets deniedBody, by default what Entra returns for an app with no service
// principal in the tenant.
func fakeIMDS(t *testing.T, deniedStatus int, deniedBody string, allowed ...string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resource := r.URL.Query().Get("resource")
		for _, a := range allowed {
			if resource == a {
				w.Write([]byte(`{"access_token":"token-for:` + resource + `"}`))
				return
			}
		}
		w.WriteHeader(deniedStatus)
		w.Write([]byte(deniedBody))
	}))
	t.Cleanup(srv.Close)
	orig := imdsTokenURL
	imdsTokenURL = srv.URL
	t.Cleanup(func() { imdsTokenURL = orig })
}

func TestFetchToken_Audience(t *testing.T) {
	tests := []struct {
		name         string
		settings     *models.ServiceRunnerPublicSettings
		settingsErr  error
		env          string
		imdsAllowed  []string
		imdsStatus   int
		imdsBody     string
		wantAudience string
		wantErr      bool
	}{
		{
			name:         "control plane audience",
			settings:     &models.ServiceRunnerPublicSettings{AzureAuthAudience: testAudience},
			imdsAllowed:  []string{testAudience, azure.ManagementAudience},
			wantAudience: testAudience,
		},
		{
			name:         "tenant has not consented falls back to ARM",
			settings:     &models.ServiceRunnerPublicSettings{AzureAuthAudience: testAudience},
			imdsAllowed:  []string{azure.ManagementAudience},
			wantAudience: azure.ManagementAudience,
		},
		{
			name:        "other IMDS errors do not fall back",
			settings:    &models.ServiceRunnerPublicSettings{AzureAuthAudience: testAudience},
			imdsAllowed: []string{azure.ManagementAudience},
			imdsStatus:  http.StatusInternalServerError,
			imdsBody:    `{"error":"unknown_error","error_description":"temporarily unavailable"}`,
			wantErr:     true,
		},
		{
			name:         "control plane without an audience uses ARM",
			settings:     &models.ServiceRunnerPublicSettings{},
			imdsAllowed:  []string{testAudience, azure.ManagementAudience},
			wantAudience: azure.ManagementAudience,
		},
		{
			name:         "older control plane keeps the stack env audience",
			settings:     &models.ServiceRunnerPublicSettings{},
			env:          testAudience,
			imdsAllowed:  []string{testAudience, azure.ManagementAudience},
			wantAudience: testAudience,
		},
		{
			name:         "settings unreachable keeps the stack env audience",
			settingsErr:  errors.New("unreachable"),
			env:          testAudience,
			imdsAllowed:  []string{testAudience, azure.ManagementAudience},
			wantAudience: testAudience,
		},
		{
			name:         "invalid control plane audience uses ARM",
			settings:     &models.ServiceRunnerPublicSettings{AzureAuthAudience: "https://vault.azure.net"},
			imdsAllowed:  []string{azure.ManagementAudience},
			wantAudience: azure.ManagementAudience,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AZURE_RUNNER_AUTH_RESOURCE", tt.env)
			status := tt.imdsStatus
			if status == 0 {
				status = http.StatusBadRequest
			}
			body := tt.imdsBody
			if body == "" {
				body = `{"error":"invalid_resource","error_description":"AADSTS500011: The resource principal named ` + testAudience + ` was not found in the tenant."}`
			}
			fakeIMDS(t, status, body, tt.imdsAllowed...)
			client := &fakeClient{settings: tt.settings, settingsErr: tt.settingsErr}

			got, err := (&Fetcher{RunnerID: "run123"}).FetchToken(context.Background(), client)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got audience %q", got.Audience)
				}
				if client.gotToken != "" {
					t.Fatalf("sent a token after a non-consent IMDS error: %q", client.gotToken)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Audience != tt.wantAudience {
				t.Errorf("audience = %q, want %q", got.Audience, tt.wantAudience)
			}
			if want := "token-for:" + tt.wantAudience; client.gotToken != want {
				t.Errorf("presented token = %q, want %q", client.gotToken, want)
			}
		})
	}
}
