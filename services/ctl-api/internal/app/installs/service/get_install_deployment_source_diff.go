package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type InstallDeploymentSourceBuild struct {
	ID         string `json:"id"`
	Repository string `json:"repository,omitempty"`
	Commit     string `json:"commit,omitempty"`
	Root       string `json:"root,omitempty"`
	Digest     string `json:"digest,omitempty"`
	Coverage   string `json:"coverage,omitempty"`
}

type InstallDeploymentSourceFile struct {
	Path             string `json:"path"`
	Change           string `json:"change" enums:"added,modified,removed"`
	BeforeExecutable bool   `json:"before_executable"`
	AfterExecutable  bool   `json:"after_executable"`
}

type InstallDeploymentSourceFileContent struct {
	Path              string  `json:"path"`
	Before            *string `json:"before,omitempty"`
	After             *string `json:"after,omitempty"`
	UnavailableReason string  `json:"unavailable_reason,omitempty"`
}

type InstallDeploymentSourceDiff struct {
	Kind         string                              `json:"kind" enums:"component,sandbox"`
	Name         string                              `json:"name"`
	Status       string                              `json:"status" enums:"available,unavailable,unsupported"`
	Message      string                              `json:"message,omitempty"`
	Before       *InstallDeploymentSourceBuild       `json:"before,omitempty"`
	After        *InstallDeploymentSourceBuild       `json:"after,omitempty"`
	Partial      bool                                `json:"partial"`
	Reasons      []string                            `json:"reasons,omitempty"`
	Files        []InstallDeploymentSourceFile       `json:"files"`
	SelectedFile *InstallDeploymentSourceFileContent `json:"selected_file,omitempty"`
}

type deploymentSourceSnapshot struct {
	id          string
	identity    *sourceidentity.Identity
	componentID string
	configID    string
}

// @ID          GetInstallDeploymentSourceDiff
// @Summary     compare captured source for a deployment resource
// @Description Compare the resource's frozen previously applied and selected builds. File content is fetched only when path is provided.
// @Param       install_id   path   string true  "install ID"
// @Param       workflow_id  path   string true  "workflow ID"
// @Param       kind         query  string true  "resource kind" Enums(component,sandbox)
// @Param       name         query  string true  "resource name"
// @Param       path         query  string false "changed source file path"
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
	var query struct {
		Kind string `form:"kind" binding:"required,oneof=component sandbox"`
		Name string `form:"name" binding:"required"`
		Path string `form:"path"`
	}
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "kind and resource name are required"})
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
	result := InstallDeploymentSourceDiff{Kind: query.Kind, Name: query.Name, Status: "unavailable", Files: []InstallDeploymentSourceFile{}}
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
	if diff == nil {
		result.Message = "This deployment did not record a source comparison baseline."
		ctx.JSON(http.StatusOK, result)
		return
	}
	var oldID, newID, componentID string
	var added, removed, found bool
	if query.Kind == "sandbox" {
		found = query.Name == "sandbox" && (diff.SandboxChanged || diff.SandboxBuildChanged)
		oldID, newID = diff.SandboxBuildOldID, diff.SandboxBuildNewID
		added = diff.SandboxOldID == "" && diff.SandboxNewID != ""
	} else {
		for groupIndex, group := range [][]app.ComponentDiffEntry{diff.Added, diff.Removed, diff.Changed, diff.Unchanged} {
			for _, entry := range group {
				if entry.ComponentName == query.Name {
					oldID, newID, componentID = entry.OldBuildID, entry.NewBuildID, entry.ComponentID
					added, removed, found = groupIndex == 0, groupIndex == 1, true
					break
				}
			}
			if found {
				break
			}
		}
	}
	if !found {
		ctx.Status(http.StatusNotFound)
		return
	}
	before, err := s.deploymentSourceSnapshot(ctx, org.ID, install.AppID, query.Kind, componentID, oldID)
	if err != nil {
		ctx.Error(err)
		return
	}
	after, err := s.deploymentSourceSnapshot(ctx, org.ID, install.AppID, query.Kind, componentID, newID)
	if err != nil {
		ctx.Error(err)
		return
	}
	result.Before, result.After = sourceBuildMetadata(before), sourceBuildMetadata(after)
	if (before == nil && !added) || (after == nil && !removed) {
		result.Message = "This deployment did not record both previously applied and selected build identities."
	} else {
		populateSourceDiff(&result, before, after)
	}
	if query.Path != "" && result.Status == "available" {
		if !slices.ContainsFunc(result.Files, func(file InstallDeploymentSourceFile) bool { return file.Path == query.Path }) {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "path must identify a changed source file"})
			return
		}
		selected := &InstallDeploymentSourceFileContent{Path: query.Path}
		readCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
		defer cancel()
		selected.Before, err = s.deploymentSourceFile(readCtx, org.ID, before, query.Path)
		if err == nil {
			selected.After, err = s.deploymentSourceFile(readCtx, org.ID, after, query.Path)
		}
		if err != nil {
			selected.Before, selected.After = nil, nil
			selected.UnavailableReason = err.Error()
		}
		result.SelectedFile = selected
	}
	ctx.Header("Cache-Control", "private, no-store")
	ctx.JSON(http.StatusOK, result)
}

