// Package scope derives list-query filters from an account's resource grants.
//
// The walk-up authorization primitive answers "may this account touch this one
// resource?" — but list endpoints return whole collections and can't ask that
// per row via middleware. This package turns an account's grants into per-tier
// id sets and GORM scopes that filter a collection down to the rows the account
// is entitled to see. Callers with an org-wide grant skip this entirely (the
// org gate authorizes them up front); only narrow grantees are filtered.
package scope

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

// IDSets holds the resource ids an account is granted, bucketed by tier, for a
// single requested permission. The *WildcardOrgs sets carry the orgs in which
// the account holds a type-wildcard grant (all apps / all installs in that org).
type IDSets struct {
	OrgIDs     []string
	AppIDs     []string
	InstallIDs []string

	AppWildcardOrgs     []string
	InstallWildcardOrgs []string
}

// satisfies reports whether a grant's permission covers the requested one.
func satisfies(granted, want permissions.Permission) bool {
	return granted == permissions.PermissionAll || granted == want
}

// ForList collects, per tier, the resources the account is granted for perm.
func ForList(acct *app.Account, perm permissions.Permission) IDSets {
	var s IDSets
	for _, g := range acct.Grants {
		granted, err := permissions.NewPermission(g.Permission)
		if err != nil || !satisfies(granted, perm) {
			continue
		}

		if g.IsWildcard() {
			switch g.ResourceType {
			case app.GrantResourceTypeApp:
				s.AppWildcardOrgs = append(s.AppWildcardOrgs, g.OrgID)
			case app.GrantResourceTypeInstall:
				s.InstallWildcardOrgs = append(s.InstallWildcardOrgs, g.OrgID)
			}
			continue
		}

		switch g.ResourceType {
		case app.GrantResourceTypeOrg:
			s.OrgIDs = append(s.OrgIDs, g.ResourceID)
		case app.GrantResourceTypeApp:
			s.AppIDs = append(s.AppIDs, g.ResourceID)
		case app.GrantResourceTypeInstall:
			s.InstallIDs = append(s.InstallIDs, g.ResourceID)
		}
	}
	return s
}

// Installs scopes an installs query to rows visible via any granted tier (the
// install itself, its parent app, or its org). Column names are passed in so
// callers querying through a view supply the correct aliases. Empty id slices
// compile to IN (NULL) and match nothing, so an account with no relevant grant
// gets an empty result rather than an error.
func (s IDSets) Installs(db *gorm.DB, idCol, appCol, orgCol string) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(
			db.Where(idCol+" IN ?", s.InstallIDs).
				Or(appCol+" IN ?", s.AppIDs).
				Or(orgCol+" IN ?", s.OrgIDs).
				// all installs in the org (install wildcard), and all installs
				// under all apps in the org (app wildcard cascades to installs)
				Or(orgCol+" IN ?", s.InstallWildcardOrgs).
				Or(orgCol+" IN ?", s.AppWildcardOrgs),
		)
	}
}

// Apps scopes an apps query to granted apps and org, plus the parent app of any
// granted install (upward visibility: a narrow grantee can see where their
// install lives; that app's other installs stay filtered by Installs).
func (s IDSets) Apps(db *gorm.DB, idCol, orgCol string) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(
			db.Where(idCol+" IN ?", s.AppIDs).
				Or(orgCol+" IN ?", s.OrgIDs).
				Or(idCol+" IN (?)",
					db.Model(&app.Install{}).
						Select("app_id").
						Where("id IN ?", s.InstallIDs)).
				// all apps in the org (app wildcard)
				Or(orgCol+" IN ?", s.AppWildcardOrgs).
				// upward visibility: parent apps of every install in an
				// install-wildcard org
				Or(idCol+" IN (?)",
					db.Model(&app.Install{}).
						Select("app_id").
						Where("org_id IN ?", s.InstallWildcardOrgs)),
		)
	}
}

// Components scopes a components query. Since component is not yet a grant
// target, visibility is only via the parent app or the org (an install-only
// grantee sees no components in the org-wide list — the install's components
// are reachable under /v1/installs/:id/components instead).
func (s IDSets) Components(db *gorm.DB, appCol, orgCol string) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(
			db.Where(appCol+" IN ?", s.AppIDs).
				Or(orgCol+" IN ?", s.OrgIDs).
				// all apps in the org (app wildcard) — an install wildcard does
				// not confer component visibility, matching install grants
				Or(orgCol+" IN ?", s.AppWildcardOrgs),
		)
	}
}
