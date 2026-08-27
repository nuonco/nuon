package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type ComputeAndStoreAppBranchRunComparisonInput struct {
	AppBranchID string `json:"app_branch_id" validate:"required"`
	RunID       string `json:"run_id" validate:"required"`
}

type ComputeAndStoreAppBranchRunComparisonOutput struct {
	Skipped          bool   `json:"skipped"`
	SkipReason       string `json:"skip_reason,omitempty"`
	BaseRunID        string `json:"base_run_id,omitempty"`
	HeadRunID        string `json:"head_run_id,omitempty"`
	BaseSHA          string `json:"base_sha,omitempty"`
	HeadSHA          string `json:"head_sha,omitempty"`
	FilesChanged     int    `json:"files_changed"`
	Additions        int    `json:"additions"`
	Removals         int    `json:"removals"`
	Changed          int    `json:"changed"`
	GitDiffStored    bool   `json:"git_diff_stored"`
	FullDiffStored   bool   `json:"full_diff_stored"`
	ConfigDiffStored bool   `json:"config_diff_stored"`
}

// ConfigDiffWithSourceOutput is FullDiff enriched with source_changed flags.
type ConfigDiffWithSourceOutput struct {
	ConfigFile string                        `json:"config_file"`
	Additions  int                           `json:"additions"`
	Removals   int                           `json:"removals"`
	Changed    int                           `json:"changed"`
	Sections   []ConfigDiffSectionWithSource `json:"sections"`
}

type ConfigDiffSectionWithSource struct {
	Name      string                      `json:"name"`
	Additions int                         `json:"additions"`
	Removals  int                         `json:"removals"`
	Changed   int                         `json:"changed"`
	Entries   []ConfigDiffEntryWithSource `json:"entries"`
}

type ConfigDiffEntryWithSource struct {
	Op            string `json:"op"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	SourceChanged bool   `json:"source_changed"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 10m
func (a *Activities) ComputeAndStoreAppBranchRunComparison(ctx context.Context, input *ComputeAndStoreAppBranchRunComparisonInput) (*ComputeAndStoreAppBranchRunComparisonOutput, error) {
	if err := a.v.Struct(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	out := &ComputeAndStoreAppBranchRunComparisonOutput{
		HeadRunID: input.RunID,
	}

	var comparison app.AppBranchRunComparison
	res := a.db.WithContext(ctx).
		Where(app.AppBranchRunComparison{HeadRunID: input.RunID}).
		First(&comparison)
	if res.Error != nil {
		out.Skipped = true
		out.SkipReason = "no comparison row"
		return out, nil
	}

	if comparison.BaseRunID == nil || *comparison.BaseRunID == "" {
		out.Skipped = true
		out.SkipReason = "no base run"
		return out, nil
	}
	out.BaseRunID = *comparison.BaseRunID

	var headRun app.AppBranchRun
	if err := a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		First(&headRun, "id = ?", input.RunID).Error; err != nil {
		return nil, fmt.Errorf("unable to load head run: %w", err)
	}

	var baseRun app.AppBranchRun
	if err := a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		First(&baseRun, "id = ?", *comparison.BaseRunID).Error; err != nil {
		return nil, fmt.Errorf("unable to load base run: %w", err)
	}

	headSHA := runCommitSHA(&headRun)
	baseSHA := runCommitSHA(&baseRun)
	out.HeadSHA = headSHA
	out.BaseSHA = baseSHA

	branch, err := a.getAppBranchByID(ctx, input.AppBranchID)
	if err != nil {
		return nil, fmt.Errorf("unable to load app branch: %w", err)
	}

	var changedPaths []string
	if headSHA != "" && baseSHA != "" {
		vcsConfigID := branchVCSConfigID(branch)
		if vcsConfigID == "" {
			a.l.Warn("no VCS config for git diff",
				zap.String("app_branch_id", input.AppBranchID),
				zap.String("run_id", input.RunID))
		} else {
			workspaceID := fmt.Sprintf("app-branch-run-comparison-%s", comparison.ID)
			gitDiff, gitErr := a.computeGitDiffBetweenSHAs(ctx, vcsConfigID, baseSHA, headSHA, workspaceID)
			if gitErr != nil {
				a.l.Warn("git diff failed; continuing with config diffs",
					zap.String("run_id", input.RunID),
					zap.Error(gitErr))
			} else if gitDiff != nil {
				changedPaths = gitDiff.ChangedPaths
				out.FilesChanged = gitDiff.FilesChanged
				if err := a.uploadComparisonBlob(ctx, comparison.ID, "git_diff", gitDiff, &comparison.GitDiff); err != nil {
					a.l.Warn("unable to store git diff blob", zap.Error(err))
				} else {
					out.GitDiffStored = true
				}
			}
		}
	} else {
		a.l.Warn("missing commit SHAs for git diff",
			zap.String("run_id", input.RunID),
			zap.String("base_sha", baseSHA),
			zap.String("head_sha", headSHA))
	}

	if headRun.AppConfigID == "" || baseRun.AppConfigID == "" {
		if err := a.persistComparisonBlobs(ctx, &comparison); err != nil {
			return nil, err
		}
		out.Skipped = !out.GitDiffStored
		if out.Skipped {
			out.SkipReason = "missing app config ids"
		}
		return out, nil
	}

	fullDiff, err := a.ComputeAppConfigDiff(ctx, &ComputeAppConfigDiffInput{
		AppID:       branch.AppID,
		NewConfigID: headRun.AppConfigID,
		OldConfigID: baseRun.AppConfigID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to compute full config diff: %w", err)
	}

	out.Additions = fullDiff.Additions
	out.Removals = fullDiff.Removals
	out.Changed = fullDiff.Changed

	if err := a.uploadComparisonBlob(ctx, comparison.ID, "full_diff", fullDiff, &comparison.FullDiff); err != nil {
		return nil, fmt.Errorf("unable to store full diff blob: %w", err)
	}
	out.FullDiffStored = true

	componentDirs, dirErr := a.loadComponentDirectories(ctx, headRun.AppConfigID)
	if dirErr != nil {
		a.l.Warn("unable to load component directories for source_changed", zap.Error(dirErr))
		componentDirs = map[string]string{}
	}

	configDiff := enrichConfigDiffWithSourceChanged(fullDiff, componentDirs, changedPaths)
	if err := a.uploadComparisonBlob(ctx, comparison.ID, "config_diff", configDiff, &comparison.ConfigDiff); err != nil {
		return nil, fmt.Errorf("unable to store config diff blob: %w", err)
	}
	out.ConfigDiffStored = true

	if err := a.persistComparisonBlobs(ctx, &comparison); err != nil {
		return nil, err
	}

	return out, nil
}

func runCommitSHA(run *app.AppBranchRun) string {
	if run == nil {
		return ""
	}
	if run.VCSConnectionCommit != nil && run.VCSConnectionCommit.SHA != "" {
		return run.VCSConnectionCommit.SHA
	}
	return run.HeadSHA
}

func branchVCSConfigID(branch *app.AppBranch) string {
	if branch == nil || len(branch.Configs) == 0 {
		return ""
	}
	cfg := branch.Configs[0]
	if cfg.ConnectedGithubVCSConfig != nil {
		return cfg.ConnectedGithubVCSConfig.ID
	}
	if cfg.PublicGitVCSConfig != nil {
		return cfg.PublicGitVCSConfig.ID
	}
	return ""
}

func (a *Activities) uploadComparisonBlob(ctx context.Context, comparisonID, kind string, payload any, dest **blobstore.Blob) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", kind, err)
	}

	blobID := domains.NewBlobID()
	s3Key := fmt.Sprintf("blobs/app_branch_run_comparisons/%s/%s/%s", kind, comparisonID, blobID)

	checksum, err := a.blobSvc.UploadStream(ctx, s3Key, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("upload %s: %w", kind, err)
	}

	metadata := blobstore.BlobMetadata{
		BlobID:      blobID,
		S3Key:       s3Key,
		Size:        int64(len(body)),
		ContentType: "application/json",
		Checksum:    checksum,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal %s metadata: %w", kind, err)
	}

	blob := &blobstore.Blob{}
	if err := blob.Scan(metadataJSON); err != nil {
		return fmt.Errorf("scan %s metadata: %w", kind, err)
	}
	*dest = blob
	return nil
}

func (a *Activities) persistComparisonBlobs(ctx context.Context, comparison *app.AppBranchRunComparison) error {
	updates := map[string]any{}
	if comparison.GitDiff != nil {
		v, err := comparison.GitDiff.Value()
		if err != nil {
			return fmt.Errorf("git_diff value: %w", err)
		}
		updates["git_diff"] = v
	}
	if comparison.FullDiff != nil {
		v, err := comparison.FullDiff.Value()
		if err != nil {
			return fmt.Errorf("full_diff value: %w", err)
		}
		updates["full_diff"] = v
	}
	if comparison.ConfigDiff != nil {
		v, err := comparison.ConfigDiff.Value()
		if err != nil {
			return fmt.Errorf("config_diff value: %w", err)
		}
		updates["config_diff"] = v
	}
	if len(updates) == 0 {
		return nil
	}

	res := a.db.WithContext(ctx).
		Model(&app.AppBranchRunComparison{}).
		Where(app.AppBranchRunComparison{ID: comparison.ID}).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("unable to update comparison blobs: %w", res.Error)
	}
	return nil
}

