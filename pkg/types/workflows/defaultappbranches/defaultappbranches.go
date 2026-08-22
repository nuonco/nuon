// Package defaultappbranches carries the input and progress types for the
// fleet-wide default app branch backfill, which gives every app the `default`
// branch and all-installs group that `nuon apps sync` otherwise creates lazily
// on its first run under the default-app-branches flag.
package defaultappbranches

const (
	WorkflowName = "BackfillDefaultAppBranches"
	WorkflowID   = "general-default-app-branch-backfill"

	ProgressQueryType = "progress"
)

// Request is the orchestrator input. Callers populate OrgIDs and DryRun; the
// workflow fills in the rest and carries it across continue-as-new.
type Request struct {
	OrgIDs []string `json:"org_ids"`
	DryRun bool     `json:"dry_run"`

	Initialized       bool     `json:"initialized"`
	Pending           []string `json:"pending"`
	AppsTotal         int      `json:"apps_total"`
	Created           int      `json:"created"`
	Existing          int      `json:"existing"`
	Claimed           int      `json:"claimed"`
	Failed            int      `json:"failed"`
	FailedAppIDs      []string `json:"failed_app_ids"`
	InstallsConnected int      `json:"installs_connected"`
}

// Progress is the live snapshot returned by the progress query.
type Progress struct {
	DryRun            bool     `json:"dry_run"`
	AppsTotal         int      `json:"apps_total"`
	AppsDone          int      `json:"apps_done"`
	Created           int      `json:"created"`
	Existing          int      `json:"existing"`
	Claimed           int      `json:"claimed"`
	Failed            int      `json:"failed"`
	FailedAppIDs      []string `json:"failed_app_ids,omitempty"`
	InstallsConnected int      `json:"installs_connected"`
	CurrentAppID      string   `json:"current_app_id,omitempty"`
	Done              bool     `json:"done"`
}
