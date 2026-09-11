package app

import (
	"time"

	"github.com/nuonco/nuon/pkg/config"
)

type AppConfigResourceKind string

const (
	AppConfigResourceKindComponent AppConfigResourceKind = "component"
	AppConfigResourceKindAction    AppConfigResourceKind = "action"
	AppConfigResourceKindRunbook   AppConfigResourceKind = "runbook"
)

type AppConfigResource struct {
	ID   string                `json:"id"`
	Name string                `json:"name"`
	Kind AppConfigResourceKind `json:"kind"`
}

// Config is keyed by Go field name, not snake_case: the CLI writes the blob with stdlib
// json.Marshal and pkg/config carries no json tags. Adding tags there would make every
// stored blob unreadable.
type AppConfigIntermediate struct {
	ConfigID    string          `json:"config_id"`
	AppID       string          `json:"app_id"`
	AppBranchID string          `json:"app_branch_id,omitzero"`
	Status      AppConfigStatus `json:"status,omitzero"`
	StatusV2    CompositeStatus `json:"status_v2,omitzero"`
	Version     int             `json:"version,omitzero"`
	CLIVersion  string          `json:"cli_version,omitzero"`
	Checksum    string          `json:"checksum,omitzero"`
	Size        int64           `json:"size,omitzero"`
	CreatedAt   time.Time       `json:"created_at,omitzero"`
	CreatedByID string          `json:"created_by_id,omitzero"`

	VCSConnectionCommit *VCSConnectionCommit `json:"vcs_connection_commit,omitzero"`

	Config    *config.AppConfig   `json:"config"`
	Resources []AppConfigResource `json:"resources"`
}
