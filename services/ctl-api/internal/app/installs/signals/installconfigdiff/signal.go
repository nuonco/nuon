package installconfigdiff

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "install-config-diff"

type Signal struct {
	InstallID             string `json:"install_id" validate:"required"`
	NewAppConfigID        string `json:"new_app_config_id" validate:"required"`
	InstallConfigUpdateID string `json:"install_config_update_id,omitempty"`

	// FlowID and StepID are injected by the flow engine via SignalWithStepContext.
	FlowID string `json:"flow_id,omitempty"`
	StepID string `json:"step_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	if s.NewAppConfigID == "" {
		return fmt.Errorf("new_app_config_id is required")
	}
	return nil
}

// ComponentDiffEntry is an alias for the shared type.
type ComponentDiffEntry = app.ComponentDiffEntry

// ConfigDiff is an alias for the shared type.
type ConfigDiff = app.InstallConfigDiff

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Get the install to find its current app config
	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	// Get new app config with component config connections
	newAppCfg, err := activities.AwaitGetAppConfigByID(ctx, s.NewAppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get new app config: %w", err)
	}

	diff := ConfigDiff{
		Added:     []ComponentDiffEntry{},
		Removed:   []ComponentDiffEntry{},
		Changed:   []ComponentDiffEntry{},
		Unchanged: []ComponentDiffEntry{},
	}

	// Build lookup for new config connections by component ID
	newConnByComponent := make(map[string]*app.ComponentConfigConnection, len(newAppCfg.ComponentConfigConnections))
	for i := range newAppCfg.ComponentConfigConnections {
		ccc := &newAppCfg.ComponentConfigConnections[i]
		newConnByComponent[ccc.ComponentID] = ccc
	}

	// If the install has an existing app config, compare against it
	if install.AppConfigID != "" {
		oldAppCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
		if err != nil {
			l.Warn("unable to get old app config, treating all components as added", "error", err)
		} else {
			// Build lookup for old config connections by component ID
			oldConnByComponent := make(map[string]*app.ComponentConfigConnection, len(oldAppCfg.ComponentConfigConnections))
			for i := range oldAppCfg.ComponentConfigConnections {
				ccc := &oldAppCfg.ComponentConfigConnections[i]
				oldConnByComponent[ccc.ComponentID] = ccc
			}

			// Compare: find changed, unchanged, and removed
			for componentID, oldConn := range oldConnByComponent {
				newConn, exists := newConnByComponent[componentID]
				if !exists {
					diff.Removed = append(diff.Removed, ComponentDiffEntry{
						ComponentID:   componentID,
						ComponentName: oldConn.ComponentName,
						OldChecksum:   oldConn.Checksum,
					})
					continue
				}

				if oldConn.Checksum != "" && newConn.Checksum != "" && oldConn.Checksum == newConn.Checksum {
					diff.Unchanged = append(diff.Unchanged, ComponentDiffEntry{
						ComponentID:   componentID,
						ComponentName: newConn.ComponentName,
						OldChecksum:   oldConn.Checksum,
						NewChecksum:   newConn.Checksum,
					})
				} else {
					diff.Changed = append(diff.Changed, ComponentDiffEntry{
						ComponentID:   componentID,
						ComponentName: newConn.ComponentName,
						OldChecksum:   oldConn.Checksum,
						NewChecksum:   newConn.Checksum,
					})
				}

				// Remove from new map so we can find truly new components
				delete(newConnByComponent, componentID)
			}

			// Remaining in newConnByComponent are new additions
			for componentID, newConn := range newConnByComponent {
				diff.Added = append(diff.Added, ComponentDiffEntry{
					ComponentID:   componentID,
					ComponentName: newConn.ComponentName,
					NewChecksum:   newConn.Checksum,
				})
			}
		}
	}

	// If no old config, everything is new
	if install.AppConfigID == "" {
		for componentID, newConn := range newConnByComponent {
			diff.Added = append(diff.Added, ComponentDiffEntry{
				ComponentID:   componentID,
				ComponentName: newConn.ComponentName,
				NewChecksum:   newConn.Checksum,
			})
		}
	}

	l.Info("config diff computed",
		"install_id", s.InstallID,
		"added", len(diff.Added),
		"removed", len(diff.Removed),
		"changed", len(diff.Changed),
		"unchanged", len(diff.Unchanged),
	)

	// Serialize and persist the diff.
	diffJSON, err := json.Marshal(diff)
	if err != nil {
		return fmt.Errorf("unable to marshal diff: %w", err)
	}

	// Save the diff blob on the InstallConfigUpdate record if we have the ID.
	if s.InstallConfigUpdateID != "" {
		if err := activities.AwaitSaveInstallConfigUpdateDiff(ctx, &activities.SaveInstallConfigUpdateDiffInput{
			InstallConfigUpdateID: s.InstallConfigUpdateID,
			DiffJSON:              string(diffJSON),
		}); err != nil {
			l.Warn("unable to save config diff blob", "error", err)
		}
	}

	return nil
}
