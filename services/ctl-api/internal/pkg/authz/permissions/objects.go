package permissions

type ResourceKind string

const (
	KindApp     ResourceKind = "app"
	KindInstall ResourceKind = "install"
	KindStack   ResourceKind = "stack"
)

func Object(orgID string, kind ResourceKind, id string) string {
	return orgID + ":" + string(kind) + "/" + id
}

func StackObject(orgID, installID string) string {
	return Object(orgID, KindStack, installID)
}
