package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

const maxPermissionEntries = 100

type PermissionEntryRequest struct {
	// ResourceType is the kind of resource the entry targets: app, install, app_branch, or org.
	ResourceType string `json:"resource_type" binding:"required"`
	// ResourceID is the resource's id or name, or "*" for every resource of the type.
	ResourceID string `json:"resource_id" binding:"required"`
	// ScopeType optionally confines a "*" entry to a parent kind (currently app).
	ScopeType string `json:"scope_type"`
	// ScopeID is the parent's id or name, of ScopeType.
	ScopeID string `json:"scope_id"`
	// Permissions is a subset of create, read, update, delete; "all" grants all four.
	Permissions []string `json:"permissions" binding:"required"`
}

// validatePermissionEntries resolves and validates requested entries into
// canonical PermissionEntry values: names become ids, every id is confined to
// the org, wildcard scopes must be legal parent tiers.
func (s *service) validatePermissionEntries(ctx *gin.Context, org *app.Org, entries []PermissionEntryRequest) ([]app.PermissionEntry, error) {
	if len(entries) == 0 {
		return nil, userErr(fmt.Errorf("at least one permission entry is required"))
	}
	if len(entries) > maxPermissionEntries {
		return nil, userErr(fmt.Errorf("a role may carry at most %d permission entries", maxPermissionEntries))
	}

	out := make([]app.PermissionEntry, 0, len(entries))
	for i, entry := range entries {
		resolved, err := s.validatePermissionEntry(ctx, org, entry)
		if err != nil {
			return nil, userErr(fmt.Errorf("permission entry %d: %w", i, err))
		}
		out = append(out, *resolved)
	}

	return out, nil
}

func (s *service) validatePermissionEntry(ctx *gin.Context, org *app.Org, entry PermissionEntryRequest) (*app.PermissionEntry, error) {
	level, err := app.NewLevel(entry.ResourceType)
	if err != nil {
		return nil, err
	}

	verbs, err := permissions.NewVerbs(entry.Permissions)
	if err != nil {
		return nil, err
	}

	if entry.ResourceID != "*" {
		if entry.ScopeType != "" || entry.ScopeID != "" {
			return nil, fmt.Errorf("scope only applies to wildcard (\"*\") entries")
		}

		resourceID, err := s.resolveResource(ctx, org, level, entry.ResourceID)
		if err != nil {
			return nil, err
		}

		return &app.PermissionEntry{
			ResourceType: level,
			ResourceID:   resourceID,
			Permissions:  verbs,
		}, nil
	}

	if level == app.LevelOrg {
		return nil, fmt.Errorf("wildcard entries on orgs are not supported")
	}
	if (entry.ScopeType == "") != (entry.ScopeID == "") {
		return nil, fmt.Errorf("scope_type and scope_id must be set together")
	}

	resolved := &app.PermissionEntry{
		ResourceType: level,
		ResourceID:   "*",
		Permissions:  verbs,
	}
	if entry.ScopeType == "" {
		return resolved, nil
	}

	scope, err := app.NewLevel(entry.ScopeType)
	if err != nil {
		return nil, err
	}
	if !level.ValidWildcardScope(scope) {
		return nil, fmt.Errorf("%s is not a valid scope for %s wildcards", scope, level)
	}

	scopeID, err := s.resolveResource(ctx, org, scope, entry.ScopeID)
	if err != nil {
		return nil, err
	}

	resolved.ScopeType = scope
	resolved.ScopeID = scopeID
	return resolved, nil
}

// resolveResource turns an id-or-name into the resource's canonical id,
// confined to the org. Names that are not unique within the org (installs and
// branches are only unique per app) are rejected as ambiguous.
func (s *service) resolveResource(ctx *gin.Context, org *app.Org, level app.Level, identifier string) (string, error) {
	if level == app.LevelOrg {
		if identifier != org.ID {
			return "", fmt.Errorf("org entries must name the current org")
		}
		return org.ID, nil
	}

	var model any
	switch level {
	case app.LevelApp:
		model = &app.App{}
	case app.LevelInstall:
		model = &app.Install{}
	case app.LevelAppBranch:
		model = &app.AppBranch{}
	default:
		return "", fmt.Errorf("unsupported resource type %q", level)
	}

	var ids []string
	res := s.db.WithContext(ctx).
		Model(model).
		Where("org_id = ?", org.ID).
		Where(s.db.Where("id = ?", identifier).Or("name = ?", identifier)).
		Limit(2).
		Pluck("id", &ids)
	if res.Error != nil {
		return "", fmt.Errorf("unable to resolve %s %q: %w", level, identifier, res.Error)
	}

	switch len(ids) {
	case 0:
		return "", fmt.Errorf("%s %q not found", level, identifier)
	case 1:
		return ids[0], nil
	}

	for _, id := range ids {
		if id == identifier {
			return id, nil
		}
	}
	return "", fmt.Errorf("%s name %q is ambiguous; use its id", level, identifier)
}

func userErr(err error) error {
	return stderr.ErrUser{
		Err:         err,
		Description: err.Error(),
	}
}
