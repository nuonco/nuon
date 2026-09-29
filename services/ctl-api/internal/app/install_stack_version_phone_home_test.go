package app

import (
	"testing"
	"time"
)

func TestInstallStackVersionPhoneHomeTokenEligible(t *testing.T) {
	revoked := time.Now()

	for name, tc := range map[string]struct {
		status    Status
		revokedAt *time.Time
		want      bool
	}{
		"generating":        {status: InstallStackVersionStatusGenerating, want: true},
		"awaiting user run": {status: InstallStackVersionStatusPendingUser, want: true},
		"provisioning":      {status: InstallStackVersionStatusProvisioning, want: true},
		"active":            {status: InstallStackVersionStatusActive, want: true},

		"outdated":  {status: InstallStackVersionStatusOutdated, want: false},
		"expired":   {status: InstallStackVersionStatusExpired, want: false},
		"cancelled": {status: StatusCancelled, want: false},

		// why: The tombstone outranks status. Without this, revocation and
		// never-minted are indistinguishable and the reconciler resurrects a
		// credential it was just told to kill.
		"active but revoked": {
			status:    InstallStackVersionStatusActive,
			revokedAt: &revoked,
			want:      false,
		},
		"generating but revoked": {
			status:    InstallStackVersionStatusGenerating,
			revokedAt: &revoked,
			want:      false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			version := &InstallStackVersion{
				Status:                  CompositeStatus{Status: tc.status},
				PhoneHomeTokenRevokedAt: tc.revokedAt,
			}

			if got := version.PhoneHomeTokenEligible(); got != tc.want {
				t.Errorf("PhoneHomeTokenEligible() = %v, want %v (status %q)", got, tc.want, tc.status)
			}
		})
	}
}

// why: The Go rule and the SQL pre-filter must agree, or the reconciler loads one set of
// versions and reasons about another.
func TestPhoneHomeTokenEligibleStatusesMatchesPredicate(t *testing.T) {
	for _, status := range PhoneHomeTokenEligibleStatuses {
		version := &InstallStackVersion{Status: CompositeStatus{Status: status}}
		if !version.PhoneHomeTokenEligible() {
			t.Errorf("%q is in PhoneHomeTokenEligibleStatuses but PhoneHomeTokenEligible() is false", status)
		}
	}
}
