package org

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

// chainParams reads a route's path params by name, returning "" when the
// matched route does not declare one.
type chainParams func(name string) string

// chainFromParams builds a resource's ownership chain from the route's path
// params alone — no database read. Links are ordered most-specific first and
// the org is always last, which is what makes an org-tier grant authorize
// every route (see Authorize).
//
// A tier named by the route but absent from the URL contributes an empty-id
// link: an install reached through /v1/installs/:install_id sits under some
// app, but which app is not knowable without a lookup. An empty-id link
// carries no object grant, so it only lets an org-wide wildcard for that tier
// authorize. A wildcard scoped to a specific app cannot match it, which is why
// app-scoped grants only take effect through genuinely app-nested URLs.
func chainFromParams(param chainParams, orgID string) []authz.Link {
	chain := make([]authz.Link, 0, 4)

	installID := param("install_id")
	appBranchID := param("app_branch_id")
	appID := param("app_id")

	if installID != "" {
		chain = append(chain, authz.Link{Type: app.LevelInstall, ID: installID})
	}
	if appBranchID != "" {
		chain = append(chain, authz.Link{Type: app.LevelAppBranch, ID: appBranchID})
	}
	if appID != "" || installID != "" || appBranchID != "" {
		chain = append(chain, authz.Link{Type: app.LevelApp, ID: appID})
	}

	// appended unconditionally, outside the tier checks above, so no change to
	// them can drop the org link and deny every request at once.
	return append(chain, authz.Link{Type: app.LevelOrg, ID: orgID})
}
