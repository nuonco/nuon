package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workloadjwt"
)

const (
	boundTenantID    = "11111111-2222-3333-4444-555555555555"
	boundSubID       = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	boundPrincipalID = "99999999-8888-7777-6666-555555555555"
	boundIdentity    = "inst123-phone-home"
)

func boundInstall() *app.Install {
	return &app.Install{
		ID: "inst123",
		CloudPlatformMetadata: app.CloudPlatformMetadata{
			TargetSubscriptionID: boundSubID,
		},
	}
}

func boundStackVersion() *app.InstallStackVersion {
	return &app.InstallStackVersion{PhoneHomeIdentityName: boundIdentity}
}

func boundVerifiedIdentity() *workloadjwt.AzureManagedIdentity {
	return &workloadjwt.AzureManagedIdentity{
		SubscriptionID: boundSubID,
		ResourceGroup:  "rg-1",
		Name:           boundIdentity,
		PrincipalID:    boundPrincipalID,
		TenantID:       boundTenantID,
	}
}

func TestBindAzureIdentity(t *testing.T) {
	for _, tc := range []struct {
		name       string
		install    func(*app.Install)
		version    func(*app.InstallStackVersion)
		identity   func(*workloadjwt.AzureManagedIdentity)
		wantReason string
	}{
		{
			name:       "the identity the template rendered is accepted",
			wantReason: phoneHomeAuthOK,
		},
		{
			name: "casing differences are not mismatches",
			identity: func(i *workloadjwt.AzureManagedIdentity) {
				i.SubscriptionID = "AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE"
				i.Name = "INST123-PHONE-HOME"
				i.TenantID = "11111111-2222-3333-4444-555555555555"
			},
			wantReason: phoneHomeAuthOK,
		},
		{
			name: "a token from another subscription is rejected",
			identity: func(i *workloadjwt.AzureManagedIdentity) {
				i.SubscriptionID = "ffffffff-bbbb-cccc-dddd-eeeeeeeeeeee"
			},
			wantReason: phoneHomeRejectAccountMismatch,
		},
		{
			name: "a neighbouring install's identity is rejected",
			identity: func(i *workloadjwt.AzureManagedIdentity) {
				i.Name = "inst999-phone-home"
			},
			wantReason: phoneHomeRejectIdentityMismatch,
		},
		{
			name:       "a version with no rendered identity cannot be satisfied",
			version:    func(v *app.InstallStackVersion) { v.PhoneHomeIdentityName = "" },
			identity:   func(i *workloadjwt.AzureManagedIdentity) { i.Name = "" },
			wantReason: phoneHomeRejectIdentityMismatch,
		},
		{
			name: "a token from another tenant is rejected",
			identity: func(i *workloadjwt.AzureManagedIdentity) {
				i.TenantID = "ffffffff-2222-3333-4444-555555555555"
			},
			wantReason: phoneHomeRejectTenantMismatch,
		},
		{
			name:       "an install with no target subscription is rejected",
			install:    func(i *app.Install) { i.CloudPlatformMetadata.TargetSubscriptionID = "" },
			wantReason: phoneHomeRejectAccountMismatch,
		},
		{
			name: "a principal other than the pinned one is rejected",
			install: func(i *app.Install) {
				i.CloudPlatformMetadata.ObservedPhoneHomePrincipalID = "00000000-8888-7777-6666-555555555555"
			},
			wantReason: phoneHomeRejectPrincipalPinned,
		},
		{
			name: "the pinned principal is accepted",
			install: func(i *app.Install) {
				i.CloudPlatformMetadata.ObservedPhoneHomePrincipalID = boundPrincipalID
			},
			wantReason: phoneHomeAuthOK,
		},
		{
			name:       "an unpinned install is accepted",
			install:    func(i *app.Install) { i.CloudPlatformMetadata.ObservedPhoneHomePrincipalID = "" },
			wantReason: phoneHomeAuthOK,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			install, version, identity := boundInstall(), boundStackVersion(), boundVerifiedIdentity()
			if tc.install != nil {
				tc.install(install)
			}
			if tc.version != nil {
				tc.version(version)
			}
			if tc.identity != nil {
				tc.identity(identity)
			}

			reason, err := bindAzureIdentity(install, version, identity, boundTenantID)

			assert.Equal(t, tc.wantReason, reason)
			if tc.wantReason == phoneHomeAuthOK {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)

			var authErr errPhoneHomeAuth
			assert.ErrorAs(t, err, &authErr)
			assert.Equal(t, tc.wantReason, authErr.reason)
		})
	}
}

func TestBindAzureIdentity_SubscriptionCheckedWithoutAPinnedTenant(t *testing.T) {
	identity := boundVerifiedIdentity()
	identity.SubscriptionID = "ffffffff-bbbb-cccc-dddd-eeeeeeeeeeee"

	reason, err := bindAzureIdentity(boundInstall(), boundStackVersion(), identity, identity.TenantID)

	assert.Equal(t, phoneHomeRejectAccountMismatch, reason)
	assert.Error(t, err)
}
