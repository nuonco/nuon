// Package configdiff computes the difference between two app config versions
// as it applies to a single install.
package configdiff

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func preload(db *gorm.DB) *gorm.DB {
	return db.
		Preload("ComponentConfigConnections").
		Preload("ComponentConfigConnections.Component").
		Preload("SandboxConfig").
		Preload("SandboxConfig.PublicGitVCSConfig").
		Preload("SandboxConfig.ConnectedGithubVCSConfig").
		Preload("StackConfig")
}

// ComputeInstallConfigDiff diffs oldAppConfigID against newAppConfigID. An empty
// oldAppConfigID means the install has never been pinned to a config, so
// everything counts as added.
func ComputeInstallConfigDiff(ctx context.Context, db *gorm.DB, oldAppConfigID, newAppConfigID string) (*app.InstallConfigDiff, error) {
	var newAppCfg app.AppConfig
	if err := preload(db.WithContext(ctx)).First(&newAppCfg, "id = ?", newAppConfigID).Error; err != nil {
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

	if oldAppConfigID != "" {
		var oldAppCfg app.AppConfig
		if err := preload(db.WithContext(ctx)).First(&oldAppCfg, "id = ?", oldAppConfigID).Error; err == nil {
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

				entry := componentDiffEntry(oldConn, newConn)
				if checksumsEqual(oldConn, newConn) {
					if entry.BuildChanged {
						diff.Changed = append(diff.Changed, entry)
					} else {
						diff.Unchanged = append(diff.Unchanged, entry)
					}
				} else {
					diff.Changed = append(diff.Changed, entry)
				}

				delete(newConnByComponent, componentID)
			}

			for _, newConn := range newConnByComponent {
				entry := componentDiffEntry(nil, newConn)
				diff.Added = append(diff.Added, entry)
			}

			// Every sync writes fresh sandbox and stack config rows, so their IDs
			// always differ between versions — only a content change is a real change.
			if oldAppCfg.SandboxConfig.ID != newAppCfg.SandboxConfig.ID &&
				!sandboxConfigEqual(oldAppCfg.SandboxConfig, newAppCfg.SandboxConfig) {
				diff.SandboxChanged = true
				diff.SandboxOldID = oldAppCfg.SandboxConfig.ID
				diff.SandboxNewID = newAppCfg.SandboxConfig.ID
			}
			if oldAppCfg.StackConfig.ID != newAppCfg.StackConfig.ID &&
				!StackConfigEqual(oldAppCfg.StackConfig, newAppCfg.StackConfig) {
				diff.StackChanged = true
				diff.StackOldID = oldAppCfg.StackConfig.ID
				diff.StackNewID = newAppCfg.StackConfig.ID
			}

			oldSandboxBuildID, err := latestActiveSandboxBuildID(ctx, db, oldAppConfigID)
			if err != nil {
				return nil, err
			}
			newSandboxBuildID, err := latestActiveSandboxBuildID(ctx, db, newAppConfigID)
			if err != nil {
				return nil, err
			}
			if oldSandboxBuildID != newSandboxBuildID {
				diff.SandboxBuildChanged = true
				diff.SandboxBuildOldID = oldSandboxBuildID
				diff.SandboxBuildNewID = newSandboxBuildID
			}

			return diff, nil
		}
	}

	for _, newConn := range newConnByComponent {
		entry := componentDiffEntry(nil, newConn)
		diff.Added = append(diff.Added, entry)
	}
	if newAppCfg.SandboxConfig.ID != "" {
		diff.SandboxChanged = true
		diff.SandboxNewID = newAppCfg.SandboxConfig.ID
	}
	if newAppCfg.StackConfig.ID != "" {
		diff.StackChanged = true
		diff.StackNewID = newAppCfg.StackConfig.ID
	}
	if newSandboxBuildID, err := latestActiveSandboxBuildID(ctx, db, newAppConfigID); err != nil {
		return nil, err
	} else if newSandboxBuildID != "" {
		diff.SandboxBuildChanged = true
		diff.SandboxBuildNewID = newSandboxBuildID
	}

	return diff, nil
}

func checksumsEqual(oldConn, newConn *app.ComponentConfigConnection) bool {
	return oldConn.Checksum != "" && newConn.Checksum != "" && oldConn.Checksum == newConn.Checksum
}

func cccBuildID(ccc *app.ComponentConfigConnection) string {
	if ccc == nil || !ccc.LatestBuildID.Valid {
		return ""
	}
	return ccc.LatestBuildID.String
}

func componentDiffEntry(oldConn, newConn *app.ComponentConfigConnection) app.ComponentDiffEntry {
	entry := app.ComponentDiffEntry{
		ComponentID:   newConn.ComponentID,
		ComponentName: newConn.ComponentName,
		ComponentType: string(newConn.Type),
		NewChecksum:   newConn.Checksum,
		NewBuildID:    cccBuildID(newConn),
	}
	if oldConn != nil {
		entry.OldChecksum = oldConn.Checksum
		entry.OldBuildID = cccBuildID(oldConn)
		entry.BuildChanged = checksumsEqual(oldConn, newConn) && entry.OldBuildID != entry.NewBuildID &&
			(entry.OldBuildID != "" || entry.NewBuildID != "")
	}
	return entry
}