func (s *service) deploymentSourceSnapshot(ctx context.Context, orgID, appID, kind, componentID, buildID string) (*deploymentSourceSnapshot, error) {
	if buildID == "" {
		return nil, nil
	}
	snapshot := &deploymentSourceSnapshot{id: buildID, componentID: componentID}
	if kind == "sandbox" {
		var build app.AppSandboxBuild
		if err := s.db.WithContext(ctx).Where(app.AppSandboxBuild{ID: buildID, OrgID: orgID, AppID: appID}).First(&build).Error; err != nil {
			return nil, err
		}
		snapshot.identity, snapshot.configID = build.SourceIdentity, build.AppSandboxConfigID
	} else {
		var component app.Component
		if err := s.db.WithContext(ctx).Where(app.Component{ID: componentID, OrgID: orgID, AppID: appID}).First(&component).Error; err != nil {
			return nil, err
		}
		var build app.ComponentBuild
		if err := s.db.WithContext(ctx).Where(app.ComponentBuild{ID: buildID, OrgID: orgID, StoredComponentID: componentID}).First(&build).Error; err != nil {
			return nil, err
		}
		snapshot.identity, snapshot.configID = build.SourceIdentity, build.ComponentConfigConnectionID
	}
	return snapshot, nil
}

func sourceBuildMetadata(snapshot *deploymentSourceSnapshot) *InstallDeploymentSourceBuild {
	if snapshot == nil {
		return nil
	}
	build := &InstallDeploymentSourceBuild{ID: snapshot.id}
	if identity := snapshot.identity; identity != nil {
		build.Repository, build.Commit, build.Root = identity.Repository, identity.Commit, identity.Root
		build.Digest, build.Coverage = identity.Digest, identity.Coverage
	}
	return build
}

