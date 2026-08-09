package org

import (
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// orgResourceRoute maps a route-prefix (matched against the gin FullPath, i.e.
// the registered pattern) to the org-owned resource type its routes operate on
// and the route param naming a specific entity. Routes under the prefix without
// that param (creates, collection reads) resolve to a type-only link, which
// only a wildcard grant on the type — or an org-wide grant — satisfies.
//
// Prefixes are matched in order; keying on the prefix rather than the bare
// param avoids collisions like :connection_id, which both VCS connections and
// AWS account connections use.
type orgResourceRoute struct {
	prefix       string
	resourceType app.GrantResourceType
	idParam      string
	model        func() any
}

var orgResourceRoutes = []orgResourceRoute{
	{
		prefix:       "/v1/orgs/current/webhooks",
		resourceType: app.GrantResourceTypeWebhook,
		idParam:      "webhook_id",
		model:        func() any { return &app.Webhook{} },
	},
	{
		prefix:       "/v1/vcs/connections",
		resourceType: app.GrantResourceTypeVCSConnection,
		idParam:      "connection_id",
		model:        func() any { return &app.VCSConnection{} },
	},
	// The slack group also contains installation/org-link/channel routes with no
	// :sub_id; those resolve type-only, so a slack_subscription wildcard covers
	// the whole surface while a single-subscription grant covers only its own
	// subscription routes.
	{
		prefix:       "/v1/orgs/:org_id/slack",
		resourceType: app.GrantResourceTypeSlackSubscription,
		idParam:      "sub_id",
		model:        func() any { return &app.SlackChannelSubscription{} },
	},
}

func matchOrgResourceRoute(fullPath string) (orgResourceRoute, bool) {
	for _, r := range orgResourceRoutes {
		if strings.HasPrefix(fullPath, r.prefix) {
			return r, true
		}
	}
	return orgResourceRoute{}, false
}
