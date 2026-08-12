package roles

import (
	"fmt"
	"slices"
	"strings"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const PermissionFlagUsage = `Scoped permission entry, repeatable: verbs:resource_type:resource[:scope=app]
  verbs         comma-joined subset of create,read,update,delete, or "all"
  resource_type one of app, install, app_branch, org
  resource      the resource's id or name, or "*" for every resource of the type
  scope=app     confines a "*" entry to an app, by id or name (install and app_branch only)

  examples: read:app:app_web
            all:install:inl4plkdhwau58atwfd92vlc8q
            read,update:app_branch:*:scope=app_web`

var (
	knownVerbs         = []string{"create", "read", "update", "delete", "all"}
	knownResourceTypes = []string{"app", "install", "app_branch", "org"}
)

// ParsePermissionEntries parses repeated --permission flag values. It rejects
// what the API would reject so a malformed entry fails before the request
// rather than coming back as a 400 naming an entry index.
func ParsePermissionEntries(raw []string) ([]*models.ServicePermissionEntryRequest, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("at least one --permission is required")
	}

	entries := make([]*models.ServicePermissionEntryRequest, 0, len(raw))
	for _, val := range raw {
		entry, err := parsePermissionEntry(val)
		if err != nil {
			return nil, fmt.Errorf("invalid --permission %q: %w", val, err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func parsePermissionEntry(val string) (*models.ServicePermissionEntryRequest, error) {
	parts := strings.Split(val, ":")
	if len(parts) < 3 || len(parts) > 4 {
		return nil, fmt.Errorf("expected verbs:resource_type:resource[:scope=app]")
	}

	verbs, err := parseVerbs(parts[0])
	if err != nil {
		return nil, err
	}

	resourceType := parts[1]
	if !slices.Contains(knownResourceTypes, resourceType) {
		return nil, fmt.Errorf("unknown resource type %q: must be one of %s", resourceType, strings.Join(knownResourceTypes, ", "))
	}

	resource := parts[2]
	if resource == "" {
		return nil, fmt.Errorf("resource is required; use \"*\" for every %s", resourceType)
	}

	entry := &models.ServicePermissionEntryRequest{
		ResourceType: &resourceType,
		ResourceID:   &resource,
		Permissions:  verbs,
	}

	if len(parts) == 3 {
		return entry, nil
	}

	scope, found := strings.CutPrefix(parts[3], "scope=")
	if !found || scope == "" {
		return nil, fmt.Errorf("expected scope=<app id or name>, got %q", parts[3])
	}
	if resource != "*" {
		return nil, fmt.Errorf("scope only applies to wildcard (\"*\") entries")
	}
	// app is the only scope the API accepts, so the entry names the app
	// directly rather than repeating the scope type.
	if resourceType != "install" && resourceType != "app_branch" {
		return nil, fmt.Errorf("%s wildcards cannot be scoped to an app", resourceType)
	}

	entry.ScopeType = "app"
	entry.ScopeID = scope
	return entry, nil
}

func parseVerbs(val string) ([]string, error) {
	if val == "" {
		return nil, fmt.Errorf("at least one verb is required")
	}

	verbs := strings.Split(val, ",")
	for _, verb := range verbs {
		if !slices.Contains(knownVerbs, verb) {
			return nil, fmt.Errorf("unknown verb %q: must be one of %s", verb, strings.Join(knownVerbs, ", "))
		}
	}

	return verbs, nil
}
