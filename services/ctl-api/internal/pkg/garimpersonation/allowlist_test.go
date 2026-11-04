package garimpersonation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	ownerOrg   = "orgowner0000000000000000000"
	otherOrg   = "orgother0000000000000000000"
	vendorSA   = "gar-reader@acme-artifacts.iam.gserviceaccount.com"
	vendorWIF  = "projects/123456789/locations/global/workloadIdentityPools/nuon/providers/aws-111111111111"
	unknownSA  = "gar-reader@other-artifacts.iam.gserviceaccount.com"
	unknownWIF = "projects/987654321/locations/global/workloadIdentityPools/nuon/providers/aws-222222222222"
)

func TestValidate(t *testing.T) {
	allowlist := []string{
		ownerOrg + ":" + vendorSA,
		" " + ownerOrg + ":" + vendorWIF + " ",
		"",
	}

	tests := []struct {
		name      string
		allowlist []string
		orgID     string
		sa        string
		wif       string
		errSubstr string
	}{
		{name: "no identity uses control plane credentials", allowlist: nil, orgID: otherOrg},
		{name: "no identity without org", allowlist: nil},
		{name: "allowlisted service account", allowlist: allowlist, orgID: ownerOrg, sa: vendorSA},
		{name: "allowlisted service account and provider", allowlist: allowlist, orgID: ownerOrg, sa: vendorSA, wif: vendorWIF},
		{name: "allowlisted provider alone", allowlist: allowlist, orgID: ownerOrg, wif: vendorWIF},
		{
			name: "service account bound to another org", allowlist: allowlist, orgID: otherOrg, sa: vendorSA,
			errSubstr: `service_account_email "` + vendorSA + `" is not allowlisted for org ` + otherOrg,
		},
		{
			name: "provider bound to another org", allowlist: allowlist, orgID: otherOrg, wif: vendorWIF,
			errSubstr: "workload_identity_provider",
		},
		{
			name: "unknown service account", allowlist: allowlist, orgID: ownerOrg, sa: unknownSA,
			errSubstr: unknownSA,
		},
		{
			name: "allowlisted service account through unknown provider", allowlist: allowlist, orgID: ownerOrg, sa: vendorSA, wif: unknownWIF,
			errSubstr: unknownWIF,
		},
		{
			name: "empty allowlist fails closed", allowlist: nil, orgID: ownerOrg, sa: vendorSA,
			errSubstr: "not allowlisted",
		},
		{
			name: "identity without org", allowlist: allowlist, sa: vendorSA,
			errSubstr: "requires an org",
		},
		{
			name: "whitespace around separator", allowlist: []string{ownerOrg + " : " + vendorSA}, orgID: ownerOrg, sa: vendorSA,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			list, err := Parse(tc.allowlist)
			require.NoError(t, err)

			err = list.Validate(tc.orgID, tc.sa, tc.wif)
			if tc.errSubstr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.errSubstr)
		})
	}
}

func TestParse(t *testing.T) {
	for _, entry := range []string{vendorSA, ownerOrg + ":", ":" + vendorSA, " : "} {
		t.Run(entry, func(t *testing.T) {
			_, err := Parse([]string{ownerOrg + ":" + vendorSA, entry})
			require.ErrorContains(t, err, "invalid gcp_gar_impersonation_allowlist entry")
		})
	}
}

func TestValidateDoesNotLeakOtherOrgs(t *testing.T) {
	list, err := Parse([]string{otherOrg + ":" + unknownSA})
	require.NoError(t, err)

	err = list.Validate(ownerOrg, vendorSA, "")
	require.Error(t, err)
	require.NotContains(t, err.Error(), otherOrg)
	require.NotContains(t, err.Error(), unknownSA)
}
