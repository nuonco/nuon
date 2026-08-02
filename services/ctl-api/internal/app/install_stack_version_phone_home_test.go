package app

import (
	"testing"
	"time"
)

// Eligibility is the reconciler's desired-state rule, so every status needs a
// verdict here — a new status silently defaulting to "no credential" (or worse, to
// "mint one") is exactly the kind of drift this pins down.
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

		// Retired: the handler rejects an expired version outright, and an outdated
		// one has been superseded by a version that already phoned home.
		"outdated":  {status: InstallStackVersionStatusOutdated, want: false},
		"expired":   {status: InstallStackVersionStatusExpired, want: false},
		"cancelled": {status: StatusCancelled, want: false},

		// The tombstone outranks status. Without this, revocation and
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

// The Go rule and the SQL pre-filter must agree, or the reconciler loads one set of
// versions and reasons about another.
func TestPhoneHomeTokenEligibleStatusesMatchesPredicate(t *testing.T) {
	for _, status := range PhoneHomeTokenEligibleStatuses {
		version := &InstallStackVersion{Status: CompositeStatus{Status: status}}
		if !version.PhoneHomeTokenEligible() {
			t.Errorf("%q is in PhoneHomeTokenEligibleStatuses but PhoneHomeTokenEligible() is false", status)
		}
	}
}
