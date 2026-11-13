package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	githubrepo "github.com/nuonco/nuon/pkg/github/repo"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildsource"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type InstallDeploymentSourceDiff struct {
	CompareURL string `json:"compare_url,omitempty"`
}

// @ID          GetInstallDeploymentSourceDiff
// @Summary     get the GitHub compare link for a deployment resource
// @Description Returns a GitHub compare URL between the commits of the resource's build for the install's pinned app config and its build for this deployment's app config. The URL is empty when the builds are not in the same GitHub repository, use the same commit, or have the same source digest. Nuon does not read the repository for this response.
// @Param       install_id   path   string true  "install ID"
// @Param       workflow_id  path   string true  "workflow ID"
// @Param       kind         query  string true  "resource kind" Enums(component,sandbox)
// @Param       name         query  string true  "resource name"
// @Tags        installs
// @Accept      json
// @Produce     json
// @Security    APIKey
// @Security    OrgID
// @Failure     400 {object} stderr.ErrResponse
// @Failure     401 {object} stderr.ErrResponse
// @Failure     403 {object} stderr.ErrResponse
// @Failure     404 {object} stderr.ErrResponse
// @Failure     500 {object} stderr.ErrResponse
// @Success     200 {object} InstallDeploymentSourceDiff
// @Router      /v1/installs/{install_id}/deployments/{workflow_id}/source-diff [GET]
func (s *service) GetInstallDeploymentSourceDiff(ctx *gin.Context) {
	ctx.Header("Cache-Control", "private, no-store")
	var query struct {
		Kind string `form:"kind" binding:"required,oneof=component sandbox"`
		Name string `form:"name" binding:"required"`
	}
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid source diff query: %w", err),
			Description: "kind and resource name are required",
		})
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var install app.Install
	if err := s.db.WithContext(ctx).Where(app.Install{ID: ctx.Param("install_id"), OrgID: org.ID}).First(&install).Error; err != nil {
		ctx.Error(err)
		return
	}
	var workflow app.Workflow
	if err := s.db.WithContext(ctx).
		Where(app.Workflow{ID: ctx.Param("workflow_id"), OwnerID: install.ID, OrgID: org.ID}).
		Where("plan_only = ?", false).First(&workflow).Error; err != nil {
		ctx.Error(err)
		return
	}
	var version app.InstallAppConfigVersion
	err = s.db.WithContext(ctx).
		Where(app.InstallAppConfigVersion{OrgID: org.ID, InstallID: install.ID, WorkflowID: &workflow.ID}).
		Order("created_at DESC").Order("id DESC").First(&version).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		ctx.Error(err)
		return
	}
	diff, err := loadInstallConfigDiff(ctx, &version)
	if err != nil {
		ctx.Error(err)
		return
	}
	var result InstallDeploymentSourceDiff
	if diff == nil {
		ctx.JSON(http.StatusOK, result)
		return
	}
	var oldID, newID, componentID string
	if query.Kind == "sandbox" {
		if query.Name == "sandbox" && (diff.SandboxChanged || diff.SandboxBuildChanged) {
			oldID, newID = diff.SandboxBuildOldID, diff.SandboxBuildNewID
		}
	} else {
		for _, entry := range diff.Changed {
			if entry.ComponentName == query.Name {
				oldID, newID, componentID = entry.OldBuildID, entry.NewBuildID, entry.ComponentID
				break
			}
		}
	}
	if oldID == "" || newID == "" {
		ctx.JSON(http.StatusOK, result)
		return
	}
	before, err := s.deploymentSourceIdentity(ctx, org.ID, install.AppID, query.Kind, componentID, oldID)
	if err != nil {
		ctx.Error(err)
		return
	}
	after, err := s.deploymentSourceIdentity(ctx, org.ID, install.AppID, query.Kind, componentID, newID)
	if err != nil {
		ctx.Error(err)
		return
	}
	result.CompareURL = githubCompareURL(before, after)
	ctx.JSON(http.StatusOK, result)
}

func (s *service) deploymentSourceIdentity(ctx context.Context, orgID, appID, kind, componentID, buildID string) (*sourceidentity.Identity, error) {
	if kind == "sandbox" {
		var build app.AppSandboxBuild
		err := s.db.WithContext(ctx).Where(app.AppSandboxBuild{ID: buildID, OrgID: orgID, AppID: appID}).First(&build).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return build.SourceIdentity, err
	}
	var component app.Component
	err := s.db.WithContext(ctx).Where(app.Component{ID: componentID, OrgID: orgID, AppID: appID}).First(&component).Error
	if err == nil {
		var build app.ComponentBuild
		err = s.db.WithContext(ctx).Where(app.ComponentBuild{ID: buildID, OrgID: orgID, StoredComponentID: componentID}).First(&build).Error
		if err == nil {
			return build.SourceIdentity, nil
		}
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return nil, err
}

var githubRepoPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9._-]+$`)

func githubRepo(identity *sourceidentity.Identity) string {
	if identity == nil || identity.Repository == "" || !sourceidentity.IsCommit(identity.Commit) {
		return ""
	}
	repository := strings.TrimSuffix(strings.TrimSuffix(identity.Repository, "/"), ".git")
	repository = strings.TrimPrefix(repository, "https://github.com/")
	owner, repo, err := githubrepo.ParseRepo(repository)
	if err != nil || repo == "." || repo == ".." || !githubRepoPattern.MatchString(owner+"/"+repo) {
		return ""
	}
	return owner + "/" + repo
}

func githubCompareURL(before, after *sourceidentity.Identity) string {
	repo := githubRepo(after)
	if repo == "" || githubRepo(before) == "" || !buildsource.SameRepo(before.Repository, after.Repository) || before.Commit == after.Commit {
		return ""
	}
	if before.Digest != "" && before.Digest == after.Digest {
		return ""
	}
	return "https://github.com/" + repo + "/compare/" + before.Commit + ".." + after.Commit
}
