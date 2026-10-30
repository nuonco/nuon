package installs

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

const ManagedByGitInstallConfig = "nuon/git/install-config"

func SyncInstall(ctx context.Context, db *gorm.DB, installHelpers *installhelpers.Helpers, appID string, install *config.Install, beforeApply ...func(*gorm.DB, *sync.InstallSyncResult) error) (*sync.InstallSyncResult, error) {
	if install == nil {
		return nil, fmt.Errorf("install config is nil")
	}
	if install.Name == "" {
		return nil, fmt.Errorf("install config name is required")
	}
	if err := install.Validate(); err != nil {
		return nil, err
	}

	var application app.App
	if err := db.WithContext(ctx).
		Preload("Org").
		Where(app.App{ID: appID}).
		First(&application).Error; err != nil {
		return nil, fmt.Errorf("unable to load app %s for install sync: %w", appID, err)
	}
	if application.Org.Features[string(app.OrgFeatureDisableAppSync)] && install.AppBranch == "" {
		return nil, fmt.Errorf("install config %q must set app_branch when disable-app-sync is enabled", install.Name)
	}

	var existing app.Install
	err := db.WithContext(ctx).
		Preload("InstallConfig").
		Preload("AWSAccount").
		Preload("GCPAccount").
		Preload("AzureAccount").
		Preload("CloudConnection").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC, id DESC").Limit(1)
		}).
		Where(app.Install{AppID: appID, Name: install.Name}).
		First(&existing).Error

	var cloudConnectionID string
	if install.CloudConnection != "" {
		connection, resolveErr := resolveCloudConnection(ctx, db, application.OrgID, install)
		if resolveErr != nil {
			return nil, resolveErr
		}
		if install.CloudFormation != nil && connection.Platform != app.CloudPlatformAWS {
			return nil, fmt.Errorf("install %q: cloudformation requires an AWS cloud_connection", install.Name)
		}
		if existing.CloudConnectionID != nil && *existing.CloudConnectionID != connection.ID {
			return nil, fmt.Errorf("refusing to change cloud_connection on existing install %q: the cloud connection is immutable after creation", install.Name)
		}
		install.CloudConnection = connection.Name
		cloudConnectionID = connection.ID
	}

	var appBranchID string
	if install.AppBranch != "" {
		branch, err := resolveAppBranch(ctx, db, appID, install.AppBranch)
		if err != nil {
			return nil, err
		}
		appBranchID = branch.ID
		install.AppBranch = branch.Name
		if err := validateAppBranchGroup(ctx, db, branch.ID, install.AppBranchGroup); err != nil {
			return nil, err
		}
	}

	if err == gorm.ErrRecordNotFound {
		return createInstall(ctx, installHelpers, appID, install, appBranchID, cloudConnectionID, beforeApply)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to look up install %s: %w", install.Name, err)
	}

	return updateInstall(ctx, db, installHelpers, &existing, install, appBranchID, cloudConnectionID, beforeApply)
}

// resolveCloudConnection resolves a cloud_connection reference from an install
// config by ID first, then by name within the org.
func resolveCloudConnection(ctx context.Context, db *gorm.DB, orgID string, installCfg *config.Install) (*app.CloudConnection, error) {
	var connection app.CloudConnection
	err := db.WithContext(ctx).
		Where(app.CloudConnection{ID: installCfg.CloudConnection, OrgID: orgID}).
		First(&connection).Error
	if err == nil {
		return &connection, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to look up cloud connection %q: %w", installCfg.CloudConnection, err)
	}

	err = db.WithContext(ctx).
		Where(app.CloudConnection{OrgID: orgID, Name: installCfg.CloudConnection}).
		First(&connection).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("cloud connection %q was not found in org %s", installCfg.CloudConnection, orgID),
			Description: "cloud_connection must reference a cloud connection in the org by name or ID",
		}
	}
	if err != nil {
		return nil, fmt.Errorf("unable to look up cloud connection %q: %w", installCfg.CloudConnection, err)
	}
	return &connection, nil
}

