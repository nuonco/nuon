package service

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// TestCanLinkWorkspace guards the authorization rule for binding a Slack
// workspace to an org: only the account that installed the app may link it.
// Slack team IDs are not secret, so a looser rule allows cross-tenant hijack
// via a guessed team_id.
func TestCanLinkWorkspace(t *testing.T) {
	cases := []struct {
		name      string
		installer string
		acctID    string
		want      bool
	}{
		{"installer matches", "acc_installer", "acc_installer", true},
		{"different account rejected", "acc_installer", "acc_attacker", false},
		{"empty caller rejected", "acc_installer", "", false},
		{"empty installer never matches", "", "acc_attacker", false},
		{"both empty rejected", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			install := app.SlackInstallation{InstalledByAccountID: tc.installer}
			if got := canLinkWorkspace(install, tc.acctID); got != tc.want {
				t.Errorf("canLinkWorkspace(installer=%q, acct=%q) = %v, want %v",
					tc.installer, tc.acctID, got, tc.want)
			}
		})
	}
}
