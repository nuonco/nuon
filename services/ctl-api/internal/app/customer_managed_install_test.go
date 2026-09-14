package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstallAppBranchUpdateEligible(t *testing.T) {
	tests := map[string]struct {
		operatingModel *InstallOperatingModel
		expected       bool
	}{
		"legacy install": {
			expected: true,
		},
		"vendor managed": {
			operatingModel: &InstallOperatingModel{ApprovalAuthority: InstallAuthorityVendor},
			expected:       true,
		},
		"connected customer managed": {
			operatingModel: &InstallOperatingModel{
				Connectivity:      InstallConnectivityConnected,
				ReleaseSelection:  InstallReleaseSelectionVendor,
				ApprovalAuthority: InstallAuthorityCustomer,
			},
			expected: true,
		},
		"offline customer managed": {
			operatingModel: &InstallOperatingModel{
				Connectivity:      "offline",
				ReleaseSelection:  InstallReleaseSelectionVendor,
				ApprovalAuthority: InstallAuthorityCustomer,
			},
		},
		"customer selected release": {
			operatingModel: &InstallOperatingModel{
				Connectivity:      InstallConnectivityConnected,
				ReleaseSelection:  InstallReleaseSelectionCustomer,
				ApprovalAuthority: InstallAuthorityCustomer,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			install := &Install{OperatingModel: test.operatingModel}
			require.Equal(t, test.expected, install.AppBranchUpdateEligible())
		})
	}
}