func resolveAppBranch(ctx context.Context, db *gorm.DB, appID, ref string) (*app.AppBranch, error) {
	var branch app.AppBranch
	err := db.WithContext(ctx).
		Where(app.AppBranch{ID: ref, AppID: appID}).
		First(&branch).Error
	if err == nil {
		return &branch, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("unable to resolve app branch %q by ID: %w", ref, err)
	}

	err = db.WithContext(ctx).
		Where(app.AppBranch{AppID: appID, Name: ref}).
		First(&branch).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("app branch %q was not found on app %s", ref, appID)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to resolve app branch %q by name: %w", ref, err)
	}
	return &branch, nil
}

func validateAppBranchGroup(ctx context.Context, db *gorm.DB, branchID, group string) error {
	if group == "" {
		return nil
	}
	var branchConfig app.AppBranchConfig
	if err := db.WithContext(ctx).
		Preload("InstallGroups").
		Where(app.AppBranchConfig{AppBranchID: branchID}).
		Order("created_at DESC, id DESC").
		First(&branchConfig).Error; err != nil {
		return fmt.Errorf("unable to load app branch groups: %w", err)
	}
	for _, candidate := range branchConfig.InstallGroups {
		if candidate.Name == group {
			return nil
		}
	}
	return fmt.Errorf("app branch group %q was not found on branch %s", group, branchID)
}

