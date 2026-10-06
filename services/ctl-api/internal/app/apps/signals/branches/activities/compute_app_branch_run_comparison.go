package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
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
	SourceDiffStored bool   `json:"source_diff_stored"`
}

// ConfigDiffWithSourceOutput is FullDiff enriched with source_changed flags.
type ConfigDiffWithSourceOutput struct {
	ConfigFile             string                        `json:"config_file"`
	Additions              int                           `json:"additions"`
	Removals               int                           `json:"removals"`
	Changed                int                           `json:"changed"`
	ComponentSourceChanged map[string]bool               `json:"component_source_changed,omitempty"`
	Sections               []ConfigDiffSectionWithSource `json:"sections"`
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
	// File is the repo-relative path of the config file this entity was
	// parsed from, resolved from the head config's source archive members.
	File string `json:"file,omitempty"`
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

	var headRun app.AppBranchRun
	if err := a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Preload("Preview").
		First(&headRun, "id = ?", input.RunID).Error; err != nil {
		return nil, fmt.Errorf("unable to load head run: %w", err)
	}

	if comparison.BaseRunID == nil || *comparison.BaseRunID == "" {
		baseRun, findErr := a.helpers.FindBaseAppBranchRunForHead(ctx, &headRun)
		if findErr == nil {
			comparison.BaseRunID = &baseRun.ID
			if err := a.db.WithContext(ctx).
				Model(&comparison).
				Select("base_run_id").
				Updates(&comparison).Error; err != nil {
				return nil, fmt.Errorf("unable to persist comparison base run: %w", err)
			}
		} else if findErr != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("unable to resolve comparison base run: %w", findErr)
		}
	}

	// A first run on a branch has no baseline: diff against an empty baseline
	// so every entity and file shows as added rather than an empty diff.
	noBase := comparison.BaseRunID == nil || *comparison.BaseRunID == ""

	var baseRun app.AppBranchRun
	if !noBase {
		out.BaseRunID = *comparison.BaseRunID
		if err := a.db.WithContext(ctx).
			Preload("VCSConnectionCommit").
			First(&baseRun, "id = ?", *comparison.BaseRunID).Error; err != nil {
			return nil, fmt.Errorf("unable to load base run: %w", err)
		}
	}

	headSHA := runCommitSHA(&headRun)
	baseSHA := ""
	if !noBase {
		baseSHA = runCommitSHA(&baseRun)
	}
	out.HeadSHA = headSHA
	out.BaseSHA = baseSHA

	branch, err := a.getAppBranchByID(ctx, input.AppBranchID)
	if err != nil {
		return nil, fmt.Errorf("unable to load app branch: %w", err)
	}

	var changedPaths []string
	switch {
	case noBase:
		a.l.Info("no base run for comparison; diffing against empty baseline",
			zap.String("run_id", input.RunID),
			zap.String("head_sha", headSHA))
	case headSHA != "" && baseSHA != "":
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
	default:
		a.l.Warn("missing commit SHAs for git diff",
			zap.String("run_id", input.RunID),
			zap.String("base_sha", baseSHA),
			zap.String("head_sha", headSHA))
	}

	if headRun.AppConfigID == "" || (!noBase && baseRun.AppConfigID == "") {
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

	componentSources, dirErr := a.loadComponentSources(ctx, headRun.AppConfigID)
	if dirErr != nil {
		a.l.Warn("unable to load component sources for source_changed", zap.Error(dirErr))
		componentSources = nil
	}

	configDiff := a.computeAndEnrichConfigDiff(ctx, branch, fullDiff, componentSources, changedPaths, headRun.AppConfigID, baseRun.AppConfigID)
	if configDiff != nil {
		baseSHAs := map[string]string{}
		if !noBase && baseRun.AppConfigID != "" {
			loaded, shaErr := a.loadComponentBuildSHAs(ctx, baseRun.AppConfigID)
			if shaErr != nil {
				a.l.Warn("unable to load base component build commits", zap.Error(shaErr))
			} else {
				baseSHAs = loaded
			}
		}
		applySourceCommitComparison(configDiff, baseSHAs, currentComponentSHAs(componentSources, branch, headSHA))
	}
	if err := a.uploadComparisonBlob(ctx, comparison.ID, "config_diff", configDiff, &comparison.ConfigDiff); err != nil {
		return nil, fmt.Errorf("unable to store config diff blob: %w", err)
	}
	out.ConfigDiffStored = true

	sourceDiff, sourceDiffErr := a.computeSourceArchiveDiff(ctx, branch.AppID, headRun.AppConfigID, baseRun.AppConfigID)
	if sourceDiffErr != nil {
		a.l.Warn("unable to compute source archive diff",
			zap.String("run_id", input.RunID),
			zap.Error(sourceDiffErr))
	} else if sourceDiff != nil {
		if err := a.uploadComparisonBlob(ctx, comparison.ID, "source_diff", sourceDiff, &comparison.SourceDiff); err != nil {
			a.l.Warn("unable to store source diff blob", zap.Error(err))
		} else {
			out.SourceDiffStored = true
		}
	}

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

func branchRepo(branch *app.AppBranch) string {
	repo, _ := branchRepoAndBranch(branch)
	return repo
}

func branchRepoAndBranch(branch *app.AppBranch) (string, string) {
	if branch == nil || len(branch.Configs) == 0 {
		return "", ""
	}
	cfg := branch.Configs[0]
	if cfg.ConnectedGithubVCSConfig != nil {
		return cfg.ConnectedGithubVCSConfig.Repo, cfg.ConnectedGithubVCSConfig.Branch
	}
	if cfg.PublicGitVCSConfig != nil {
		return cfg.PublicGitVCSConfig.Repo, cfg.PublicGitVCSConfig.Branch
	}
	return "", ""
}

func currentComponentSHAs(sources []componentSource, branch *app.AppBranch, headSHA string) map[string]string {
	out := map[string]string{}
	if headSHA == "" {
		return out
	}
	repo, branchName := branchRepoAndBranch(branch)
	for _, src := range sources {
		if componentTracksRun(src, repo, branchName) {
			out[src.Name] = headSHA
		}
	}
	return out
}

func (a *Activities) loadComponentBuildSHAs(ctx context.Context, appConfigID string) (map[string]string, error) {
	out := map[string]string{}
	if appConfigID == "" {
		return out, nil
	}

	var conns []app.ComponentConfigConnection
	err := a.db.WithContext(ctx).
		Preload("Component").
		Where(app.ComponentConfigConnection{AppConfigID: appConfigID}).
		Find(&conns).Error
	if err != nil {
		return nil, err
	}

	nameByBuild := map[string]string{}
	buildIDs := make([]string, 0, len(conns))
	for i := range conns {
		c := &conns[i]
		if !c.LatestBuildID.Valid || c.LatestBuildID.String == "" {
			continue
		}
		name := c.ComponentName
		if name == "" {
			name = c.Component.Name
		}
		if name == "" {
			continue
		}
		nameByBuild[c.LatestBuildID.String] = name
		buildIDs = append(buildIDs, c.LatestBuildID.String)
	}
	if len(buildIDs) == 0 {
		return out, nil
	}

	var builds []app.ComponentBuild
	err = a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Where("id IN ?", buildIDs).
		Find(&builds).Error
	if err != nil {
		return nil, err
	}
	for i := range builds {
		b := &builds[i]
		if b.VCSConnectionCommit == nil || b.VCSConnectionCommit.SHA == "" {
			continue
		}
		if name := nameByBuild[b.ID]; name != "" {
			out[name] = b.VCSConnectionCommit.SHA
		}
	}
	return out, nil
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
	if comparison.SourceDiff != nil {
		v, err := comparison.SourceDiff.Value()
		if err != nil {
			return fmt.Errorf("source_diff value: %w", err)
		}
		updates["source_diff"] = v
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

// componentRepoDirectory reads the repo from the preloaded typed config.
// AfterQuery copies those pointers onto the connection before preloads run,
// so the connection-level fields are still empty here.
func componentRepoDirectory(c *app.ComponentConfigConnection) (string, string, string) {
	switch {
	case c.HelmComponentConfig != nil:
		return vcsRepoDirectory(c.HelmComponentConfig.ConnectedGithubVCSConfig, c.HelmComponentConfig.PublicGitVCSConfig)
	case c.TerraformModuleComponentConfig != nil:
		return vcsRepoDirectory(c.TerraformModuleComponentConfig.ConnectedGithubVCSConfig, c.TerraformModuleComponentConfig.PublicGitVCSConfig)
	case c.DockerBuildComponentConfig != nil:
		return vcsRepoDirectory(c.DockerBuildComponentConfig.ConnectedGithubVCSConfig, c.DockerBuildComponentConfig.PublicGitVCSConfig)
	case c.KubernetesManifestComponentConfig != nil:
		return vcsRepoDirectory(c.KubernetesManifestComponentConfig.ConnectedGithubVCSConfig, c.KubernetesManifestComponentConfig.PublicGitVCSConfig)
	case c.PulumiComponentConfig != nil:
		return vcsRepoDirectory(c.PulumiComponentConfig.ConnectedGithubVCSConfig, c.PulumiComponentConfig.PublicGitVCSConfig)
	default:
		return vcsRepoDirectory(c.ConnectedGithubVCSConfig, c.PublicGitVCSConfig)
	}
}

func vcsRepoDirectory(github *app.ConnectedGithubVCSConfig, public *app.PublicGitVCSConfig) (string, string, string) {
	if github != nil && github.Repo != "" {
		return github.Repo, github.Directory, github.Branch
	}
	if public != nil && public.Repo != "" {
		return public.Repo, public.Directory, public.Branch
	}
	return "", "", ""
}

func (a *Activities) loadComponentSources(ctx context.Context, appConfigID string) ([]componentSource, error) {
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

	out := make([]componentSource, 0, len(conns))
	for i := range conns {
		c := &conns[i]
		name := c.ComponentName
		if name == "" && c.Component.Name != "" {
			name = c.Component.Name
		}
		if name == "" {
			continue
		}
		src := componentSource{Name: name}
		src.Repo, src.Directory, src.Branch = componentRepoDirectory(c)
		if c.KubernetesManifestComponentConfig != nil &&
			c.KubernetesManifestComponentConfig.Kustomize != nil &&
			c.KubernetesManifestComponentConfig.Kustomize.Path != "" {
			kustomizePath := c.KubernetesManifestComponentConfig.Kustomize.Path
			if src.Directory == "." || src.Directory == "" {
				src.Directory = kustomizePath
			} else {
				src.Directory = strings.TrimSuffix(src.Directory, "/") + "/" + strings.TrimPrefix(kustomizePath, "./")
			}
		}
		out = append(out, src)
	}
	return out, nil
}

// computeAndEnrichConfigDiff enriches the config diff with source_changed
// flags and per-entity file refs. Member refs resolve from the head archive
// first (added/changed entities), falling back to the base archive (removed
// entities exist only there).
func (a *Activities) computeAndEnrichConfigDiff(
	ctx context.Context,
	branch *app.AppBranch,
	fullDiff *ComputeAppConfigDiffOutput,
	componentSources []componentSource,
	changedPaths []string,
	headConfigID, baseConfigID string,
) *ConfigDiffWithSourceOutput {
	headArchive, headErr := a.loadSourceArchive(ctx, branch.AppID, headConfigID)
	if headErr != nil {
		a.l.Warn("unable to load head source archive for member refs", zap.Error(headErr))
		headArchive = nil
	}
	var baseArchive *pkgconfig.SourceArchive
	if baseConfigID != "" {
		var baseErr error
		baseArchive, baseErr = a.loadSourceArchive(ctx, branch.AppID, baseConfigID)
		if baseErr != nil {
			a.l.Warn("unable to load base source archive for member refs", zap.Error(baseErr))
			baseArchive = nil
		}
	}

	return enrichConfigDiffWithSourceChanged(fullDiff, componentSources, branchRepo(branch), changedPaths, headArchive, baseArchive)
}
