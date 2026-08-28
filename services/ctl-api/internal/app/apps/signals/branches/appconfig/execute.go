package appconfig

import (
	"encoding/json"
	"fmt"
	"strconv"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	preCompiled := s.AppConfigID != ""

	var (
		run       *app.AppBranchRun
		commitSHA string
		err       error
	)
	if preCompiled {
		run, err = activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
		if err != nil {
			return fmt.Errorf("unable to get app branch run: %w", err)
		}
	} else {
		run, err = activities.AwaitGetAppBranchRunWithCommitByRunID(ctx, s.RunID)
		if err != nil {
			return fmt.Errorf("unable to get app branch run with commit: %w", err)
		}
		commitSHA = run.VCSConnectionCommit.SHA
	}

	// Create log stream for this run
	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		AppBranchRunID: s.RunID,
	})
	if err != nil {
		l.Warn("unable to create log stream, continuing without it", "error", err)
	}

	if logStream != nil {
		if err := activities.AwaitUpdateAppBranchRunLogStream(ctx, activities.UpdateAppBranchRunLogStreamRequest{
			Req: &activities.UpdateAppBranchRunLogStreamInput{
				RunID:       s.RunID,
				LogStreamID: logStream.ID,
			},
		}); err != nil {
			l.Warn("unable to update run with log stream ID", "error", err)
		}
	}

	// Ensure log stream is closed when we're done
	closeLogStream := func() {
		if logStream == nil {
			return
		}
		if err := activities.AwaitCloseLogStream(ctx, activities.CloseLogStreamRequest{
			LogStreamID: logStream.ID,
		}); err != nil {
			l.Warn("unable to close log stream", "error", err)
		}
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		closeLogStream()
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	if len(branch.Configs) == 0 {
		closeLogStream()
		return fmt.Errorf("app branch has no config")
	}

	var vcsConfigID string
	if cfg := branch.Configs[0].ConnectedGithubVCSConfig; cfg != nil {
		vcsConfigID = cfg.ID
	} else if cfg := branch.Configs[0].PublicGitVCSConfig; cfg != nil {
		vcsConfigID = cfg.ID
	} else if !preCompiled {
		closeLogStream()
		return fmt.Errorf("app branch has no VCS config")
	}

	if preCompiled {
		return s.syncAndFinalize(ctx, finalizeParams{
			run:         run,
			branch:      branch,
			appConfigID: s.AppConfigID,
			vcsConfigID: vcsConfigID,
			isPreview:   run.IsPreview(),
		}, closeLogStream)
	}

	cloneResult, err := activities.LocalAwaitCloneRepo(ctx, activities.CloneRepoRequest{
		RunID:       s.RunID,
		VcsConfigID: vcsConfigID,
		CommitSHA:   commitSHA,
	})
	if err != nil {
		closeLogStream()
		return fmt.Errorf("unable to clone repo: %w", err)
	}

	sourceDir := cloneResult.SourceDir

	l.Info("repo cloned successfully",
		"app_branch_id", branch.ID,
		"commit_sha", commitSHA,
		"source_dir", sourceDir)

	intermediateConfig, err := activities.LocalAwaitFetchIntermediateConfig(ctx, activities.FetchIntermediateConfigRequest{
		SourceDir: sourceDir,
	})
	if err != nil {
		closeLogStream()
		return fmt.Errorf("unable to fetch intermediate config: %w", err)
	}

	l.Info("intermediate config fetched",
		"app_branch_id", branch.ID,
		"commit_sha", commitSHA,
		"config_version", intermediateConfig.Version,
		"num_components", len(intermediateConfig.Components))

	// Override branches for all entities (components, sandbox, actions) when
	// their repo matches the branch config's repo.
	branchRepo := ""
	branchName := ""
	if cfg := branch.Configs[0].ConnectedGithubVCSConfig; cfg != nil {
		branchRepo = cfg.Repo
		branchName = cfg.Branch
	} else if cfg := branch.Configs[0].PublicGitVCSConfig; cfg != nil {
		branchRepo = cfg.Repo
		branchName = cfg.Branch
	}
	if branchRepo != "" {
		overrideBranches(intermediateConfig, branchRepo, branchName)
	}

	config.NormalizeIntermediateConfig(intermediateConfig)

	configJSON, err := json.Marshal(intermediateConfig)
	if err != nil {
		closeLogStream()
		return fmt.Errorf("unable to serialize intermediate config: %w", err)
	}
	var sourceConfigJSON []byte
	if intermediateConfig.SourceArchive != nil {
		sourceConfigJSON, err = json.Marshal(intermediateConfig.SourceArchive)
		if err != nil {
			closeLogStream()
			return fmt.Errorf("unable to serialize authored config: %w", err)
		}
	}

	isPreview := run.RunType == app.AppBranchRunTypeGitPreview

	// For preview runs, diff intermediate configs before creating the DB AppConfig.
	// If nothing changed, short-circuit and skip the rest of the workflow.
	var previewDiff *activities.ComputeAppConfigDiffOutput
	var previewBaselineConfigID string
	if isPreview {
		var oldConfigID string
		baseline, baselineErr := activities.AwaitFindLatestNonPreviewAppConfig(ctx, &activities.FindLatestNonPreviewAppConfigInput{
			AppID: branch.AppID,
		})
		if baselineErr == nil && baseline.AppConfigID != "" {
			oldConfigID = baseline.AppConfigID
			previewBaselineConfigID = oldConfigID
		}

		diffResult, diffErr := activities.AwaitDiffIntermediateConfigs(ctx, &activities.DiffIntermediateConfigsInput{
			AppID:                     branch.AppID,
			NewIntermediateConfigJSON: string(configJSON),
			OldAppConfigID:            oldConfigID,
		})
		if diffErr != nil {
			l.Warn("unable to diff intermediate configs, continuing", "error", diffErr)
		} else if !diffResult.Changed && !run.Force {
			_ = activities.AwaitUpdateAppBranchRunNoConfigChanges(ctx, &activities.UpdateAppBranchRunNoConfigChangesInput{
				RunID: s.RunID,
			})

			if run.PRNumber != nil {
				commentBody := activities.BuildPRCommentBody(&activities.PRCommentParams{
					AppName: branch.Name,
					RunID:   s.RunID,
					Status:  activities.PRCommentStatusSkipped,
				})
				_, _ = activities.AwaitCreateOrUpdatePRComment(ctx, &activities.CreateOrUpdatePRCommentInput{
					VcsConfigID:       vcsConfigID,
					PRNumber:          *run.PRNumber,
					ExistingCommentID: run.GithubCommentID,
					Body:              commentBody,
				})
				_ = activities.AwaitSetGithubCommitStatus(ctx, &activities.SetGithubCommitStatusInput{
					VcsConfigID: vcsConfigID,
					CommitSHA:   run.HeadSHA,
					State:       "success",
					Context:     "nuon/preview",
					Description: "No config changes detected",
				})
			}

			if s.StepID != "" {
				_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
					ID: s.StepID,
					Status: app.CompositeStatus{
						Status:                 app.StatusSuccess,
						StatusHumanDescription: "no config changes detected",
					},
				})
			}

			closeLogStream()
			return nil
		}

		if diffResult.Diff != nil {
			previewDiff = diffResult.Diff
		}
	}

	// The baseline lookup for PR comments excludes configs labelled as previews,
	// so the label has to track preview-ness rather than how the run was
	// triggered — otherwise a plan-only run's config becomes the baseline that
	// the next preview diffs against.
	source := string(run.RunType)
	if run.IsPreview() {
		source = string(app.AppBranchRunTypeGitPreview)
	}
	configLabels := labels.Labels{
		"source": source,
	}
	if run.HeadSHA != "" {
		configLabels["commit"] = run.HeadSHA
	}
	if run.PRNumber != nil {
		configLabels["pr"] = strconv.Itoa(*run.PRNumber)
	}
	configLabels.Merge(run.Labels)

	createResp, err := activities.AwaitCreateAppConfig(ctx, activities.CreateAppConfigRequest{
		Req: &activities.CreateAppConfigInput{
			AppID:                  branch.AppID,
			OrgID:                  branch.OrgID,
			AppBranchID:            branch.ID,
			CreatedByID:            branch.CreatedByID,
			IntermediateConfigJSON: string(configJSON),
			SourceConfigJSON:       string(sourceConfigJSON),
			Labels:                 configLabels,
		},
	})
	if err != nil {
		closeLogStream()
		return fmt.Errorf("unable to create app config: %w", err)
	}

	l.Info("app config created",
		"app_config_id", createResp.AppConfigID,
		"app_branch_id", branch.ID)

	return s.syncAndFinalize(ctx, finalizeParams{
		run:                     run,
		branch:                  branch,
		appConfigID:             createResp.AppConfigID,
		vcsConfigID:             vcsConfigID,
		isPreview:               isPreview,
		previewDiff:             previewDiff,
		previewBaselineConfigID: previewBaselineConfigID,
	}, closeLogStream)
}

