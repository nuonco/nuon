package sync

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// Syncer defines the interface for syncing app configurations to a backing store.
// Implementations can sync via API calls (apisyncer) or direct database access (dbsyncer).
//
// The Syncer interface provides a pluggable architecture that allows different sync
// strategies while maintaining a consistent API for consumers.
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

	// GetComponentsScheduled returns the components that had builds scheduled
	// during the most recent sync operation.
	//
	// The CLI uses this to poll build status after syncing.
	// This should only be called after a successful Sync() operation.
	GetComponentsScheduled() []ComponentState

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