func createInstall(ctx context.Context, installHelpers *installhelpers.Helpers, appID string, installCfg *config.Install, appBranchID, cloudConnectionID string, beforeApply []func(*gorm.DB, *sync.InstallSyncResult) error) (*sync.InstallSyncResult, error) {
	inputs := make(map[string]*string)
	for k, v := range installCfg.FlattenedInputs() {
		val := v
		inputs[k] = &val
	}

	req := &installhelpers.CreateInstallParams{
		Name:              installCfg.Name,
		Inputs:            inputs,
		Labels:            installCfg.Labels,
		CloudConnectionID: cloudConnectionID,
		Metadata: installhelpers.InstallMetadata{
			ManagedBy: ManagedByGitInstallConfig,
		},
		AppBranchID:    appBranchID,
		AppBranchGroup: installCfg.AppBranchGroup,
		AfterCreate: func(tx *gorm.DB, install *app.Install) error {
			for _, hook := range beforeApply {
				if err := hook(tx, &sync.InstallSyncResult{
					InstallID: install.ID, InstallName: install.Name, Created: true, Changed: true,
				}); err != nil {
					return err
				}
			}
			return nil
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
		installCfg.CloudFormation != nil ||
		len(installCfg.ComponentToggles) > 0 || installCfg.Telemetry != nil ||
		installCfg.ActionTimeout != "" {
		params := build.InstallConfigInputFromConfig(installCfg)
		req.InstallConfig = &params
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

func updateInstall(ctx context.Context, db *gorm.DB, installHelpers *installhelpers.Helpers, existing *app.Install, installCfg *config.Install, appBranchID, cloudConnectionID string, beforeApply []func(*gorm.DB, *sync.InstallSyncResult) error) (*sync.InstallSyncResult, error) {
	upstream := existingToConfig(existing)
	for key := range upstream.Labels {
		if _, declared := installCfg.Labels[key]; !declared && existing.LabelTemplates[key] == "" {
			delete(upstream.Labels, key)
		}
	}
	if installCfg.CloudConnection == "" {
		upstream.CloudConnection = ""
	}
	if installCfg.StackOverrides == nil {
		upstream.StackOverrides = nil
	} else if upstream.StackOverrides != nil {
		if installCfg.StackOverrides.VPCNestedTemplateURL == "" {
			upstream.StackOverrides.VPCNestedTemplateURL = ""
		}
		if installCfg.StackOverrides.RunnerNestedTemplateURL == "" {
			upstream.StackOverrides.RunnerNestedTemplateURL = ""
		}
		if len(installCfg.StackOverrides.CustomNestedStacks) == 0 {
			upstream.StackOverrides.CustomNestedStacks = nil
		}
	}

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

	for _, hook := range beforeApply {
		if err := hook(db, &sync.InstallSyncResult{
			InstallID: existing.ID, InstallName: existing.Name, Changed: true, Diff: d,
			AppBranchID: appBranchID,
			AppBranchChanged: appBranchID != "" && (!existing.AppBranchID.Valid || existing.AppBranchID.String != appBranchID ||
				(installCfg.AppBranchGroup != "" && existing.AppBranchGroup != installCfg.AppBranchGroup)),
		}); err != nil {
			return nil, err
		}
	}

	definedInputs := installCfg.FlattenedInputs()
	if len(definedInputs) > 0 {
		if err := syncInstallInputs(ctx, db, installHelpers, existing, definedInputs); err != nil {
			return nil, fmt.Errorf("unable to update inputs for install %s: %w", installCfg.Name, err)
		}
		// syncLabels only renders when the label config itself changed; an
		// inputs-only sync still moves state that templates can reference.
		if len(existing.LabelTemplates) > 0 {
			if err := installHelpers.RenderInstallLabels(ctx, existing.ID); err != nil {
				return nil, fmt.Errorf("unable to render install label templates for install %s: %w", installCfg.Name, err)
			}
		}
	}

	params := build.InstallConfigInputFromConfig(installCfg)
	installConfig := build.InstallConfig(params, existing.ID)
	var fields []string
	if installConfig.ApprovalOption != "" {
		fields = append(fields, "approval_option")
	}
	if installConfig.TelemetryEnabled != nil {
		fields = append(fields, "telemetry_enabled")
	}
	if installConfig.VPCNestedTemplateURL != nil {
		fields = append(fields, "vpc_nested_template_url")
	}
	if installConfig.RunnerNestedTemplateURL != nil {
		fields = append(fields, "runner_nested_template_url")
	}
	if len(installConfig.CustomNestedStacks) > 0 {
		fields = append(fields, "custom_nested_stacks")
	}
	if installConfig.AWSStackParameters != nil {
		fields = append(fields, "aws_stack_parameters")
	}
	if installConfig.AWSNestedStackParameters != nil {
		fields = append(fields, "aws_nested_stack_parameters")
	}
	if installConfig.RunnerImageAutoUpdate != nil {
		fields = append(fields, "runner_image_auto_update")
	}
	if installConfig.RunnerImageID != nil {
		fields = append(fields, "runner_image_id")
	}
	if installConfig.ActionTimeout != nil {
		fields = append(fields, "action_timeout")
	}
	if len(fields) > 0 {
		if existing.InstallConfig != nil {
			if err := db.WithContext(ctx).Model(&app.InstallConfig{}).
				Where(app.InstallConfig{ID: existing.InstallConfig.ID, InstallID: existing.ID, OrgID: existing.OrgID}).
				Select(fields).Updates(installConfig).Error; err != nil {
				return nil, fmt.Errorf("unable to update config for install %s: %w", installCfg.Name, err)
			}
		} else {
			if _, err := installHelpers.CreateInstallConfig(ctx, existing.ID, &params); err != nil {
				return nil, fmt.Errorf("unable to create config for install %s: %w", installCfg.Name, err)
			}
		}
	}

	if len(installCfg.Labels) > 0 || len(existing.Labels) > 0 || len(existing.LabelTemplates) > 0 {
		if err := syncLabels(ctx, db, installHelpers, existing, installCfg.Labels); err != nil {
			return nil, fmt.Errorf("unable to sync labels for install %s: %w", installCfg.Name, err)
		}
	}

	if cloudConnectionID != "" && existing.CloudConnectionID == nil {
		if err := db.WithContext(ctx).Model(&app.Install{}).
			Where(app.Install{ID: existing.ID, OrgID: existing.OrgID}).
			Update("cloud_connection_id", cloudConnectionID).Error; err != nil {
			return nil, fmt.Errorf("unable to update cloud connection for install %s: %w", installCfg.Name, err)
		}
		existing.CloudConnectionID = &cloudConnectionID
	}

	appBranchChanged := appBranchID != "" && (!existing.AppBranchID.Valid || existing.AppBranchID.String != appBranchID)
	appBranchGroupChanged := false
	if appBranchID != "" {
		var candidate app.Install
		if err := db.WithContext(ctx).First(&candidate, "id = ?", existing.ID).Error; err != nil {
			return nil, fmt.Errorf("unable to reload install for app branch assignment: %w", err)
		}
		candidate.AppBranchGroup = installCfg.AppBranchGroup
		candidate.AppBranchGroupAssignmentSource = ""
		if installCfg.AppBranchGroup != "" {
			candidate.AppBranchGroupAssignmentSource = app.InstallAppBranchGroupAssignmentSourceExplicit
		}
		groups, err := appshelpers.LatestConfigInstallGroupsWithDB(ctx, db, appBranchID)
		if err != nil {
			return nil, err
		}
		if err := appshelpers.ValidateInstallSingleGroup(groups, &candidate); err != nil {
			return nil, err
		}
		group, source, err := appshelpers.ResolveInstallGroupAssignment(groups, &candidate)
		if err != nil {
			return nil, err
		}
		if group == nil {
			return nil, appshelpers.NoMatchingInstallGroupError(&candidate)
		}
		appBranchGroupChanged = existing.AppBranchGroup != group.Name || existing.AppBranchGroupAssignmentSource != source
		if err := appshelpers.SetInstallAppBranchGroupAssignmentWithDB(ctx, db, existing.ID, appBranchID, group.Name, source); err != nil {
			return nil, fmt.Errorf("unable to update install app branch connection: %w", err)
		}
	}

	db.WithContext(ctx).Model(&app.Install{}).Where("id = ?", existing.ID).Updates(map[string]any{
		"metadata": map[string]string{"managed_by": ManagedByGitInstallConfig},
	})

	return &sync.InstallSyncResult{
		InstallID:        existing.ID,
		InstallName:      existing.Name,
		Created:          false,
		Changed:          true,
		Diff:             d,
		AppBranchChanged: appBranchChanged || appBranchGroupChanged,
		AppBranchID:      appBranchID,
	}, nil
}

// syncLabels merges config labels over current labels so out-of-band values
// survive a sync. Removed keys are only cleaned up when template-managed — a
// stale rendered value would otherwise keep matching selectors forever.
// Template text is never written to the labels column.
func syncLabels(ctx context.Context, db *gorm.DB, installHelpers *installhelpers.Helpers, existing *app.Install, desired labels.Labels) error {
	// Default-label keys are owned by the app config; an install config echoing
	// the rendered value or the default itself round-trips, anything else fails.
	if len(existing.AppDefaultLabels) > 0 && len(desired) > 0 {
		pruned := make(labels.Labels, len(desired))
		for key, val := range desired {
			def, isDefault := existing.AppDefaultLabels[key]
			if !isDefault {
				pruned[key] = val
				continue
			}
			if val == existing.Labels[key] || val == def {
				continue
			}
			return fmt.Errorf("label %q is a default label; edit default_labels in the app config", key)
		}
		desired = pruned
	}

	static, templated := desired.SplitTemplated()
	for key, tmpl := range templated {
		if err := render.ValidateTextTemplate(tmpl); err != nil {
			return fmt.Errorf("label %q template is invalid: %w", key, err)
		}
	}

	newTemplates := make(labels.Labels, len(templated)+len(existing.AppDefaultLabels))
	for k, v := range templated {
		newTemplates[k] = v
	}

	merged := make(labels.Labels, len(existing.Labels)+len(static))
	for k, v := range existing.Labels {
		merged[k] = v
	}
	for key, tmpl := range existing.LabelTemplates {
		if _, isDefault := existing.AppDefaultLabels[key]; isDefault {
			newTemplates[key] = tmpl
			continue
		}
		if _, ok := templated[key]; !ok {
			delete(merged, key)
		}
	}
	merged.Merge(static)

	updates := map[string]any{}
	if !maps.Equal(map[string]string(existing.LabelTemplates), map[string]string(newTemplates)) {
		updates["label_templates"] = newTemplates
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

	// Render now so a new template doesn't wait for the next state change.
	if len(templated) > 0 {
		if err := installHelpers.RenderInstallLabels(ctx, existing.ID); err != nil {
			return fmt.Errorf("unable to render install label templates: %w", err)
		}
	}

	return nil
}

func existingToConfig(install *app.Install) *config.Install {
	var appBranchName string
	if install.AppBranch != nil {
		appBranchName = install.AppBranch.Name
	}
	cfg := &config.Install{
		Name:      install.Name,
		AppBranch: appBranchName,
		Labels:    maps.Clone(upstreamLabels(install)),
	}
	if install.AppBranchGroupAssignmentSource == app.InstallAppBranchGroupAssignmentSourceExplicit {
		cfg.AppBranchGroup = install.AppBranchGroup
	}
	if install.CurrentInstallInputs != nil {
		values := make(map[string]string, len(install.CurrentInstallInputs.Values))
		for key, value := range install.CurrentInstallInputs.Values {
			if value != nil {
				values[key] = *value
			}
		}
		cfg.InputGroups = []config.InputGroup{{Inputs: values}}
	}
	if install.CloudConnection != nil {
		cfg.CloudConnection = install.CloudConnection.Name
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
		if install.InstallConfig.TelemetryEnabled != nil {
			cfg.Telemetry = &config.InstallTelemetry{Enabled: install.InstallConfig.TelemetryEnabled}
		}
		if install.InstallConfig.VPCNestedTemplateURL != nil ||
			install.InstallConfig.RunnerNestedTemplateURL != nil ||
			len(install.InstallConfig.CustomNestedStacks) > 0 ||
			install.InstallConfig.RunnerImageAutoUpdate != nil ||
			install.InstallConfig.RunnerImageID != nil {
			cfg.StackOverrides = &config.InstallStackOverrides{}
			if install.InstallConfig.VPCNestedTemplateURL != nil {
				cfg.StackOverrides.VPCNestedTemplateURL = *install.InstallConfig.VPCNestedTemplateURL
			}
			if install.InstallConfig.RunnerNestedTemplateURL != nil {
				cfg.StackOverrides.RunnerNestedTemplateURL = *install.InstallConfig.RunnerNestedTemplateURL
			}
			cfg.StackOverrides.CustomNestedStacks = install.InstallConfig.CustomNestedStacks
			cfg.StackOverrides.RunnerImageAutoUpdate = install.InstallConfig.RunnerImageAutoUpdate
			if install.InstallConfig.RunnerImageID != nil {
				cfg.StackOverrides.RunnerImageID = *install.InstallConfig.RunnerImageID
			}
		}
		cfg.CloudFormation = &config.CloudFormation{
			Root:         config.CloudFormationStack{Parameters: install.InstallConfig.AWSStackParameters},
			NestedStacks: install.InstallConfig.AWSNestedStackParameters,
		}
		cfg.ComponentToggles = install.InstallConfig.ComponentToggles
		if install.InstallConfig.ActionTimeout != nil {
			cfg.ActionTimeout = config.FormatActionTimeout(*install.InstallConfig.ActionTimeout)
		}
	}

	return cfg
}

// upstreamLabels echoes template text for template-managed keys so a dynamic
// label diffs clean instead of showing drift against its rendered value, and
// strips app-default keys, which install configs never declare.
func upstreamLabels(install *app.Install) labels.Labels {
	if len(install.LabelTemplates) == 0 && len(install.AppDefaultLabels) == 0 {
		return install.Labels
	}

	merged := make(labels.Labels, len(install.Labels)+len(install.LabelTemplates))
	for k, v := range install.Labels {
		merged[k] = v
	}
	merged.Merge(install.LabelTemplates)
	for key := range install.AppDefaultLabels {
		delete(merged, key)
	}
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

var _ func(ctx context.Context, db *gorm.DB, h *installhelpers.Helpers, appID string, i *config.Install, hooks ...func(*gorm.DB, *sync.InstallSyncResult) error) (*sync.InstallSyncResult, error) = SyncInstall

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
		return tx.WithContext(ctx).Create(&installInputs).Error
	})
}
