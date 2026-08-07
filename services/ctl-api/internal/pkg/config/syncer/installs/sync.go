package installs

import (
	"context"
	"fmt"
	"maps"

	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
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
		req.AWSAccount = &installhelpers.CreateInstallAWSAccountParams{
			Region:    installCfg.AWSAccount.Region,
			AccountID: installCfg.AWSAccount.AccountID,
		}
	}
	if installCfg.GCPAccount != nil {
		req.GCPAccount = &installhelpers.CreateInstallGCPAccountParams{
			ProjectID: installCfg.GCPAccount.ProjectID,
			Region:    installCfg.GCPAccount.Region,
		}
	}
	if installCfg.AzureAccount != nil {
		req.AzureAccount = &installhelpers.CreateInstallAzureAccountParams{
			Location:       installCfg.AzureAccount.Location,
			SubscriptionID: installCfg.AzureAccount.SubscriptionID,
		}
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

	// Refused rather than ignored: nothing below writes the account, so without this
	// a changed identifier would diff forever and never converge. Shared with the CLI
	// syncer so the rule cannot differ by interface.
	if err := installCfg.CheckImmutableTargetAccount(upstream); err != nil {
		return nil, err
	}

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
		if err := syncInstallInputs(ctx, db, installHelpers, existing, definedInputs); err != nil {
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

	if len(installCfg.Labels) > 0 || len(existing.Labels) > 0 || len(existing.LabelTemplates) > 0 {
		if err := syncLabels(ctx, db, installHelpers, existing, installCfg.Labels); err != nil {
			return nil, fmt.Errorf("unable to sync labels for install %s: %w", installCfg.Name, err)
		}
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

// syncLabels merges config labels over the install's current labels so values
// set out of band survive a sync; keys removed from the config are only
// cleaned up when they were template-managed, since a stale rendered value
// would otherwise keep matching selectors forever. Templated values are stored
// on label_templates and rendered against install state, never written to the
// labels column as raw template text.
func syncLabels(ctx context.Context, db *gorm.DB, installHelpers *installhelpers.Helpers, existing *app.Install, desired labels.Labels) error {
	static, templated := desired.SplitTemplated()
	for key, tmpl := range templated {
		if err := render.ValidateTextTemplate(tmpl); err != nil {
			return fmt.Errorf("label %q template is invalid: %w", key, err)
		}
	}

	merged := make(labels.Labels, len(existing.Labels)+len(static))
	for k, v := range existing.Labels {
		merged[k] = v
	}
	for key := range existing.LabelTemplates {
		if _, ok := templated[key]; !ok {
			delete(merged, key)
		}
	}
	merged.Merge(static)

	updates := map[string]any{}
	if !maps.Equal(map[string]string(existing.LabelTemplates), map[string]string(templated)) {
		updates["label_templates"] = templated
	}
	if !maps.Equal(map[string]string(existing.Labels), map[string]string(merged)) {
		updates["labels"] = merged
	}
	if len(updates) == 0 {
		return nil
	}

	if err := db.WithContext(ctx).Model(&app.Install{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
		return fmt.Errorf("unable to update install labels: %w", err)
	}

	// Render immediately so a new template does not wait for the next state
	// change; unresolvable templates stay absent until state populates.
	if len(templated) > 0 {
		if err := installHelpers.RenderInstallLabels(ctx, existing.ID); err != nil {
			return fmt.Errorf("unable to render install label templates: %w", err)
		}
	}

	return nil
}

func existingToConfig(install *app.Install) *config.Install {
	cfg := &config.Install{
		Name:   install.Name,
		Labels: upstreamLabels(install),
	}

	// The target identifiers must be echoed back, otherwise a config that legitimately
	// declares them diffs against an upstream that never reports them and every sync
	// shows drift that no update can ever resolve. Kept in step with genCLIInstallConfig,
	// which does the same for the CLI's view of upstream.
	if install.AWSAccount != nil {
		cfg.AWSAccount = &config.AWSAccount{
			Region:    install.AWSAccount.Region,
			AccountID: install.CloudPlatformMetadata.TargetAccountID,
		}
	}
	// Azure and GCP already carry their identifier on the account record, so installs
	// created before CloudPlatformMetadata existed still round-trip.
	if install.GCPAccount != nil {
		cfg.GCPAccount = &config.GCPAccount{
			ProjectID: firstNonEmpty(
				install.CloudPlatformMetadata.TargetProjectID,
				install.GCPAccount.ProjectID,
			),
			Region: install.GCPAccount.Region,
		}
	}
	if install.AzureAccount != nil {
		cfg.AzureAccount = &config.AzureAccount{
			Location: install.AzureAccount.Location,
			SubscriptionID: firstNonEmpty(
				install.CloudPlatformMetadata.TargetSubscriptionID,
				install.AzureAccount.SubscriptionID,
			),
		}
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

// upstreamLabels echoes template text back for template-managed keys, so a
// config that declares a dynamic label diffs clean against its rendered value
// instead of showing drift on every sync.
func upstreamLabels(install *app.Install) labels.Labels {
	if len(install.LabelTemplates) == 0 {
		return install.Labels
	}

	merged := make(labels.Labels, len(install.Labels)+len(install.LabelTemplates))
	for k, v := range install.Labels {
		merged[k] = v
	}
	merged.Merge(install.LabelTemplates)
	return merged
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}

	return ""
}

var _ func(ctx context.Context, db *gorm.DB, h *installhelpers.Helpers, appID string, i *config.Install) (*sync.InstallSyncResult, error) = SyncInstall

// syncInstallInputs merges rather than replaces so values set out of band survive a sync.
func syncInstallInputs(
	ctx context.Context,
	db *gorm.DB,
	installHelpers *installhelpers.Helpers,
	existing *app.Install,
	definedInputs map[string]string,
) error {
	pinned, err := installHelpers.GetPinnedAppInputConfig(ctx, existing.AppID, existing.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get pinned app input config: %w", err)
	}
	if pinned == nil {
		return fmt.Errorf("no app input config on app config %s", existing.AppConfigID)
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := installhelpers.LockInstallInputs(ctx, tx, existing.ID); err != nil {
			return err
		}

		var latest app.InstallInputs
		if err := tx.WithContext(ctx).
			Where(app.InstallInputs{InstallID: existing.ID}).
			Order("created_at DESC").
			Limit(1).
			Find(&latest).Error; err != nil {
			return fmt.Errorf("unable to get install inputs: %w", err)
		}

		merged := make(pgtype.Hstore, len(latest.Values)+len(definedInputs))
		for k, v := range latest.Values {
			merged[k] = v
		}
		for k, v := range definedInputs {
			val := v
			merged[k] = &val
		}

		installInputs := app.InstallInputs{
			InstallID:        existing.ID,
			AppInputConfigID: pinned.ID,
			Values:           merged,
		}
		if err := tx.WithContext(ctx).Create(&installInputs).Error; err != nil {
			return err
		}

		// nothing else on this path invalidates state
		return installHelpers.MarkInstallStatePartialsStale(ctx, tx, existing.ID, pkgstate.PartialInputs)
	})
}
