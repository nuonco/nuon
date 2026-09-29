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
	InstallID                 string `json:"install_id" validate:"required"`
	NewAppConfigID            string `json:"new_app_config_id" validate:"required"`
	InstallAppConfigVersionID string `json:"install_config_update_id,omitempty"`

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

type ComponentDiffEntry = app.ComponentDiffEntry

type ConfigDiff = app.InstallConfigDiff

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

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

	newConnByComponent := make(map[string]*app.ComponentConfigConnection, len(newAppCfg.ComponentConfigConnections))
	for i := range newAppCfg.ComponentConfigConnections {
		ccc := &newAppCfg.ComponentConfigConnections[i]
		newConnByComponent[ccc.ComponentID] = ccc
	}

	if install.AppConfigID != "" {
		oldAppCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
		if err != nil {
			l.Warn("unable to get old app config, treating all components as added", "error", err)
		} else {
			oldConnByComponent := make(map[string]*app.ComponentConfigConnection, len(oldAppCfg.ComponentConfigConnections))
			for i := range oldAppCfg.ComponentConfigConnections {
				ccc := &oldAppCfg.ComponentConfigConnections[i]
				oldConnByComponent[ccc.ComponentID] = ccc
			}

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

				delete(newConnByComponent, componentID)
			}

			for componentID, newConn := range newConnByComponent {
				diff.Added = append(diff.Added, ComponentDiffEntry{
					ComponentID:   componentID,
					ComponentName: newConn.ComponentName,
					NewChecksum:   newConn.Checksum,
				})
			}
		}
	}

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

	diffJSON, err := json.Marshal(diff)
	if err != nil {
		return fmt.Errorf("unable to marshal diff: %w", err)
	}

	if s.InstallAppConfigVersionID != "" {
		if err := activities.AwaitSaveInstallAppConfigVersionDiff(ctx, &activities.SaveInstallAppConfigVersionDiffInput{
			InstallAppConfigVersionID: s.InstallAppConfigVersionID,
			DiffJSON:                  string(diffJSON),
		}); err != nil {
			l.Warn("unable to save config diff blob", "error", err)
		}
	}

	return nil
}
