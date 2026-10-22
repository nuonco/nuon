package service

import (
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
