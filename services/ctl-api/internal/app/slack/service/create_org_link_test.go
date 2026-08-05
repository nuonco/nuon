package service

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Only the installer account may bind a workspace (team_ids aren't secret).
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

// Only the owner org passes the channel/subscription gate; non-owner links must not.
func TestOwnsWorkspace(t *testing.T) {
	cases := []struct {
		name    string
		ownerID string
		orgID   string
		want    bool
	}{
		{"owner matches", "org_owner", "org_owner", true},
		{"non-owner rejected", "org_owner", "org_guest", false},
		{"empty caller rejected", "org_owner", "", false},
		{"unowned install never matches", "", "org_guest", false},
		{"both empty rejected", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			install := app.SlackInstallation{OwnerOrgID: tc.ownerID}
			if got := ownsWorkspace(install, tc.orgID); got != tc.want {
				t.Errorf("ownsWorkspace(owner=%q, org=%q) = %v, want %v",
					tc.ownerID, tc.orgID, got, tc.want)
			}
		})
	}
}
