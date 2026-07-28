package installs

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

const ManagedByGitInstallConfig = "nuon/git/install-config"

func SyncInstall(ctx context.Context, db *gorm.DB, installHelpers *installhelpers.Helpers, appID string, install *config.Install) (*sync.InstallSyncResult, error) {
	if install == nil {
		return nil, fmt.Errorf("install config is nil")
	}
	if install.Name == "" {
		return nil, fmt.Errorf("install config name is required")
	}

	var existing app.Install
	err := db.WithContext(ctx).
		Preload("InstallConfig").
		Preload("AWSAccount").
		Preload("GCPAccount").
		Preload("AzureAccount").
		Where(app.Install{AppID: appID, Name: install.Name}).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return createInstall(ctx, db, installHelpers, appID, install)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to look up install %s: %w", install.Name, err)
	}

	return updateInstall(ctx, db, installHelpers, &existing, install)
}

func createInstall(ctx context.Context, db *gorm.DB, installHelpers *installhelpers.Helpers, appID string, installCfg *config.Install) (*sync.InstallSyncResult, error) {
	inputs := make(map[string]*string)
	for k, v := range installCfg.FlattenedInputs() {
		val := v
		inputs[k] = &val
	}

	req := &installhelpers.CreateInstallParams{
		Name:   installCfg.Name,
		Inputs: inputs,
		Labels: installCfg.Labels,
		Metadata: installhelpers.InstallMetadata{
			ManagedBy: ManagedByGitInstallConfig,
		},
	}

	if installCfg.AWSAccount != nil {
		req.AWSAccount = &struct {
			Region       string `json:"region"`
			ConnectionID string `json:"connection_id,omitempty"`
		}{Region: installCfg.AWSAccount.Region}
	}
	if installCfg.GCPAccount != nil {
		req.GCPAccount = &struct {
			ProjectID string `json:"project_id"`
			Region    string `json:"region"`
		}{ProjectID: installCfg.GCPAccount.ProjectID, Region: installCfg.GCPAccount.Region}
	}
	if installCfg.AzureAccount != nil {
		req.AzureAccount = &struct {
			Location string `json:"location"`
		}{Location: installCfg.AzureAccount.Location}
	}

	if installCfg.ApprovalOption != config.InstallApprovalOptionUnknown ||
		installCfg.StackOverrides.HasOverrides() ||
		len(installCfg.ComponentToggles) > 0 {
		icParams := &installhelpers.CreateInstallConfigParams{}
		if installCfg.ApprovalOption != config.InstallApprovalOptionUnknown {
			icParams.ApprovalOption = app.InstallApprovalOption(installCfg.ApprovalOption)
		}
		if installCfg.StackOverrides != nil {
			if installCfg.StackOverrides.VPCNestedTemplateURL != "" {
				url := installCfg.StackOverrides.VPCNestedTemplateURL
				icParams.VPCNestedTemplateURL = &url
			}
			if installCfg.StackOverrides.RunnerNestedTemplateURL != "" {
				url := installCfg.StackOverrides.RunnerNestedTemplateURL
				icParams.RunnerNestedTemplateURL = &url
			}
			if len(installCfg.StackOverrides.CustomNestedStacks) > 0 {
				icParams.CustomNestedStacks = installCfg.StackOverrides.CustomNestedStacks
			}
		}
		req.InstallConfig = icParams
	}

	created, err := installHelpers.CreateInstall(ctx, appID, req)
	if err != nil {
		return nil, fmt.Errorf("unable to create install %s: %w", installCfg.Name, err)
	}

	d, _ := installCfg.Diff(nil)

	return &sync.InstallSyncResult{
		InstallID:   created.ID,
		InstallName: created.Name,
		Created:     true,
		Changed:     true,
		Diff:        d,
	}, nil
}

