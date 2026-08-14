// Package azureroles resolves Azure built-in role names to their definition
// GUIDs. ARM role assignments reference a definition by GUID and have no name
// lookup, so anything an app config expresses as a name has to be translated
// before it reaches a template.
package azureroles

import (
	"regexp"
	"sort"
)

// builtInRoleGUIDs maps built-in role names to their definition GUIDs. These are
// stable, well-known values assigned by Azure.
var builtInRoleGUIDs = map[string]string{
	"Owner":                     "8e3af657-a8ff-443c-a75c-2fe8c4bcb635",
	"Contributor":               "b24988ac-6180-42a0-ab88-20f7382dd24c",
	"Reader":                    "acdd72a7-3385-48ef-bd42-f606fba81ae7",
	"User Access Administrator": "18d7d88d-d35e-4fb5-a5c3-7773c20a72d9",
	"Role Based Access Control Administrator":     "f58310d9-a9f6-439a-9e8d-f62e7b41a168",
	"Azure Kubernetes Service RBAC Cluster Admin": "b1ff04bb-8a4e-4dc4-8eb5-8693973ce19b",
	"Azure Kubernetes Service RBAC Admin":         "3498e952-d568-435e-9b2c-8d77e338d7f7",
	"Azure Kubernetes Service RBAC Writer":        "a7ffa36f-339b-4b5c-8bdf-e2c188b2c0eb",
	"Azure Kubernetes Service RBAC Reader":        "7f6c6a51-bcf8-42ba-9220-52d62157d7db",
	"Azure Kubernetes Service Cluster Admin Role": "0ab0b1a8-8aac-4efd-b8c2-3ee1fb270be8",
	"Azure Kubernetes Service Cluster User Role":  "4abbcc35-e782-43d8-92c5-2d3f1bd2253f",
	"Key Vault Administrator":                     "00482a5a-887f-4fb3-b363-3b7fe8e74483",
	"Key Vault Secrets User":                      "4633458b-17de-408a-b874-0445c86b69e6",
	"Key Vault Secrets Officer":                   "b86a8fe4-44ce-4948-aee5-eccb2c155cd7",
	"Storage Blob Data Contributor":               "ba92f5b4-2d11-453d-a403-e96b0029c9fe",
	"Storage Blob Data Owner":                     "b7e6dc6d-f1e8-4753-8033-0f276bb0955b",
	"Network Contributor":                         "4d97b98b-1d4f-4787-a291-c67834d212e7",
}

var guidRegexp = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// GUID resolves a role name to its definition GUID. A value that is already a
// GUID is returned unchanged, so a config can name any role -- including one
// absent from the map above -- by passing its GUID directly.
func GUID(nameOrGUID string) string {
	if guid, ok := builtInRoleGUIDs[nameOrGUID]; ok {
		return guid
	}
	return nameOrGUID
}

// Resolvable reports whether GUID will produce something ARM can accept: either
// a mapped name or a literal GUID. Anything else -- a typo, or a real role name
// missing from the map -- would otherwise be forwarded to ARM verbatim and fail
// the customer's stack deployment with InvalidRoleDefinitionId.
func Resolvable(nameOrGUID string) bool {
	if _, ok := builtInRoleGUIDs[nameOrGUID]; ok {
		return true
	}
	return guidRegexp.MatchString(nameOrGUID)
}

// KnownNames returns the mapped role names, sorted, for error messages.
func KnownNames() []string {
	names := make([]string, 0, len(builtInRoleGUIDs))
	for name := range builtInRoleGUIDs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
