package activities

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ComputeInstallConfigDiffInput struct {
	OldAppConfigID string `json:"old_app_config_id"`
	NewAppConfigID string `json:"new_app_config_id" validate:"required"`
}

type ComputeInstallConfigDiffOutput struct {
	Diff *app.InstallConfigDiff `json:"diff"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ComputeInstallConfigDiff(ctx context.Context, input *ComputeInstallConfigDiffInput) (*ComputeInstallConfigDiffOutput, error) {
	var newAppCfg app.AppConfig
	if err := a.db.WithContext(ctx).
		Preload("ComponentConfigConnections").
		Preload("ComponentConfigConnections.Component").
		Preload("SandboxConfig").
		Preload("StackConfig").
		First(&newAppCfg, "id = ?", input.NewAppConfigID).Error; err != nil {
		return nil, fmt.Errorf("unable to get new app config: %w", err)
	}

	diff := &app.InstallConfigDiff{
		Added:     []app.ComponentDiffEntry{},
		Removed:   []app.ComponentDiffEntry{},
		Changed:   []app.ComponentDiffEntry{},
		Unchanged: []app.ComponentDiffEntry{},
	}

	newConnByComponent := make(map[string]*app.ComponentConfigConnection, len(newAppCfg.ComponentConfigConnections))
	for i := range newAppCfg.ComponentConfigConnections {
		ccc := &newAppCfg.ComponentConfigConnections[i]
		newConnByComponent[ccc.ComponentID] = ccc
	}

	if input.OldAppConfigID != "" {
		var oldAppCfg app.AppConfig
		if err := a.db.WithContext(ctx).
			Preload("ComponentConfigConnections").
			Preload("ComponentConfigConnections.Component").
			Preload("SandboxConfig").
			Preload("StackConfig").
			First(&oldAppCfg, "id = ?", input.OldAppConfigID).Error; err == nil {

			oldConnByComponent := make(map[string]*app.ComponentConfigConnection, len(oldAppCfg.ComponentConfigConnections))
			for i := range oldAppCfg.ComponentConfigConnections {
				ccc := &oldAppCfg.ComponentConfigConnections[i]
				oldConnByComponent[ccc.ComponentID] = ccc
			}

			for componentID, oldConn := range oldConnByComponent {
				newConn, exists := newConnByComponent[componentID]
				if !exists {
					diff.Removed = append(diff.Removed, app.ComponentDiffEntry{
						ComponentID:   componentID,
						ComponentName: oldConn.ComponentName,
						ComponentType: string(oldConn.Type),
						OldChecksum:   oldConn.Checksum,
					})
					continue
				}

				if oldConn.Checksum != "" && newConn.Checksum != "" && oldConn.Checksum == newConn.Checksum {
					diff.Unchanged = append(diff.Unchanged, app.ComponentDiffEntry{
						ComponentID:   componentID,
						ComponentName: newConn.ComponentName,
						ComponentType: string(newConn.Type),
						OldChecksum:   oldConn.Checksum,
						NewChecksum:   newConn.Checksum,
					})
				} else {
					diff.Changed = append(diff.Changed, app.ComponentDiffEntry{
						ComponentID:   componentID,
						ComponentName: newConn.ComponentName,
						ComponentType: string(newConn.Type),
						OldChecksum:   oldConn.Checksum,
						NewChecksum:   newConn.Checksum,
					})
				}

				delete(newConnByComponent, componentID)
			}

			for componentID, newConn := range newConnByComponent {
				diff.Added = append(diff.Added, app.ComponentDiffEntry{
					ComponentID:   componentID,
					ComponentName: newConn.ComponentName,
					ComponentType: string(newConn.Type),
					NewChecksum:   newConn.Checksum,
				})
			}

			if oldAppCfg.SandboxConfig.ID != newAppCfg.SandboxConfig.ID {
				if !sandboxConfigEqual(oldAppCfg.SandboxConfig, newAppCfg.SandboxConfig) {
					diff.SandboxChanged = true
					diff.SandboxOldID = oldAppCfg.SandboxConfig.ID
					diff.SandboxNewID = newAppCfg.SandboxConfig.ID
				}
			}
			if oldAppCfg.StackConfig.ID != newAppCfg.StackConfig.ID {
				if !stackConfigEqual(oldAppCfg.StackConfig, newAppCfg.StackConfig) {
					diff.StackChanged = true
					diff.StackOldID = oldAppCfg.StackConfig.ID
					diff.StackNewID = newAppCfg.StackConfig.ID
				}
			}
		}
	}

	if input.OldAppConfigID == "" {
		for componentID, newConn := range newConnByComponent {
			diff.Added = append(diff.Added, app.ComponentDiffEntry{
				ComponentID:   componentID,
				ComponentName: newConn.ComponentName,
				ComponentType: string(newConn.Type),
				NewChecksum:   newConn.Checksum,
			})
		}
		if newAppCfg.SandboxConfig.ID != "" {
			diff.SandboxChanged = true
			diff.SandboxNewID = newAppCfg.SandboxConfig.ID
		}
		if newAppCfg.StackConfig.ID != "" {
			diff.StackChanged = true
			diff.StackNewID = newAppCfg.StackConfig.ID
		}
	}

	return &ComputeInstallConfigDiffOutput{Diff: diff}, nil
}

func sandboxConfigEqual(a, b app.AppSandboxConfig) bool {
	type content struct {
		Variables      any    `json:"variables"`
		EnvVars        any    `json:"env_vars"`
		VariablesFiles any    `json:"variables_files"`
		Type           string `json:"type"`
		TerraformVer   string `json:"terraform_version"`
		DriftSchedule  string `json:"drift_schedule"`
		Runtime        string `json:"runtime"`
		PulumiVersion  string `json:"pulumi_version"`
		PulumiConfig   any    `json:"pulumi_config"`
	}
	ac := content{a.Variables, a.EnvVars, a.VariablesFiles, a.Type, a.TerraformVersion, a.DriftSchedule, a.Runtime, a.PulumiVersion, a.PulumiConfig}
	bc := content{b.Variables, b.EnvVars, b.VariablesFiles, b.Type, b.TerraformVersion, b.DriftSchedule, b.Runtime, b.PulumiVersion, b.PulumiConfig}
	return contentHashEqual(ac, bc)
}

func stackConfigEqual(a, b app.AppStackConfig) bool {
	type content struct {
		Type                    string `json:"type"`
		Name                    string `json:"name"`
		Description             string `json:"description"`
		RunnerNestedTemplateURL string `json:"runner_nested_template_url"`
		VPCNestedTemplateURL    string `json:"vpc_nested_template_url"`
		CustomNestedStacks      any    `json:"custom_nested_stacks"`
	}
	ac := content{string(a.Type), a.Name, a.Description, a.RunnerNestedTemplateURL, a.VPCNestedTemplateURL, a.CustomNestedStacks}
	bc := content{string(b.Type), b.Name, b.Description, b.RunnerNestedTemplateURL, b.VPCNestedTemplateURL, b.CustomNestedStacks}
	return contentHashEqual(ac, bc)
}

func contentHashEqual(a, b any) bool {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return sha256.Sum256(aJSON) == sha256.Sum256(bJSON)
}
