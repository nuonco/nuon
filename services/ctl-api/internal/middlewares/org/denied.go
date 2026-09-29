package org

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func permissionDeniedError(acct *app.Account, orgID string, perm permissions.Permission, scope string) stderr.ErrAuthorization {
	level := "write"
	if perm == permissions.PermissionRead {
		level = "read"
	}

	desc := fmt.Sprintf("Ask an organization admin to assign you a role with %s access to %s.", level, scope)
	if titles := orgRoleTitles(acct, orgID); len(titles) > 0 {
		if len(titles) == 1 {
			desc = fmt.Sprintf("Your role (%s) does not have %s access to %s. Ask an organization admin to assign a role that does.",
				titles[0], level, scope)
		} else {
			desc = fmt.Sprintf("Your roles (%s) do not have %s access to %s. Ask an organization admin to assign a role that does.",
				strings.Join(titles, ", "), level, scope)
		}
	}

	return stderr.ErrAuthorization{
		Err:         fmt.Errorf("this action requires %s access to %s", level, scope),
		Description: desc,
	}
}

var scopeOverrides = map[string]string{
	"account/static-token":  "API tokens in this organization",
	"account/static-tokens": "API tokens in this organization",
	"account":               "your account",
	"orgs/webhooks":         "webhooks in this organization",
	"orgs/invites":          "team members in this organization",
	"orgs/accounts":         "team members in this organization",
	"orgs/user":             "team members in this organization",
	"orgs/remove-user":      "team members in this organization",
	"orgs/slack":            "Slack notifications in this organization",
	"orgs":                  "organization settings",
	"oidc":                  "OIDC federation in this organization",
	"vcs":                   "VCS connections in this organization",
}

func scopeFromPath(fullPath string) string {
	path, ok := strings.CutPrefix(fullPath, "/v1/")
	if !ok || path == "" {
		return "this organization"
	}

	segs := make([]string, 0, 2)
	for _, seg := range strings.Split(path, "/") {
		if seg == "" || seg == "current" || strings.HasPrefix(seg, ":") {
			continue
		}
		segs = append(segs, seg)
		if len(segs) == 2 {
			break
		}
	}
	if len(segs) == 0 {
		return "this organization"
	}
	if len(segs) == 2 {
		if label, ok := scopeOverrides[segs[0]+"/"+segs[1]]; ok {
			return label
		}
	}
	if label, ok := scopeOverrides[segs[0]]; ok {
		return label
	}
	return strings.ReplaceAll(segs[0], "-", " ") + " in this organization"
}

func orgRoleTitles(acct *app.Account, orgID string) []string {
	var titles []string
	for _, role := range acct.Roles {
		if role.Org == nil || role.Org.ID != orgID {
			continue
		}
		title := role.Title
		if title == "" {
			title = string(role.RoleType)
		}
		titles = append(titles, title)
	}
	return titles
}