type finalizeParams struct {
	run                     *app.AppBranchRun
	branch                  *app.AppBranch
	appConfigID             string
	vcsConfigID             string
	isPreview               bool
	previewDiff             *activities.ComputeAppConfigDiffOutput
	previewBaselineConfigID string
}

// syncAndFinalize turns an app config into database records and reports the
// result onto the step. Shared by the VCS path, which has just created the
// config from a cloned repo, and the pre-compiled path, which was handed one.
func (s *Signal) syncAndFinalize(ctx workflow.Context, p finalizeParams, closeLogStream func()) error {
	l := workflow.GetLogger(ctx)
	run, branch := p.run, p.branch

	syncResp, err := activities.AwaitSyncAppConfig(ctx, activities.SyncAppConfigRequest{
		Req: &activities.SyncAppConfigInput{
			AppConfigID: p.appConfigID,
			AppID:       branch.AppID,
			AppBranchID: branch.ID,
		},
	})
	if err != nil {
		closeLogStream()
		return fmt.Errorf("unable to sync app config: %w", err)
	}

	l.Info("app config synced",
		"app_config_id", syncResp.AppConfigID,
		"component_count", len(syncResp.ComponentIDs),
		"action_count", len(syncResp.ActionIDs))

	// Update AppBranchConfig with component and action IDs
	if err := activities.AwaitUpdateAppBranchConfigIDs(ctx, activities.UpdateAppBranchConfigIDsRequest{
		Req: &activities.UpdateAppBranchConfigIDsInput{
			AppBranchConfigID: branch.Configs[0].ID,
			ComponentIDs:      syncResp.ComponentIDs,
			ActionIDs:         syncResp.ActionIDs,
			RunbookIDs:        syncResp.RunbookIDs,
		},
	}); err != nil {
		l.Warn("unable to update app branch config IDs", "error", err)
	}

	if err := activities.AwaitUpdateAppBranchRunAppConfig(ctx, activities.UpdateAppBranchRunAppConfigRequest{
		Req: &activities.UpdateAppBranchRunAppConfigInput{
			RunID:       s.RunID,
			AppConfigID: syncResp.AppConfigID,
		},
	}); err != nil {
		closeLogStream()
		return fmt.Errorf("unable to update run with app config ID: %w", err)
	}

	if s.StepID != "" {
		meta := map[string]any{
			"app_config_id":   syncResp.AppConfigID,
			"component_count": len(syncResp.ComponentIDs),
			"action_count":    len(syncResp.ActionIDs),
		}

		var configDiff *activities.ComputeAppConfigDiffOutput
		if p.previewDiff != nil {
			configDiff = p.previewDiff
		} else if p.isPreview {
			var oldConfigID string
			if run.Comparison != nil && run.Comparison.BaseRun != nil && run.Comparison.BaseRun.AppConfigID != "" {
				oldConfigID = run.Comparison.BaseRun.AppConfigID
			} else if run.Comparison != nil && run.Comparison.BaseRunID != nil && *run.Comparison.BaseRunID != "" {
				prevRun, prevErr := activities.AwaitGetAppBranchRunByIDByRunID(ctx, *run.Comparison.BaseRunID)
				if prevErr == nil && prevRun.AppConfigID != "" {
					oldConfigID = prevRun.AppConfigID
				}
			}

			computed, diffErr := activities.AwaitComputeAppConfigDiff(ctx, &activities.ComputeAppConfigDiffInput{
				AppID:       branch.AppID,
				NewConfigID: syncResp.AppConfigID,
				OldConfigID: oldConfigID,
			})
			if diffErr == nil {
				configDiff = computed
			}
		}

		if p.previewBaselineConfigID != "" {
			meta["baseline_app_config_id"] = p.previewBaselineConfigID
		}

		_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.StepID,
			Status: app.CompositeStatus{
				Status:                 app.StatusSuccess,
				StatusHumanDescription: fmt.Sprintf("synced %d components, %d actions, %d runbooks", len(syncResp.ComponentIDs), len(syncResp.ActionIDs), len(syncResp.RunbookIDs)),
				Metadata:               meta,
			},
		})

		if p.isPreview && run.PRNumber != nil {
			commentBody := activities.BuildPRCommentBody(&activities.PRCommentParams{
				AppName: branch.Name,
				RunID:   s.RunID,
				Status:  activities.PRCommentStatusPending,
				Diff:    configDiff,
			})
			_, _ = activities.AwaitCreateOrUpdatePRComment(ctx, &activities.CreateOrUpdatePRCommentInput{
				VcsConfigID:       p.vcsConfigID,
				PRNumber:          *run.PRNumber,
				ExistingCommentID: run.GithubCommentID,
				Body:              commentBody,
			})
		}
	}

	closeLogStream()
	return nil
}
