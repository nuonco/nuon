package org

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/global"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/public"
)

// orgLevelRoutes lists routes that are correctly gated at the org tier:
// grant-scoped accounts have no business here (org settings, team, account
// and credential management, and the grant surface itself), so failing closed
// below an org-wide permission is intentional, not a coverage gap.
var orgLevelRoutes = map[string]struct{}{
	"POST /v1/account/static-token":                    {},
	"GET /v1/account/static-tokens":                    {},
	"DELETE /v1/account/static-tokens/:token_id":       {},
	"GET /v1/orgs/current":                             {},
	"PATCH /v1/orgs/current":                           {},
	"DELETE /v1/orgs/current":                          {},
	"GET /v1/orgs/current/accounts":                    {},
	"PATCH /v1/orgs/current/accounts/:account_id/role": {},
	"GET /v1/orgs/current/features":                    {},
	"PATCH /v1/orgs/current/features":                  {},
	"GET /v1/orgs/current/invites":                     {},
	"POST /v1/orgs/current/invites":                    {},
	"POST /v1/orgs/current/invites/:invite_id/resend":  {},
	"POST /v1/orgs/current/invites/:invite_id/revoke":  {},
	"POST /v1/orgs/current/remove-user":                {},
	"GET /v1/orgs/current/runner-group":                {},
	"GET /v1/orgs/current/stats":                       {},
	"POST /v1/orgs/current/user":                       {},
	"GET /v1/orgs/features":                            {},
	"POST /v1/grants":                                  {},
	"GET /v1/grants":                                   {},
	"DELETE /v1/grants/:grant_id":                      {},
	"GET /v1/roles":                                    {},
	"GET /v1/service-accounts":                         {},
	"POST /v1/service-accounts":                        {},
	"PATCH /v1/service-accounts/:account_id":           {},
	"DELETE /v1/service-accounts/:account_id":          {},
	"PATCH /v1/service-accounts/:account_id/role":      {},
	"POST /v1/service-accounts/:account_id/tokens":     {},
	"POST /v1/oidc/trust-policies":                     {},
	"GET /v1/oidc/trust-policies":                      {},
	"GET /v1/oidc/trust-policies/:policy_id":           {},
	"PATCH /v1/oidc/trust-policies/:policy_id":         {},
	"DELETE /v1/oidc/trust-policies/:policy_id":        {},

	// deliberately not a grantable resource type
	"POST /v1/aws-account-connections":                       {},
	"GET /v1/aws-account-connections":                        {},
	"GET /v1/aws-account-connections/:connection_id":         {},
	"PATCH /v1/aws-account-connections/:connection_id":       {},
	"DELETE /v1/aws-account-connections/:connection_id":      {},
	"POST /v1/aws-account-connections/:connection_id/verify": {},

	// post-org onboarding steps (pre-org steps are global)
	"POST /v1/onboarding/current/steps/deploy":      {},
	"POST /v1/onboarding/current/steps/get-started": {},
	"POST /v1/onboarding/current/steps/install":     {},
	"POST /v1/onboarding/current/steps/your-stack":  {},
}

