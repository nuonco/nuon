package service

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	runnerTenantID    = "11111111-2222-3333-4444-555555555555"
	runnerSubID       = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	runnerPrincipalID = "99999999-8888-7777-6666-555555555555"
	runnerInstallID   = "inst123"
)

func storedAzureOutputs() *app.AzureStackOutputs {
	return &app.AzureStackOutputs{
		SubscriptionID:            runnerSubID,
		SubscriptionTenantID:      runnerTenantID,
		RunnerIdentityPrincipalID: runnerPrincipalID,
	}
}

func runnerClaims() *azureClaims {
	return &azureClaims{
		TenantID: runnerTenantID,
		ObjectID: runnerPrincipalID,
		XMSMirID: "/subscriptions/" + runnerSubID + "/resourcegroups/rg-1/providers/Microsoft.ManagedIdentity/userAssignedIdentities/" + runnerInstallID,
	}
}

func TestMatchAzureRunnerClaims(t *testing.T) {
	for _, tc := range []struct {
		name          string
		outputs       func(*app.AzureStackOutputs)
		claims        func(*azureClaims)
		subscription  string
		wantSubstring string
	}{
		{
			name:         "the stored principal is accepted",
			subscription: runnerSubID,
		},
		{
			name: "a differently cased principal is accepted",
			claims: func(c *azureClaims) {
				c.ObjectID = strings.ToUpper(runnerPrincipalID)
			},
			subscription: runnerSubID,
		},
		{
			name: "an identity named like the runner with a different principal is rejected",
			claims: func(c *azureClaims) {
				c.ObjectID = "00000000-8888-7777-6666-555555555555"
			},
			subscription:  runnerSubID,
			wantSubstring: "principal id mismatch",
		},
		{
			name: "an install with no stored principal is rejected",
			outputs: func(o *app.AzureStackOutputs) {
				o.RunnerIdentityPrincipalID = ""
			},
			subscription:  runnerSubID,
			wantSubstring: "does not have a runner identity principal id",
		},
		{
			name: "a token from another tenant is rejected",
			claims: func(c *azureClaims) {
				c.TenantID = "ffffffff-2222-3333-4444-555555555555"
			},
			subscription:  runnerSubID,
			wantSubstring: "tenant ID mismatch",
		},
		{
			name:          "a token from another subscription is rejected",
			subscription:  "ffffffff-bbbb-cccc-dddd-eeeeeeeeeeee",
			wantSubstring: "subscription ID mismatch",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			outputs, claims := storedAzureOutputs(), runnerClaims()
			if tc.outputs != nil {
				tc.outputs(outputs)
			}
			if tc.claims != nil {
				tc.claims(claims)
			}

			err := matchAzureRunnerClaims(runnerInstallID, outputs, claims, tc.subscription)
			if tc.wantSubstring == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantSubstring)
		})
	}
}

func TestParseUnverifiedTenantID_AcceptsCompactGUID(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"tid":"` + runnerTenantID + `"}`))
	token := "eyJhbGciOiJSUzI1NiJ9." + payload + ".sig"

	got, err := parseUnverifiedTenantID(token)
	require.NoError(t, err)
	assert.Equal(t, runnerTenantID, got)
}

func TestParseUnverifiedTenantID_RejectsNUO1188Shapes(t *testing.T) {
	attackerTenant := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	smuggled := base64.RawURLEncoding.EncodeToString([]byte(
		`{"tid":"` + attackerTenant + `?appid=bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb#"}`,
	))
	signedTarget := base64.RawURLEncoding.EncodeToString([]byte(`{"tid":"` + runnerTenantID + `"}`))

	flattened, err := json.Marshal(map[string]string{
		"protected": "eyJhbGciOiJSUzI1NiJ9",
		"payload":   signedTarget,
		"signature": "sig",
		"extra":     "." + smuggled + ".",
	})
	require.NoError(t, err)

	compactInjected := "eyJhbGciOiJSUzI1NiJ9." + smuggled + ".sig"

	for _, tc := range []struct {
		name  string
		token string
	}{
		{"flattened JSON with dotted extra", string(flattened)},
		{"compact tid with appid query", compactInjected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseUnverifiedTenantID(tc.token)
			require.Error(t, err)
		})
	}
}

func TestBindAzureKeyTenant(t *testing.T) {
	require.NoError(t, bindAzureKeyTenant(runnerTenantID, &azureClaims{TenantID: runnerTenantID}))
	require.NoError(t, bindAzureKeyTenant(runnerTenantID, &azureClaims{TenantID: strings.ToUpper(runnerTenantID)}))

	err := bindAzureKeyTenant(runnerTenantID, &azureClaims{TenantID: "ffffffff-2222-3333-4444-555555555555"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not match key tenant")
}
