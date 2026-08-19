package sync

import (
	"context"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// Syncer defines the interface for syncing app configurations to a backing store.
//
// The only implementation is the database-backed syncer in
// services/ctl-api/internal/pkg/config/syncer. Clients do not sync directly:
// they push a config to the API in its intermediate form and ask the API to
// apply it. This interface lives here because the state and error types it
// works with are shared with those clients.
type Syncer interface {
	// Sync performs the full synchronization operation, creating or updating
	// app configs, components, and their configurations.
	//
	// The context must contain org and account information set via cctx.SetOrgContext()
	// and cctx.SetAccountContext() before calling this method.
	//
	// Returns an error if the sync operation fails at any step.
	Sync(ctx context.Context) error

	// GetAppConfigID returns the ID of the app config that was created or updated
	// during the most recent sync operation.
	//
	// This should only be called after a successful Sync() operation.
	GetAppConfigID() string

	// GetComponentStateIds returns the IDs of all components that were synced
	// during the most recent sync operation.
	//
	// This should only be called after a successful Sync() operation.
	GetComponentStateIds() []string

	// GetActionStateIds returns the IDs of all actions that were synced
	// during the most recent sync operation.
	//
	// This should only be called after a successful Sync() operation.
	GetActionStateIds() []string

	// GetComponentsScheduled returns the components that need a build as a
	// result of the most recent sync — those whose config changed. It is empty
	// unless the sync was configured to own build scheduling.
	//
	// The set is persisted in State.Result so the CLI, which only sees the app
	// config it polls, can wait on those builds.
	// This should only be called after a successful Sync() operation.
	GetComponentsScheduled() []ComponentState

	GetComponentsCreated() []string

	GetAppBranchesCreated() []string

	// OrphanedComponents returns a map of component names to IDs for components
	// that existed in the previous config but are no longer in the current config.
	//
	// This allows consumers to notify users about removed components.
	// This should only be called after a successful Sync() operation.
	OrphanedComponents() map[string]string

	// OrphanedActions returns a map of action names to IDs for actions that
	// existed in the previous config but are no longer in the current config.
	//
	// This allows consumers to notify users about removed actions.
	// This should only be called after a successful Sync() operation.
	OrphanedActions() map[string]string

	// GetRunbookStateIds returns the IDs of all runbooks that were synced
	// during the most recent sync operation.
	//
	// This should only be called after a successful Sync() operation.
	GetRunbookStateIds() []string

	// OrphanedRunbooks returns a map of runbook names to IDs for runbooks that
	// existed in the previous config but are no longer in the current config.
	//
	// This allows consumers to notify users about removed runbooks.
	// This should only be called after a successful Sync() operation.
	OrphanedRunbooks() map[string]string

	// SyncInstall syncs a single install config to the database.
	// If the install does not exist, it is created. If it exists and has
	// changed, it is updated (inputs, labels, config, component toggles).
	SyncInstall(ctx context.Context, install *config.Install) (*InstallSyncResult, error)
}

type InstallSyncResult struct {
	InstallID   string     `json:"install_id"`
	InstallName string     `json:"install_name"`
	Created     bool       `json:"created"`
	Changed     bool       `json:"changed"`
	Diff        *diff.Diff `json:"diff,omitempty"`
}

// ComponentState represents the synchronized state of a component.
// This is stored in the app config's state field as JSON to track
// what was synced in each config version.
type ComponentState struct {
	Name     string                  `json:"name"`
	ID       string                  `json:"id"`
	ConfigID string                  `json:"config_id"`
	Type     models.AppComponentType `json:"type"`
	Checksum string                  `json:"checksum"`
}
