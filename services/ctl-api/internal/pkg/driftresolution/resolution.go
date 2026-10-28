package driftresolution

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildselection"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/inputstate"
)

type TargetChanged string

func (e TargetChanged) Error() string { return string(e) }

func ApplyTarget(wf *app.Workflow, step *app.WorkflowStep) *app.DriftResolutionTarget {
	if wf.OwnerType != "installs" || wf.Metadata[app.WorkflowMetadataKeyResolveDrift] == nil ||
		*wf.Metadata[app.WorkflowMetadataKeyResolveDrift] != "true" || wf.Request == nil || wf.Request.DriftResolution == nil ||
		step.QueueSignal == nil || step.QueueSignal.Signal == nil || step.ExecutionType == app.WorkflowStepExecutionTypeSkipped {
		return nil
	}
	target := wf.Request.DriftResolution
	switch step.QueueSignal.Signal.Type() {
	case "component-deploy-apply-plan", "component-teardown-apply-plan", "reprovision-sandbox-apply-plan", "generate-install-stack-version":
		return target
	case "component-sync-image":
		if target.Kind == "image" {
			return target
		}
	}
	return nil
}

func LockInstall(ctx context.Context, tx *gorm.DB, installID string) (*app.Install, error) {
	var install app.Install
	err := tx.WithContext(ctx).InstanceSet(views.DisableViewsKey, true).Select("id", "app_config_id").
		Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Install{ID: installID}).First(&install).Error
	return &install, err
}

func InputsDigest(ctx context.Context, tx *gorm.DB, installID, configID string, target *app.DriftResolutionTarget) (string, string, error) {
	var cfg app.AppConfig
	if err := tx.WithContext(ctx).Preload("InputConfig.AppInputs").Where(app.AppConfig{ID: configID}).First(&cfg).Error; err != nil {
		return "", "", err
	}
	var inputs app.InstallInputs
	if err := tx.WithContext(ctx).Where(app.InstallInputs{InstallID: installID}).
		Order("created_at DESC, id DESC").Limit(1).Find(&inputs).Error; err != nil {
		return "", "", err
	}
	resolved := &state.State{Inputs: &state.InputsState{}}
	if inputs.ID != "" {
		resolved.Inputs = inputstate.Resolve(&inputs, &cfg, false)
	}
	if target.BuildAppConfigID != "" && target.BuildAppConfigID != configID {
		var buildCfg app.AppConfig
		if err := tx.WithContext(ctx).Unscoped().Preload("InputConfig.AppInputs").
			Where(app.AppConfig{ID: target.BuildAppConfigID}).First(&buildCfg).Error; err != nil {
			return "", "", err
		}
		inputstate.FillDefaults(resolved, &buildCfg)
	}
	return inputstate.Digest(target.InputNames, resolved.Inputs), inputs.ID, nil
}

func AppliedRef(ctx context.Context, db *gorm.DB, target *app.DriftResolutionTarget) (app.AppConfigRef, error) {
	query := db.WithContext(ctx).Select("app_config_ref")
	switch target.Kind {
	case "stack":
		var stack app.InstallStack
		err := query.Where(app.InstallStack{ID: target.ResourceID}).First(&stack).Error
		return stack.AppConfigRef, err
	case "sandbox":
		var sandbox app.InstallSandbox
		err := query.Where(app.InstallSandbox{ID: target.ResourceID}).First(&sandbox).Error
		return sandbox.AppConfigRef, err
	default:
		var component app.InstallComponent
		err := query.Where(app.InstallComponent{ID: target.ResourceID}).First(&component).Error
		return component.AppConfigRef, err
	}
}

func CheckTarget(ctx context.Context, tx *gorm.DB, install *app.Install, target *app.DriftResolutionTarget) error {
	if target.AppConfigID != install.AppConfigID {
		diff, err := configdiff.ComputeInstallConfigDiff(ctx, tx, target.AppConfigID, install.AppConfigID)
		if err != nil {
			return err
		}
		switch target.Kind {
		case "stack":
			if diff.StackChanged {
				return TargetChanged("the selected stack configuration changed; resolve drift again")
			}
		case "sandbox":
			if diff.SandboxChanged || diff.SandboxBuildChanged {
				return TargetChanged("the selected sandbox configuration changed; resolve drift again")
			}
		default:
			for _, entries := range [][]app.ComponentDiffEntry{diff.Changed, diff.Removed, diff.Added} {
				for _, entry := range entries {
					if entry.ComponentID == target.ComponentID {
						return TargetChanged("the selected component configuration changed; resolve drift again")
					}
				}
			}
		}
	}
	if target.ComponentID != "" {
		var cfg app.ComponentConfigConnection
		if err := tx.WithContext(ctx).Where(app.ComponentConfigConnection{AppConfigID: install.AppConfigID, ComponentID: target.ComponentID}).First(&cfg).Error; err != nil {
			return err
		}
		build, err := buildselection.Component(ctx, tx, cfg.ID)
		if err != nil {
			return err
		}
		if build == nil || build.ID != target.BuildID {
			return TargetChanged("the selected component build changed; resolve drift again")
		}
	}
	digest, _, err := InputsDigest(ctx, tx, install.ID, install.AppConfigID, target)
	if err != nil {
		return err
	}
	if digest != target.InputsDigest {
		return TargetChanged("input values used by this resource changed; resolve drift again")
	}
	ref, err := AppliedRef(ctx, tx, target)
	if err != nil {
		return err
	}
	if ref.AppliedConfigByID != target.AppliedByID || (ref.AppliedConfigAt == nil) != (target.AppliedAt == nil) ||
		(ref.AppliedConfigAt != nil && !ref.AppliedConfigAt.Equal(*target.AppliedAt)) {
		return TargetChanged("another deployment applied to this resource; resolve drift again")
	}
	return nil
}