// uncoveredRoutes lists accepted coverage gaps: routes a grant semantically
// covers but that cannot yet resolve their resource. Grant-scoped accounts
// fail closed (403) here until an entry graduates into a resolver or a
// filtered collection. Grouped by why they are uncovered.
var uncoveredRoutes = map[string]struct{}{
	// org-wide collections/aggregates pending handler-level scope filtering
	"GET /v1/builds":                      {},
	"GET /v1/component-builds":            {},
	"GET /v1/installs/health":             {},
	"GET /v1/installs/label-keys":         {},
	"GET /v1/workflows":                   {},
	"GET /v1/workflows/pending-approvals": {},
	"GET /v1/policy-reports":              {},
	"GET /v1/queues":                      {},
	"GET /v1/runner-jobs":                 {},
	"GET /v1/terraform-workspaces":        {},
	"GET /v1/triggers":                    {},
	"GET /v1/triggers/dispatches":         {},

	// resource IDs carried in the request body, not the path
	"POST /v1/workflows/cancel": {},

	// runner / execution infrastructure: owner resolution (runner -> install/org)
	// not yet defined for grants
	"GET /v1/runners/:runner_id":                                          {},
	"GET /v1/runners/:runner_id/card-details":                             {},
	"GET /v1/runners/:runner_id/connected":                                {},
	"POST /v1/runners/:runner_id/force-shutdown":                          {},
	"POST /v1/runners/:runner_id/graceful-shutdown":                       {},
	"GET /v1/runners/:runner_id/heart-beats/latest":                       {},
	"GET /v1/runners/:runner_id/jobs":                                     {},
	"GET /v1/runners/:runner_id/latest-heart-beat":                        {},
	"POST /v1/runners/:runner_id/mng/restart":                             {},
	"POST /v1/runners/:runner_id/mng/shutdown":                            {},
	"POST /v1/runners/:runner_id/mng/shutdown-vm":                         {},
	"POST /v1/runners/:runner_id/mng/update":                              {},
	"GET /v1/runners/:runner_id/processes":                                {},
	"GET /v1/runners/:runner_id/processes/current":                        {},
	"GET /v1/runners/:runner_id/processes/:process_id":                    {},
	"GET /v1/runners/:runner_id/processes/:process_id/heart-beats/latest": {},
	"POST /v1/runners/:runner_id/processes/:process_id/shutdown":          {},
	"POST /v1/runners/:runner_id/prune-tokens":                            {},
	"GET /v1/runners/:runner_id/recent-health-checks":                     {},
	"GET /v1/runners/:runner_id/settings":                                 {},
	"PATCH /v1/runners/:runner_id/settings":                               {},
	"GET /v1/runner-jobs/:runner_job_id":                                  {},
	"POST /v1/runner-jobs/:runner_job_id/cancel":                          {},
	"GET /v1/runner-jobs/:runner_job_id/composite-plan":                   {},
	"GET /v1/runner-jobs/:runner_job_id/plan":                             {},
	"GET /v1/log-streams/:log_stream_id":                                  {},
	"GET /v1/log-streams/:log_stream_id/logs":                             {},
	"GET /v1/log-streams/:log_stream_id/logs/tail":                        {},
	"GET /v1/log-streams/:log_stream_id/spans":                            {},

	// terraform state surface (runner-adjacent)
	"GET /v1/runners/terraform-workspace/:workspace_id/state-json":                     {},
	"GET /v1/runners/terraform-workspace/:workspace_id/state-json/:state_id":           {},
	"GET /v1/runners/terraform-workspace/:workspace_id/state-json/:state_id/raw":       {},
	"GET /v1/runners/terraform-workspace/:workspace_id/state-json/:state_id/resources": {},
	"GET /v1/runners/terraform-workspace/:workspace_id/states":                         {},
	"GET /v1/runners/terraform-workspace/:workspace_id/states/:state_id":               {},
	"GET /v1/runners/terraform-workspace/:workspace_id/states/:state_id/resources":     {},
	"GET /v1/terraform-backend":                                                 {},
	"POST /v1/terraform-backend":                                                {},
	"POST /v1/terraform-workspace":                                              {},
	"POST /v1/terraform-workspaces":                                             {},
	"GET /v1/terraform-workspaces/:workspace_id":                                {},
	"DELETE /v1/terraform-workspaces/:workspace_id":                             {},
	"GET /v1/terraform-workspaces/:workspace_id/lock":                           {},
	"POST /v1/terraform-workspaces/:workspace_id/lock":                          {},
	"POST /v1/terraform-workspaces/:workspace_id/unlock":                        {},
	"GET /v1/terraform-workspaces/:workspace_id/state-json":                     {},
	"GET /v1/terraform-workspaces/:workspace_id/state-json/:state_id":           {},
	"GET /v1/terraform-workspaces/:workspace_id/state-json/:state_id/raw":       {},
	"GET /v1/terraform-workspaces/:workspace_id/state-json/:state_id/resources": {},
	"GET /v1/terraform-workspaces/:workspace_id/states":                         {},
	"GET /v1/terraform-workspaces/:workspace_id/states/:state_id":               {},
	"GET /v1/terraform-workspaces/:workspace_id/states/:state_id/resources":     {},

	// queue introspection: owner resolution not yet defined for grants
	"GET /v1/queues/:queue_id":                          {},
	"GET /v1/queues/:queue_id/signals":                  {},
	"GET /v1/queues/:queue_id/signals/:signal_id":       {},
	"GET /v1/queues/:queue_id/signals/:signal_id/await": {},
	"GET /v1/queues/:queue_id/signals/:signal_id/graph": {},
	"GET /v1/queues/:queue_id/status":                   {},

	// policy reports: owner resolution not yet defined for grants
	"GET /v1/policy-reports/:report_id":        {},
	"GET /v1/policy-reports/:report_id/export": {},

	// triggers: owner resolution not yet defined for grants
	"POST /v1/triggers":                                        {},
	"GET /v1/triggers/dispatches/:dispatch_id":                 {},
	"POST /v1/triggers/dispatches/:dispatch_id/retry":          {},
	"GET /v1/triggers/events/:event_id":                        {},
	"GET /v1/triggers/events/:event_id/raw":                    {},
	"POST /v1/triggers/events/:event_id/replay":                {},
	"GET /v1/triggers/:trigger_id":                             {},
	"DELETE /v1/triggers/:trigger_id":                          {},
	"POST /v1/triggers/:trigger_id/disable":                    {},
	"POST /v1/triggers/:trigger_id/enable":                     {},
	"GET /v1/triggers/:trigger_id/event-types":                 {},
	"GET /v1/triggers/:trigger_id/events":                      {},
	"PATCH /v1/triggers/:trigger_id/ingress-url":               {},
	"POST /v1/triggers/:trigger_id/rotate-ingress-url":         {},
	"POST /v1/triggers/:trigger_id/rotate-secret":              {},
	"GET /v1/triggers/:trigger_id/rules":                       {},
	"GET /v1/triggers/:trigger_id/rules/:rule_id":              {},
	"PATCH /v1/triggers/:trigger_id/secrets/:secret_id/reveal": {},
	"POST /v1/triggers/:trigger_id/secrets/:secret_id/revoke":  {},
}

