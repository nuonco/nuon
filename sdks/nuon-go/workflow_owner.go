package nuon

// WorkflowOwner names the ancestor a workflow hangs off. When set, workflow
// calls take the nested route, which authorizes against that ancestor rather
// than the org — the only route a resource-scoped identity can use. The zero
// value keeps the bare, org-tier route.
//
// A branch-owned workflow needs both ids: the workflow record carries its
// owner id but not the app above it, so the caller supplies the pair.
type WorkflowOwner struct {
	InstallID   string
	AppID       string
	AppBranchID string
}

func (o WorkflowOwner) ownedByInstall() bool {
	return o.InstallID != ""
}

func (o WorkflowOwner) ownedByAppBranch() bool {
	return o.AppID != "" && o.AppBranchID != ""
}