func updateInstall(ctx context.Context, db *gorm.DB, installHelpers *installhelpers.Helpers, existing *app.Install, installCfg *config.Install) (*sync.InstallSyncResult, error) {
	upstream := existingToConfig(existing)
	d, err := installCfg.Diff(upstream)
	if err != nil {
		return nil, fmt.Errorf("unable to compute diff for install %s: %w", installCfg.Name, err)
	}

	summary := d.Summary()
	if !summary.HasChanged {
		return &sync.InstallSyncResult{
			InstallID:   existing.ID,
			InstallName: existing.Name,
			Created:     false,
			Changed:     false,
		}, nil
	}

	definedInputs := installCfg.FlattenedInputs()
	if len(definedInputs) > 0 {
		inputs := make(map[string]*string)
		for k, v := range definedInputs {
			val := v
			inputs[k] = &val
		}
		installInputs := app.InstallInputs{
			InstallID: existing.ID,
			Values:    inputs,
		}
		if err := db.WithContext(ctx).Create(&installInputs).Error; err != nil {
			return nil, fmt.Errorf("unable to update inputs for install %s: %w", installCfg.Name, err)
		}
	}

	hasConfigFields := installCfg.ApprovalOption != config.InstallApprovalOptionUnknown ||
		installCfg.StackOverrides.HasOverrides() ||
		len(installCfg.ComponentToggles) > 0

	if hasConfigFields {
		updates := map[string]any{}
		if installCfg.ApprovalOption != config.InstallApprovalOptionUnknown {
			updates["approval_option"] = string(installCfg.ApprovalOption)
		}
		if installCfg.StackOverrides != nil {
			if installCfg.StackOverrides.VPCNestedTemplateURL != "" {
				url := installCfg.StackOverrides.VPCNestedTemplateURL
				updates["vpc_nested_template_url"] = &url
			}
			if installCfg.StackOverrides.RunnerNestedTemplateURL != "" {
				url := installCfg.StackOverrides.RunnerNestedTemplateURL
				updates["runner_nested_template_url"] = &url
			}
			if len(installCfg.StackOverrides.CustomNestedStacks) > 0 {
				updates["custom_nested_stacks"] = installCfg.StackOverrides.CustomNestedStacks
			}
		}

		if len(updates) > 0 && existing.InstallConfig != nil {
			db.WithContext(ctx).Model(&app.InstallConfig{}).
				Where("id = ?", existing.InstallConfig.ID).
				Updates(updates)
		} else if len(updates) > 0 {
			icParams := &installhelpers.CreateInstallConfigParams{
				ApprovalOption: app.InstallApprovalOption(installCfg.ApprovalOption),
			}
			if _, err := installHelpers.CreateInstallConfig(ctx, existing.ID, icParams); err != nil {
				return nil, fmt.Errorf("unable to create config for install %s: %w", installCfg.Name, err)
			}
		}
	}

	if len(installCfg.Labels) > 0 || len(existing.Labels) > 0 {
		syncLabels(ctx, db, existing.ID, installCfg.Labels, existing.Labels)
	}

	db.WithContext(ctx).Model(&app.Install{}).Where("id = ?", existing.ID).Updates(map[string]any{
		"metadata": map[string]string{"managed_by": ManagedByGitInstallConfig},
	})

	return &sync.InstallSyncResult{
		InstallID:   existing.ID,
		InstallName: existing.Name,
		Created:     false,
		Changed:     true,
		Diff:        d,
	}, nil
}

func syncLabels(ctx context.Context, db *gorm.DB, installID string, desired, current labels.Labels) {
	toSet := make(labels.Labels)
	for k, v := range desired {
		if cur, ok := current[k]; !ok || cur != v {
			toSet[k] = v
		}
	}

	if len(toSet) > 0 {
		db.WithContext(ctx).Model(&app.Install{}).Where("id = ?", installID).Update("labels", toSet)
	}
}

func existingToConfig(install *app.Install) *config.Install {
	cfg := &config.Install{
		Name:   install.Name,
		Labels: install.Labels,
	}

	if install.AWSAccount != nil {
		cfg.AWSAccount = &config.AWSAccount{Region: install.AWSAccount.Region}
	}
	if install.GCPAccount != nil {
		cfg.GCPAccount = &config.GCPAccount{
			ProjectID: install.GCPAccount.ProjectID,
			Region:    install.GCPAccount.Region,
		}
	}
	if install.AzureAccount != nil {
		cfg.AzureAccount = &config.AzureAccount{Location: install.AzureAccount.Location}
	}

	if install.InstallConfig != nil {
		cfg.ApprovalOption = config.InstallApprovalOption(install.InstallConfig.ApprovalOption)
		if install.InstallConfig.VPCNestedTemplateURL != nil ||
			install.InstallConfig.RunnerNestedTemplateURL != nil ||
			len(install.InstallConfig.CustomNestedStacks) > 0 {
			cfg.StackOverrides = &config.InstallStackOverrides{}
			if install.InstallConfig.VPCNestedTemplateURL != nil {
				cfg.StackOverrides.VPCNestedTemplateURL = *install.InstallConfig.VPCNestedTemplateURL
			}
			if install.InstallConfig.RunnerNestedTemplateURL != nil {
				cfg.StackOverrides.RunnerNestedTemplateURL = *install.InstallConfig.RunnerNestedTemplateURL
			}
			cfg.StackOverrides.CustomNestedStacks = install.InstallConfig.CustomNestedStacks
		}
		cfg.ComponentToggles = install.InstallConfig.ComponentToggles
	}

	return cfg
}

var _ func(ctx context.Context, db *gorm.DB, h *installhelpers.Helpers, appID string, i *config.Install) (*sync.InstallSyncResult, error) = SyncInstall
