package org

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestMatchOrgResourceRoute(t *testing.T) {
	tests := []struct {
		fullPath string
		wantType app.GrantResourceType
		wantOK   bool
	}{
		{"/v1/orgs/current/webhooks", app.GrantResourceTypeWebhook, true},
		{"/v1/orgs/current/webhooks/:webhook_id", app.GrantResourceTypeWebhook, true},
		{"/v1/vcs/connections", app.GrantResourceTypeVCSConnection, true},
		{"/v1/vcs/connections/:connection_id/repos", app.GrantResourceTypeVCSConnection, true},
		{"/v1/orgs/:org_id/slack/channel-subscriptions/:sub_id", app.GrantResourceTypeSlackSubscription, true},
		{"/v1/orgs/:org_id/slack/install-url", app.GrantResourceTypeSlackSubscription, true},

		{"/v1/aws-account-connections/:connection_id", "", false},
		{"/v1/orgs/current/invites", "", false},
		{"/v1/vcs/connection-callback", "", false},
		{"/v1/installs/:install_id", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.fullPath, func(t *testing.T) {
			route, ok := matchOrgResourceRoute(tt.fullPath)
			if ok != tt.wantOK {
				t.Fatalf("matched=%v, want %v", ok, tt.wantOK)
			}
			if ok && route.resourceType != tt.wantType {
				t.Fatalf("resource type=%q, want %q", route.resourceType, tt.wantType)
			}
		})
	}
}
