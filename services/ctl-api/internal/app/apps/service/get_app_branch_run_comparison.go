package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type AppBranchRunComparisonResponse struct {
	ID        string `json:"id"`
	HeadRunID string `json:"head_run_id"`
	BaseRunID string `json:"base_run_id,omitempty"`
	BaseSHA   string `json:"base_sha,omitempty"`
	HeadSHA   string `json:"head_sha,omitempty"`

	GitDiff    *blobstore.BlobMetadata `json:"git_diff,omitempty"`
	FullDiff   *blobstore.BlobMetadata `json:"full_diff,omitempty"`
	ConfigDiff *blobstore.BlobMetadata `json:"config_diff,omitempty"`

	GitDiffContent    any `json:"git_diff_content,omitempty"`
	FullDiffContent   any `json:"full_diff_content,omitempty"`
	ConfigDiffContent any `json:"config_diff_content,omitempty"`
}

// @ID						GetAppBranchRunComparison
// @Summary				get comparison for an app branch run
// @Description			Returns the AppBranchRunComparison for a run (as head), optionally including diff blob contents via include_diff=git|full|config
// @Tags					apps
// @Param					app_id			path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					run_id			path	string	true	"app branch run ID"
// @Param					include_diff	query	string	false	"comma-separated: git,full,config"
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	AppBranchRunComparisonResponse
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/runs/{run_id}/comparison [get]
func (s *service) GetAppBranchRunComparison(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureAppBranches)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check feature: %w", err))
		return
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureAppBranches))
		return
	}

	appID := ctx.Param("app_id")
	appBranchID := ctx.Param("app_branch_id")
	runID := ctx.Param("run_id")

	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Where(app.AppBranch{OrgID: org.ID, AppID: appID}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
		return
	}

	var run app.AppBranchRun
	res = s.db.WithContext(ctx).
		Where(app.AppBranchRun{AppBranchID: appBranchID}).
		First(&run, "id = ?", runID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch run: %w", res.Error))
		return
	}

	var comparison app.AppBranchRunComparison
	res = s.db.WithContext(ctx).
		Preload("HeadRun.VCSConnectionCommit").
		Preload("BaseRun.VCSConnectionCommit").
		Where(app.AppBranchRunComparison{HeadRunID: runID}).
		First(&comparison)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find run comparison: %w", res.Error))
		return
	}

	resp := AppBranchRunComparisonResponse{
		ID:         comparison.ID,
		HeadRunID:  comparison.HeadRunID,
		GitDiff:    blobMetadataPtr(comparison.GitDiff),
		FullDiff:   blobMetadataPtr(comparison.FullDiff),
		ConfigDiff: blobMetadataPtr(comparison.ConfigDiff),
	}
	if comparison.BaseRunID != nil {
		resp.BaseRunID = *comparison.BaseRunID
	}
	resp.HeadSHA = comparisonRunSHA(&comparison.HeadRun)
	if comparison.BaseRun != nil {
		resp.BaseSHA = comparisonRunSHA(comparison.BaseRun)
	}

	include := parseIncludeDiff(ctx.Query("include_diff"))
	if len(include) > 0 {
		blobCtx := blobstore.WithBlobService(ctx.Request.Context(), s.blobSvc)
		if include["git"] && comparison.GitDiff != nil {
			content, err := loadBlobJSONContent(blobCtx, comparison.GitDiff)
			if err != nil {
				ctx.Error(fmt.Errorf("unable to load git diff: %w", err))
				return
			}
			resp.GitDiffContent = content
		}
		if include["full"] && comparison.FullDiff != nil {
			content, err := loadBlobJSONContent(blobCtx, comparison.FullDiff)
			if err != nil {
				ctx.Error(fmt.Errorf("unable to load full diff: %w", err))
				return
			}
			resp.FullDiffContent = content
		}
		if include["config"] && comparison.ConfigDiff != nil {
			content, err := loadBlobJSONContent(blobCtx, comparison.ConfigDiff)
			if err != nil {
				ctx.Error(fmt.Errorf("unable to load config diff: %w", err))
				return
			}
			resp.ConfigDiffContent = content
		}
	}

	ctx.JSON(http.StatusOK, resp)
}

func blobMetadataPtr(b *blobstore.Blob) *blobstore.BlobMetadata {
	if b == nil || !b.IsSet() {
		return nil
	}
	m := b.Metadata()
	return &m
}

func comparisonRunSHA(run *app.AppBranchRun) string {
	if run == nil {
		return ""
	}
	if run.VCSConnectionCommit != nil && run.VCSConnectionCommit.SHA != "" {
		return run.VCSConnectionCommit.SHA
	}
	return run.HeadSHA
}

func parseIncludeDiff(raw string) map[string]bool {
	out := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if part == "" {
			continue
		}
		out[part] = true
	}
	return out
}

func loadBlobJSONContent(ctx context.Context, blob *blobstore.Blob) (any, error) {
	raw, err := blob.Get(ctx)
	if err != nil {
		return nil, err
	}
	if raw == "" {
		return nil, nil
	}
	var content any
	if err := json.Unmarshal([]byte(raw), &content); err != nil {
		return raw, nil
	}
	return content, nil
}
