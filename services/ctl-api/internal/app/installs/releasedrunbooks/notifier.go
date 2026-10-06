package releasedrunbooks

import (
	"context"
	"fmt"
	"strconv"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// Notifier starts the branch's post-deploy runbooks after an install the group
// already released finishes successfully. The group step has returned, so these
// runs have no completion callback back into it.
type Notifier struct {
	db      *gorm.DB
	helpers *runbookshelpers.Helpers
	l       *zap.Logger
}

type Params struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Helpers *runbookshelpers.Helpers
	L       *zap.Logger
}

func NewNotifier(params Params) statusactivities.ReleasedInstallNotifier {
	l := params.L
	if l == nil {
		l = zap.NewNop()
	}
	return &Notifier{db: params.DB, helpers: params.Helpers, l: l}
}

func (n *Notifier) ReleasedInstallSucceeded(ctx context.Context, workflowID, appBranchRunID, installID string) {
	l := n.l.With(
		zap.String("workflow_id", workflowID),
		zap.String("app_branch_run_id", appBranchRunID),
		zap.String("install_id", installID),
	)
	if appBranchRunID == "" || installID == "" {
		return
	}

	var run app.AppBranchRun
	if err := n.db.WithContext(ctx).First(&run, "id = ?", appBranchRunID).Error; err != nil {
		l.Warn("released install runbooks: unable to get app branch run", zap.Error(err))
		return
	}
	if run.IsPreview() {
		return
	}

	var config app.AppBranchConfig
	if err := n.db.WithContext(ctx).First(&config, "id = ?", run.AppBranchConfigID).Error; err != nil {
		l.Warn("released install runbooks: unable to get app branch config", zap.Error(err))
		return
	}
	if len(config.PostDeployRunbookIDs) == 0 {
		return
	}

	ctx = cctx.SetAccountIDContext(ctx, run.CreatedByID)
	ctx = cctx.SetOrgIDContext(ctx, config.OrgID)
	inputs := branchRunbookInputs(&run)
	appConfigID := run.AppConfigID

	for _, runbookID := range config.PostDeployRunbookIDs {
		if err := n.start(ctx, l, &run, installID, runbookID, appConfigID, inputs); err != nil {
			l.Warn("released install runbooks: unable to start runbook",
				zap.String("runbook_id", runbookID),
				zap.Error(err),
			)
		}
	}
}

func (n *Notifier) start(ctx context.Context, l *zap.Logger, run *app.AppBranchRun, installID, runbookID, appConfigID string, inputs map[string]string) error {
	runbookConfig, err := n.resolveRunbookConfig(ctx, run.OrgID, runbookID, appConfigID)
	if err != nil {
		return err
	}

	row := app.InstallRunbook{InstallID: installID, RunbookID: runbookID}
	if err := n.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
		return fmt.Errorf("unable to ensure install runbook: %w", err)
	}
	var installRunbook app.InstallRunbook
	if err := n.db.WithContext(ctx).
		Where(app.InstallRunbook{InstallID: installID, RunbookID: runbookID}).
		First(&installRunbook).Error; err != nil {
		return fmt.Errorf("unable to get install runbook: %w", err)
	}

	key := fmt.Sprintf("branch-run:%s:%s:%s:0", run.ID, installID, runbookID)
	if _, err := n.helpers.TriggerRunbookRun(ctx, runbookshelpers.TriggerRunbookRunRequest{
		InstallRunbookID: installRunbook.ID,
		RunbookConfigID:  runbookConfig.ID,
		TriggeredByID:    run.CreatedByID,
		Inputs:           declaredRunbookInputs(runbookConfig, inputs),
		IdempotencyKey:   &key,
	}); err != nil {
		return err
	}
	l.Info("started post-deploy runbook for released install", zap.String("runbook_id", runbookID))
	return nil
}

func (n *Notifier) resolveRunbookConfig(ctx context.Context, orgID, runbookID, appConfigID string) (*app.RunbookConfig, error) {
	base := func() *gorm.DB {
		return n.db.WithContext(ctx).
			Preload("Runbook").
			Preload("Inputs", func(tx *gorm.DB) *gorm.DB { return tx.Order("idx ASC") }).
			Where(app.RunbookConfig{RunbookID: runbookID, OrgID: orgID})
	}
	var cfg app.RunbookConfig
	if appConfigID != "" {
		if err := base().Where(app.RunbookConfig{AppConfigID: appConfigID}).First(&cfg).Error; err == nil {
			return &cfg, nil
		}
	}
	if err := base().Order("created_at DESC").First(&cfg).Error; err != nil {
		return nil, fmt.Errorf("runbook %s has no configurations: %w", runbookID, err)
	}
	return &cfg, nil
}

func declaredRunbookInputs(cfg *app.RunbookConfig, supplied map[string]string) map[string]string {
	if cfg == nil || len(cfg.Inputs) == 0 || len(supplied) == 0 {
		return nil
	}
	declared := map[string]struct{}{}
	for _, input := range cfg.Inputs {
		declared[input.Name] = struct{}{}
	}
	kept := map[string]string{}
	for name, value := range supplied {
		if _, ok := declared[name]; ok {
			kept[name] = value
		}
	}
	if len(kept) == 0 {
		return nil
	}
	return kept
}

func branchRunbookInputs(run *app.AppBranchRun) map[string]string {
	inputs := map[string]string{}
	if run.HeadSHA != "" {
		inputs["commit_sha"] = run.HeadSHA
	}
	if run.BaseBranch != "" {
		inputs["base_branch"] = run.BaseBranch
	}
	if run.PRNumber != nil {
		inputs["pr_number"] = strconv.Itoa(*run.PRNumber)
	}
	return inputs
}
