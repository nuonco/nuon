package activities

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type ResolveBranchPostDeployRunbooksInput struct {
	AppBranchID string `json:"app_branch_id"`
	// AppBranchConfigID is the config the run was planned from. Resolving the
	// branch's latest config instead would let a sync that lands mid-run change
	// which runbooks execute.
	AppBranchConfigID string            `json:"app_branch_config_id"`
	InstallIDs        []string          `json:"install_ids"`
	NewAppConfigID    string            `json:"new_app_config_id"`
	CreatedByID       string            `json:"created_by_id"`
	Inputs            map[string]string `json:"inputs,omitempty"`
}

type ResolvedPostDeployRunbook struct {
	RunbookID       string            `json:"runbook_id"`
	RunbookConfigID string            `json:"runbook_config_id"`
	RunbookName     string            `json:"runbook_name"`
	Inputs          map[string]string `json:"inputs,omitempty"`
}

type ResolveBranchPostDeployRunbooksOutput struct {
	Runbooks []ResolvedPostDeployRunbook `json:"runbooks"`
}

// ResolveBranchPostDeployRunbooks resolves the branch config's post-deploy
// runbooks into the exact runbook config version to run, validates the inputs
// once for the whole group, and ensures every target install has an
// InstallRunbook row to hang a run off.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) ResolveBranchPostDeployRunbooks(ctx context.Context, input *ResolveBranchPostDeployRunbooksInput) (*ResolveBranchPostDeployRunbooksOutput, error) {
	out := &ResolveBranchPostDeployRunbooksOutput{}

	var config app.AppBranchConfig
	err := a.db.WithContext(ctx).
		First(&config, "id = ?", input.AppBranchConfigID).Error
	// This activity runs for every install group of every branch run, including the
	// overwhelming majority that configure no post-deploy runbooks. A missing
	// config means "nothing to run", not a reason to fail an otherwise clean group.
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return out, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get app branch config: %w", err)
	}

	if len(config.PostDeployRunbookIDs) == 0 {
		return out, nil
	}

	// Activities carry no request auth context, so set the created-by/org the
	// InstallRunbook and InstallRunbookRun BeforeCreate hooks read; without this
	// the not-null created_by_id insert fails.
	ctx = cctx.SetAccountIDContext(ctx, input.CreatedByID)
	ctx = cctx.SetOrgIDContext(ctx, config.OrgID)

	// The install group's app-config update already reconciles runbooks, so this is
	// normally a no-op. It stays because a run whose group update partially failed
	// would otherwise have no InstallRunbook row to hang a run off.
	for _, installID := range input.InstallIDs {
		if err := a.installHelpers.ReconcileInstallRunbooks(ctx, installID); err != nil {
			return nil, fmt.Errorf("unable to reconcile install runbooks for %s: %w", installID, err)
		}
	}

	for _, runbookID := range config.PostDeployRunbookIDs {
		runbookConfig, pinned, err := a.resolveRunbookConfig(ctx, config.OrgID, runbookID, input.NewAppConfigID)
		if err != nil {
			return nil, err
		}
		if !pinned {
			a.l.Warn("no runbook config pinned to app config version; falling back to latest",
				zap.String("runbook_id", runbookID),
				zap.String("app_config_id", input.NewAppConfigID),
				zap.String("runbook_config_id", runbookConfig.ID),
			)
		}

		inputs, err := a.mergeAndValidateInputs(runbookConfig, input.Inputs)
		if err != nil {
			return nil, fmt.Errorf("runbook %s: %w", runbookConfig.Runbook.Name, err)
		}

		out.Runbooks = append(out.Runbooks, ResolvedPostDeployRunbook{
			RunbookID:       runbookID,
			RunbookConfigID: runbookConfig.ID,
			RunbookName:     runbookConfig.Runbook.Name,
			Inputs:          inputs,
		})
	}

	return out, nil
}

// resolveRunbookConfig returns the runbook config pinned to appConfigID, falling
// back to the runbook's latest config. The bool is true when the version-pinned
// config was found.
func (a *Activities) resolveRunbookConfig(ctx context.Context, orgID, runbookID, appConfigID string) (*app.RunbookConfig, bool, error) {
	base := func() *gorm.DB {
		return a.db.WithContext(ctx).
			Preload("Runbook").
			Preload("Inputs", func(tx *gorm.DB) *gorm.DB { return tx.Order("idx ASC") }).
			Where(app.RunbookConfig{RunbookID: runbookID, OrgID: orgID})
	}

	var cfg app.RunbookConfig
	if appConfigID != "" {
		if err := base().Where(app.RunbookConfig{AppConfigID: appConfigID}).First(&cfg).Error; err == nil {
			return &cfg, true, nil
		}
	}

	if err := base().Order("created_at DESC").First(&cfg).Error; err != nil {
		return nil, false, fmt.Errorf("runbook %s has no configurations: %w", runbookID, err)
	}
	return &cfg, false, nil
}

// mergeAndValidateInputs keeps only the inputs the runbook config declares — the
// branch injects VCS context opportunistically and undeclared names must not trip
// validation. Validation runs against the merged result rather than the raw branch
// inputs, so a required input satisfied by its default passes while one with
// neither default nor branch value fails before any install is touched.
func (a *Activities) mergeAndValidateInputs(runbookConfig *app.RunbookConfig, supplied map[string]string) (map[string]string, error) {
	declared := make(map[string]struct{}, len(runbookConfig.Inputs))
	for _, inp := range runbookConfig.Inputs {
		declared[inp.Name] = struct{}{}
	}

	kept := make(map[string]*string, len(supplied))
	for name, val := range supplied {
		if _, ok := declared[name]; ok {
			v := val
			kept[name] = &v
		}
	}

	merged := runbookshelpers.MergeRunbookInputDefaults(runbookConfig, kept)
	if err := a.runbooksHelpers.ValidateRunbookInputs(runbookConfig, merged); err != nil {
		return nil, err
	}

	out := make(map[string]string, len(merged))
	for k, v := range merged {
		if v != nil {
			out[k] = *v
		}
	}
	return out, nil
}
