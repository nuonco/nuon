package authz

import (
	"strings"

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
	_, err := Decide(perms, wildcards, chain, verb)
	return err
}

// Decide is Authorize with the reason it allowed the verb, for operator
// logging: "grant:install:inl_…", "wildcard:app", or "none" on denial.
func Decide(perms permissions.Set, wildcards map[app.Level][]app.TypeGrant, chain []Link, verb permissions.Permission) (string, error) {
	for i, link := range chain {
		// An empty id names a tier the URL did not identify, so there is no
		// object to hold a grant. Skipping the check also keeps a "*" key in a
		// permission set from authorizing through such a link, since
		// CanPerform treats "*" as matching any object.
		if link.ID != "" {
			if err := perms.CanPerform(link.ID, verb); err == nil {
				return "grant:" + string(link.Type) + ":" + link.ID, nil
			}
		}

		for _, grant := range wildcards[link.Type] {
			if !grant.Verbs.Can(verb) {
				continue
			}
			if grant.ScopeID == "" {
				return "wildcard:" + string(link.Type), nil
			}
			for _, ancestor := range chain[i+1:] {
				if ancestor.ID == grant.ScopeID {
					return "wildcard:" + string(link.Type) + ":scoped:" + grant.ScopeID, nil
				}
			}
		}
	}

	objectID := ""
	if len(chain) > 0 {
		objectID = chain[0].ID
	}
	return "none", permissions.NoAccessError{
		Permission: verb,
		ObjectID:   objectID,
	}
}

// String renders a chain for logs: install:inl_… > app:* > org:org_…
func ChainString(chain []Link) string {
	parts := make([]string, 0, len(chain))
	for _, link := range chain {
		id := link.ID
		if id == "" {
			id = "*"
		}
		parts = append(parts, string(link.Type)+":"+id)
	}
	return strings.Join(parts, " > ")
}
