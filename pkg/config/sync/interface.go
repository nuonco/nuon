package sync

import (
	"context"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type Syncer interface {
	Sync(ctx context.Context) error

	GetAppConfigID() string

	GetComponentStateIds() []string

	GetActionStateIds() []string

	GetComponentsScheduled() []ComponentState

	GetComponentsCreated() []string

	GetAppBranchesCreated() []string

	OrphanedComponents() map[string]string

	OrphanedActions() map[string]string

	GetRunbookStateIds() []string

	OrphanedRunbooks() map[string]string

	GetAppBranchConfigsUpdated() []AppBranchConfigState

	SyncInstall(ctx context.Context, install *config.Install) (*InstallSyncResult, error)
}

type InstallSyncResult struct {
	InstallID        string     `json:"install_id"`
	InstallName      string     `json:"install_name"`
	Created          bool       `json:"created"`
	Changed          bool       `json:"changed"`
	Diff             *diff.Diff `json:"diff,omitempty"`
	AppBranchChanged bool       `json:"app_branch_changed,omitempty"`
	AppBranchID      string     `json:"app_branch_id,omitempty"`
}

type ComponentState struct {
	Name     string                  `json:"name"`
	ID       string                  `json:"id"`
	ConfigID string                  `json:"config_id"`
	Type     models.AppComponentType `json:"type"`
	Checksum string                  `json:"checksum"`
}
