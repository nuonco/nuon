package authz

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

// Link is one step of a resource's ownership chain, most-specific first; the
// org is always the last link.
type Link struct {
	Type app.Level
	ID   string
}

// Authorize walks a resource's ownership chain and allows the verb when any
// link carries a sufficient grant:
//
//   - an object grant on the link's id in perms (org-tier hstore permissions
//     land here too, which is how managed org-role holders keep working), or
//   - a type-wildcard grant for the link's type whose scope is empty
//     (org-wide) or names an ancestor link further up the chain.
//
// wildcards is the account's TypeGrants entry for the org the chain ends in.
func Authorize(perms permissions.Set, wildcards map[app.Level][]app.TypeGrant, chain []Link, verb permissions.Permission) error {
	for i, link := range chain {
		if err := perms.CanPerform(link.ID, verb); err == nil {
			return nil
		}

		for _, grant := range wildcards[link.Type] {
			if !grant.Verbs.Can(verb) {
				continue
			}
			if grant.ScopeID == "" {
				return nil
			}
			for _, ancestor := range chain[i+1:] {
				if ancestor.ID == grant.ScopeID {
					return nil
				}
			}
		}
	}

	objectID := ""
	if len(chain) > 0 {
		objectID = chain[0].ID
	}
	return permissions.NoAccessError{
		Permission: verb,
		ObjectID:   objectID,
	}
}