func latestActiveSandboxBuildID(ctx context.Context, db *gorm.DB, appConfigID string) (string, error) {
	if appConfigID == "" {
		return "", nil
	}
	var build app.AppSandboxBuild
	err := db.WithContext(ctx).
		Where(app.AppSandboxBuild{
			AppConfigID: appConfigID,
			Status:      app.AppSandboxBuildStatusActive,
		}).
		Order("created_at DESC").
		First(&build).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("unable to get active sandbox build for app config %s: %w", appConfigID, err)
	}
	return build.ID, nil
}

// sandboxContent is everything about a sandbox config that decides what gets
// deployed. Orchestration knobs (max_auto_retries, skip_noops,
// auto_approve_on_policies_passing) are deliberately excluded: changing a retry
// count should not force a reprovision. Anything that selects or renders the
// sandbox belongs here — omitting a field means a real change is read as a
// no-op and never reaches installs.
type sandboxContent struct {
	Source         sandboxSource `json:"source"`
	Variables      any           `json:"variables"`
	EnvVars        any           `json:"env_vars"`
	VariablesFiles any           `json:"variables_files"`
	References     any           `json:"references"`
	Type           string        `json:"type"`
	TerraformVer   string        `json:"terraform_version"`
	DriftSchedule  string        `json:"drift_schedule"`
	Runtime        string        `json:"runtime"`
	PulumiVersion  string        `json:"pulumi_version"`
	PulumiConfig   any           `json:"pulumi_config"`
	OperationRoles any           `json:"operation_roles"`
	AWSRegionType  string        `json:"aws_region_type"`
}

// sandboxSource identifies which code the sandbox runs. A ref, directory or
// repo bump changes the deployed infrastructure while leaving every other
// field identical.
type sandboxSource struct {
	Kind       string `json:"kind"`
	Repo       string `json:"repo"`
	Directory  string `json:"directory"`
	Branch     string `json:"branch"`
	PathFilter string `json:"path_filter"`
	Connection string `json:"connection,omitempty"`
}

func sandboxSourceOf(c app.AppSandboxConfig) sandboxSource {
	switch {
	case c.ConnectedGithubVCSConfig != nil:
		v := c.ConnectedGithubVCSConfig
		return sandboxSource{
			Kind:       "connected-github",
			Repo:       v.Repo,
			Directory:  v.Directory,
			Branch:     v.Branch,
			PathFilter: v.PathFilter,
			Connection: v.VCSConnectionID,
		}
	case c.PublicGitVCSConfig != nil:
		v := c.PublicGitVCSConfig
		return sandboxSource{
			Kind:       "public-git",
			Repo:       v.Repo,
			Directory:  v.Directory,
			Branch:     v.Branch,
			PathFilter: v.PathFilter,
		}
	default:
		return sandboxSource{Kind: "builtin"}
	}
}

func sandboxContentOf(c app.AppSandboxConfig) sandboxContent {
	return sandboxContent{
		Source:         sandboxSourceOf(c),
		Variables:      c.Variables,
		EnvVars:        c.EnvVars,
		VariablesFiles: c.VariablesFiles,
		References:     c.References,
		Type:           c.Type,
		TerraformVer:   c.TerraformVersion,
		DriftSchedule:  c.DriftSchedule,
		Runtime:        c.Runtime,
		PulumiVersion:  c.PulumiVersion,
		PulumiConfig:   c.PulumiConfig,
		OperationRoles: c.OperationRoles,
		AWSRegionType:  c.AWSRegionType.String,
	}
}

func sandboxConfigEqual(a, b app.AppSandboxConfig) bool {
	return contentHashEqual(sandboxContentOf(a), sandboxContentOf(b))
}

func StackConfigEqual(a, b app.AppStackConfig) bool {
	type content struct {
		Type                    string                   `json:"type"`
		Name                    string                   `json:"name"`
		Description             string                   `json:"description"`
		RunnerNestedTemplateURL string                   `json:"runner_nested_template_url"`
		VPCNestedTemplateURL    string                   `json:"vpc_nested_template_url"`
		DeploymentScope         app.StackDeploymentScope `json:"deployment_scope"`
		CustomNestedStacks      any                      `json:"custom_nested_stacks"`
	}
	ac := content{
		Type:                    string(a.Type),
		Name:                    a.Name,
		Description:             a.Description,
		RunnerNestedTemplateURL: a.RunnerNestedTemplateURL,
		VPCNestedTemplateURL:    a.VPCNestedTemplateURL,
		DeploymentScope:         a.DeploymentScope,
		CustomNestedStacks:      a.CustomNestedStacks,
	}
	bc := content{
		Type:                    string(b.Type),
		Name:                    b.Name,
		Description:             b.Description,
		RunnerNestedTemplateURL: b.RunnerNestedTemplateURL,
		VPCNestedTemplateURL:    b.VPCNestedTemplateURL,
		DeploymentScope:         b.DeploymentScope,
		CustomNestedStacks:      b.CustomNestedStacks,
	}
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
