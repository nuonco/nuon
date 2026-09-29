package cloudconnections

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/agentmode"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type verificationClient struct {
	nuon.Client
	result   *models.ServiceConnectionResponse
	blockGet bool
	getCalls int
}

func (c *verificationClient) VerifyCloudConnection(context.Context, string) (*models.ServiceConnectionResponse, error) {
	return &models.ServiceConnectionResponse{Status: models.AppCloudConnectionStatusError, VerificationInProgress: true}, nil
}

func (c *verificationClient) GetCloudConnection(ctx context.Context, _ string) (*models.ServiceConnectionResponse, error) {
	c.getCalls++
	if c.blockGet {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return c.result, nil
}

func TestVerify(t *testing.T) {
	for name, tc := range map[string]struct {
		wait     bool
		result   *models.ServiceConnectionResponse
		blockGet bool
		wantErr  string
		elapsed  time.Duration
	}{
		"verified": {
			wait: true, result: &models.ServiceConnectionResponse{Status: models.AppCloudConnectionStatusVerified}, elapsed: 2 * time.Second,
		},
		"terminal error": {
			wait: true, result: &models.ServiceConnectionResponse{Status: models.AppCloudConnectionStatusError, StatusMessage: "Nuon OIDC identity is not trusted by this role."},
			wantErr: "cloud connection clc-example verification ended with status error: Nuon OIDC identity is not trusted by this role.", elapsed: 2 * time.Second,
		},
		"timeout while polling": {
			wait: true, result: &models.ServiceConnectionResponse{Status: models.AppCloudConnectionStatusVerified, VerificationInProgress: true},
			wantErr: "cloud connection clc-example verification timed out after 2 minutes", elapsed: 2 * time.Minute,
		},
		"timeout during request": {
			wait: true, blockGet: true,
			wantErr: "cloud connection clc-example verification timed out after 2 minutes", elapsed: 2 * time.Minute,
		},
		"no wait": {},
	} {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				api := &verificationClient{result: tc.result, blockGet: tc.blockGet}
				service := New(api, nil)
				started := time.Now()
				err := service.Verify(context.Background(), "clc-example", tc.wait, true)
				if tc.wantErr == "" {
					require.NoError(t, err)
				} else {
					var userErr *ui.CLIUserError
					require.ErrorAs(t, err, &userErr)
					assert.Equal(t, tc.wantErr, userErr.Msg)
				}
				assert.Equal(t, tc.elapsed, time.Since(started))
				if tc.wait {
					assert.Positive(t, api.getCalls)
				} else {
					assert.Zero(t, api.getCalls)
				}
			})
		})
	}
}

type creationClient struct {
	nuon.Client
	connection  *models.ServiceConnectionResponse
	dashboard   string
	configErr   error
	configCalls int
}

func (c *creationClient) CreateCloudConnection(context.Context, *models.ServiceCreateRequest) (*models.ServiceConnectionResponse, error) {
	return c.connection, nil
}

func (c *creationClient) GetCLIConfig(context.Context) (*models.ServiceCLIConfig, error) {
	c.configCalls++
	return &models.ServiceCLIConfig{DashboardURL: c.dashboard}, c.configErr
}

func TestCreateNextSteps(t *testing.T) {
	for name, tc := range map[string]struct {
		dashboard string
		configErr error
		asJSON    bool
		agent     bool
		wantLink  string
	}{
		"dashboard link": {dashboard: "https://app.example.com", wantLink: "https://app.example.com/org-acme/settings/cloud-connections/clc-example/setup"},
		"BYOC subpath":   {dashboard: "https://example.com/dashboard/", wantLink: "https://example.com/dashboard/org-acme/settings/cloud-connections/clc-example/setup"},
		"config failure": {configErr: errors.New("unavailable")},
		"missing URL":    {},
		"JSON":           {asJSON: true},
		"agent":          {asJSON: true, agent: true},
	} {
		t.Run(name, func(t *testing.T) {
			output, err := os.CreateTemp(t.TempDir(), "stdout")
			require.NoError(t, err)
			stdout, wasAgent := os.Stdout, agentmode.Enabled()
			os.Stdout = output
			agentmode.SetEnabled(tc.agent)
			t.Cleanup(func() {
				os.Stdout = stdout
				agentmode.SetEnabled(wasAgent)
				output.Close()
			})
			connection := &models.ServiceConnectionResponse{
				ID: "clc-example", OrgID: "org-acme", Name: "acme-production", Status: models.AppCloudConnectionStatusPending,
				Setup: &models.ServiceSetupResponse{IssuerURL: "https://api.example.com"},
			}
			api := &creationClient{connection: connection, dashboard: tc.dashboard, configErr: tc.configErr}
			require.NoError(t, New(api, nil).Create(context.Background(), "acme-production", "aws", "123456789012", "arn:aws:iam::123456789012:role/nuon", "us-west-2", "stacks", tc.asJSON))
			bytes, err := os.ReadFile(output.Name())
			require.NoError(t, err)
			if tc.asJSON {
				var want any = connection
				if tc.agent {
					want = map[string]any{"ok": true, "data": connection}
				}
				expected, err := json.Marshal(want)
				require.NoError(t, err)
				assert.JSONEq(t, string(expected), string(bytes))
				assert.Zero(t, api.configCalls)
				return
			}
			assert.Equal(t, 1, api.configCalls)
			assert.Contains(t, string(bytes), "AWS CLI, Terraform, or CloudFormation")
			assert.Contains(t, string(bytes), "nuon cloud-connections verify clc-example")
			if tc.wantLink != "" {
				assert.Contains(t, string(bytes), tc.wantLink)
			} else {
				assert.Contains(t, string(bytes), "Settings > Cloud connections > select this connection > View setup runbook")
			}
		})
	}
}