func routeKey(method, path string) string {
	return method + " " + path
}

// ValidateRouteCoverage asserts every org-scoped route is consciously
// classified: resolvable by the chain builders, exempt (global / public /
// non-v1), explicitly org-level, or explicitly uncovered. It runs at startup
// so an unclassified route fails the boot (and CI) of the PR that adds it
// instead of silently 403ing grant-scoped accounts in production.
func ValidateRouteCoverage(routes gin.RoutesInfo) error {
	var unclassified []string
	for _, r := range routes {
		if isClassifiedRoute(r.Method, r.Path) {
			continue
		}
		unclassified = append(unclassified, routeKey(r.Method, r.Path))
	}
	if len(unclassified) == 0 {
		return nil
	}

	sort.Strings(unclassified)
	return fmt.Errorf(
		"%d route(s) lack a resource-grant coverage classification; add a resolver (owners.go / resources.go), or list them in orgLevelRoutes or uncoveredRoutes (middlewares/org/coverage.go):\n  %s",
		len(unclassified), strings.Join(unclassified, "\n  "))
}

func isClassifiedRoute(method, path string) bool {
	if !strings.HasPrefix(path, "/v1/") {
		return true
	}
	if global.IsGlobalEndpoint(method, path) || public.IsPublicEndpoint(method, path) {
		return true
	}
	if strings.Contains(path, ":install_id") || strings.Contains(path, ":app_id") {
		return true
	}
	if _, ok := matchOrgResourceRoute(path); ok {
		return true
	}
	// owned routes only count when the path actually carries the resolving
	// param — a prefix match alone would classify collection/aggregate routes
	// that the resolver cannot actually cover.
	if r, ok := matchOwnedResourceRoute(path); ok && strings.Contains(path, ":"+r.idParam) {
		return true
	}
	if _, ok := matchTypeOnlyCreateRoute(method, path); ok {
		return true
	}
	if isFilteredCollectionRoute(method, path) {
		return true
	}
	if _, ok := orgLevelRoutes[routeKey(method, path)]; ok {
		return true
	}
	if _, ok := uncoveredRoutes[routeKey(method, path)]; ok {
		return true
	}
	return false
}
