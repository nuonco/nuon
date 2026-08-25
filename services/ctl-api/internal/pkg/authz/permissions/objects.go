package permissions

type ResourceKind string

const (
	KindApp     ResourceKind = "app"
	KindInstall ResourceKind = "install"
	KindStack   ResourceKind = "stack"
)

// Object is the canonical grant key for a resource-scoped permission:
// "<orgID>:<kind>/<id>". Exactly two levels — CanPerform's parent fallback
// resolves the org tier, so org-wide roles cover every scoped object.
func Object(orgID string, kind ResourceKind, id string) string {
	return orgID + ":" + string(kind) + "/" + id
}

// StackObject keeps the key shape already minted into stack role rows.
func StackObject(orgID, installID string) string {
	return Object(orgID, KindStack, installID)
}
