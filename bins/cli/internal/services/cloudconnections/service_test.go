package cloudconnections

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