func (a *Activities) loadComponentDirectories(ctx context.Context, appConfigID string) (map[string]string, error) {
	var conns []app.ComponentConfigConnection
	err := a.db.WithContext(ctx).
		Preload("Component").
		Preload("TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("HelmComponentConfig.PublicGitVCSConfig").
		Preload("DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("KubernetesManifestComponentConfig.ConnectedGithubVCSConfig").
		Preload("KubernetesManifestComponentConfig.PublicGitVCSConfig").
		Preload("PulumiComponentConfig.ConnectedGithubVCSConfig").
		Preload("PulumiComponentConfig.PublicGitVCSConfig").
		Where(app.ComponentConfigConnection{AppConfigID: appConfigID}).
		Find(&conns).Error
	if err != nil {
		return nil, err
	}

	out := make(map[string]string, len(conns))
	for i := range conns {
		c := &conns[i]
		name := c.ComponentName
		if name == "" && c.Component.Name != "" {
			name = c.Component.Name
		}
		if name == "" {
			continue
		}
		dir := ""
		if c.ConnectedGithubVCSConfig != nil {
			dir = c.ConnectedGithubVCSConfig.Directory
		} else if c.PublicGitVCSConfig != nil {
			dir = c.PublicGitVCSConfig.Directory
		}
		if c.KubernetesManifestComponentConfig != nil &&
			c.KubernetesManifestComponentConfig.Kustomize != nil &&
			c.KubernetesManifestComponentConfig.Kustomize.Path != "" {
			kustomizePath := c.KubernetesManifestComponentConfig.Kustomize.Path
			if dir == "." || dir == "" {
				dir = kustomizePath
			} else {
				dir = strings.TrimSuffix(dir, "/") + "/" + strings.TrimPrefix(kustomizePath, "./")
			}
		}
		// Omit missing / repo-root "." so inputs-only git changes do not
		// mark every component source_changed (enrich skips missing names).
		if normalizeRepoPath(dir) == "" {
			continue
		}
		out[name] = dir
	}
	return out, nil
}