func populateSourceDiff(result *InstallDeploymentSourceDiff, before, after *deploymentSourceSnapshot) {
	files := [2]map[string]sourceidentity.File{{}, {}}
	for index, snapshot := range []*deploymentSourceSnapshot{before, after} {
		if snapshot == nil {
			continue
		}
		identity := snapshot.identity
		if identity == nil || !identity.Captured || identity.Validate() != nil {
			result.Message = "A build has no valid captured source manifest. Source changes are unavailable for legacy builds."
			return
		}
		if identity.Kind == "image" {
			result.Status, result.Message = "unsupported", "Image sources are compared by digest, not as text files."
			return
		}
		if identity.Repository == "" || !sourceidentity.IsCommit(identity.Commit) {
			result.Status, result.Message = "unsupported", "This source is not backed by a captured Git repository."
			return
		}
		if identity.Coverage != sourceidentity.Complete {
			result.Partial = true
			result.Reasons = append(result.Reasons, identity.Reasons...)
		}
		for _, file := range identity.Files {
			files[index][file.Path] = file
		}
	}
	paths := make(map[string]bool, len(files[0])+len(files[1]))
	for _, side := range files {
		for filePath := range side {
			paths[filePath] = true
		}
	}
	for filePath := range paths {
		oldFile, oldOK := files[0][filePath]
		newFile, newOK := files[1][filePath]
		change := "modified"
		switch {
		case !oldOK:
			change = "added"
		case !newOK:
			change = "removed"
		case oldFile.SHA256 == newFile.SHA256 && oldFile.Executable == newFile.Executable:
			continue
		}
		result.Files = append(result.Files, InstallDeploymentSourceFile{Path: filePath, Change: change, BeforeExecutable: oldFile.Executable, AfterExecutable: newFile.Executable})
	}
	slices.SortFunc(result.Files, func(a, b InstallDeploymentSourceFile) int { return strings.Compare(a.Path, b.Path) })
	slices.Sort(result.Reasons)
	result.Reasons = slices.Compact(result.Reasons)
	result.Status = "available"
}

func (s *service) deploymentSourceFile(ctx context.Context, orgID string, snapshot *deploymentSourceSnapshot, filePath string) (*string, error) {
	if snapshot == nil {
		return nil, nil
	}
	identity := snapshot.identity
	index := slices.IndexFunc(identity.Files, func(file sourceidentity.File) bool { return file.Path == filePath })
	if index < 0 {
		return nil, nil
	}
	root := identity.Root
	if path.IsAbs(root) || path.Clean(root) == ".." || strings.HasPrefix(path.Clean(root), "../") {
		return nil, fmt.Errorf("The captured source root is not a repository-relative path.")
	}
	var connected *app.ConnectedGithubVCSConfig
	var public *app.PublicGitVCSConfig
	if snapshot.componentID != "" {
		var build app.ComponentBuild
		if err := s.db.WithContext(ctx).Scopes(componenthelpers.PreloadComponentBuildConfig).
			Where(app.ComponentBuild{ID: snapshot.id, OrgID: orgID, StoredComponentID: snapshot.componentID}).First(&build).Error; err != nil {
			return nil, fmt.Errorf("The build's repository access configuration is unavailable.")
		}
		connected, public = build.ComponentConfigConnection.ConnectedGithubVCSConfig, build.ComponentConfigConnection.PublicGitVCSConfig
	} else {
		var config app.AppSandboxConfig
		if err := s.db.WithContext(ctx).Preload("ConnectedGithubVCSConfig.VCSConnection").Preload("PublicGitVCSConfig").
			Where(app.AppSandboxConfig{ID: snapshot.configID, OrgID: orgID}).First(&config).Error; err != nil {
			return nil, fmt.Errorf("The sandbox's repository access configuration is unavailable.")
		}
		connected, public = config.ConnectedGithubVCSConfig, config.PublicGitVCSConfig
	}
	var connection *app.VCSConnection
	if connected != nil && connected.OrgID == orgID && connected.Repo == identity.Repository {
		connection = &connected.VCSConnection
	} else if public == nil || public.OrgID != orgID || public.Repo != identity.Repository {
		return nil, fmt.Errorf("The captured repository does not match the build's source configuration.")
	}
	contents, err := s.appsHelpers.VCSHelpers().GetSourceFile(ctx, orgID, identity.Repository, identity.Commit, path.Join(root, filePath), connection)
	if err != nil {
		return nil, fmt.Errorf("The pinned source file could not be read. Repository access may have changed, or the file is unavailable or exceeds 512 KiB.")
	}
	if fmt.Sprintf("%x", sha256.Sum256(contents)) != identity.Files[index].SHA256 {
		return nil, fmt.Errorf("The retrieved file does not match the captured source checksum. Linked or transformed files cannot be displayed.")
	}
	if !utf8.Valid(contents) || bytes.IndexByte(contents, 0) >= 0 {
		return nil, fmt.Errorf("This is a binary file; a text diff is unavailable.")
	}
	text := string(contents)
	return &text, nil
}
