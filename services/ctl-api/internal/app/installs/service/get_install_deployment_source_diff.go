package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v50/github"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
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
	ctx.Header("Cache-Control", "private, no-store")
	var query struct {
		Kind string `form:"kind" binding:"required,oneof=component sandbox"`
		Name string `form:"name" binding:"required"`
		Path string `form:"path"`
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
	if message := missingSourceRecordMessage(err, "previously applied"); message != "" {
		result.Message = message
		ctx.JSON(http.StatusOK, result)
		return
	}
	if err != nil {
		ctx.Error(err)
		return
	}
	after, err := s.deploymentSourceSnapshot(ctx, org.ID, install.AppID, query.Kind, componentID, newID)
	if message := missingSourceRecordMessage(err, "selected"); message != "" {
		result.Message = message
		ctx.JSON(http.StatusOK, result)
		return
	}
	if err != nil {
		ctx.Error(err)
		return
	}
	result.Before, result.After = sourceBuildMetadata(before), sourceBuildMetadata(after)
	readCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()
	var changes map[string]deploymentSourceChange
	if (before == nil && !added) || (after == nil && !removed) {
		result.Message = "This deployment did not record both previously applied and selected build identities."
	} else {
		changes = s.populateSourceDiff(readCtx, org.ID, install.ID, &result, before, after)
	}
	if query.Path != "" && result.Status == "available" {
		change, ok := changes[query.Path]
		if !ok {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("path %q is not a changed source file", query.Path),
				Description: "path must identify a changed source file",
			})
			return
		}
		selected := &InstallDeploymentSourceFileContent{Path: query.Path}
		selected.Before, err = s.deploymentSourceFile(readCtx, org.ID, install.ID, before, change.beforePath, change.before)
		if err == nil {
			selected.After, err = s.deploymentSourceFile(readCtx, org.ID, install.ID, after, change.afterPath, change.after)
		}
		if err != nil {
			selected.Before, selected.After = nil, nil
			selected.UnavailableReason = err.Error()
		}
		result.SelectedFile = selected
	}
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
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errSourceComponentMissing
			}
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

var errSourceComponentMissing = errors.New("source component not found")

func missingSourceRecordMessage(err error, build string) string {
	switch {
	case errors.Is(err, errSourceComponentMissing):
		return "This component has been deleted. Source changes are unavailable."
	case errors.Is(err, gorm.ErrRecordNotFound):
		return fmt.Sprintf("The %s build has been deleted. Source changes are unavailable.", build)
	}
	return ""
}

func sourceReadFailureMessage(err error, fallback string) string {
	var githubErr *github.ErrorResponse
	if errors.Is(err, vcshelpers.ErrSourceCommitNotFound) ||
		(errors.As(err, &githubErr) && githubErr.Response != nil && githubErr.Response.StatusCode == http.StatusNotFound) {
		return "The pinned commit or file no longer exists in the repository."
	}
	return fallback
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

type deploymentSourceChange struct {
	beforePath, afterPath string
	before, after         *object.TreeEntry
}

const maxDeploymentSourceChanges = 1000

func (s *service) populateSourceDiff(ctx context.Context, orgID, installID string, result *InstallDeploymentSourceDiff, before, after *deploymentSourceSnapshot) map[string]deploymentSourceChange {
	var trees [2]*vcshelpers.SourceTree
	var roots [2]string
	for index, snapshot := range []*deploymentSourceSnapshot{before, after} {
		if snapshot == nil {
			continue
		}
		identity := snapshot.identity
		if identity == nil || identity.Validate() != nil {
			result.Message = "A build has no valid captured source identity. Source changes are unavailable for legacy builds."
			return nil
		}
		if identity.Kind == "image" {
			result.Status, result.Message = "unsupported", "Image sources are compared by digest, not as text files."
			return nil
		}
		if identity.Repository == "" || !sourceidentity.IsCommit(identity.Commit) || !strings.HasPrefix(identity.Digest, "git:") {
			result.Status, result.Message = "unsupported", "This source is not backed by a captured Git repository."
			return nil
		}
		if identity.Coverage != sourceidentity.Complete {
			result.Partial = true
			result.Reasons = append(result.Reasons, identity.Reasons...)
		}
		connection, err := s.deploymentSourceAccess(ctx, orgID, snapshot)
		if err == nil {
			trees[index], err = s.appsHelpers.VCSHelpers().GetSourceTree(ctx, orgID, identity.Repository, identity.Commit, connection)
		}
		if err != nil {
			s.l.Warn("unable to read deployment source tree",
				zap.String("org_id", orgID),
				zap.String("install_id", installID),
				zap.String("build_id", snapshot.id),
				zap.String("path", identity.Root),
				zap.Error(err))
			result.Message = sourceReadFailureMessage(err, "The pinned source could not be read. Repository access may have changed.")
			return nil
		}
		roots[index] = sourceidentity.CleanRoot(identity.Root)
	}
	display := roots[1]
	if after == nil {
		display = roots[0]
	}
	changes := map[string]deploymentSourceChange{}
	add := func(beforeName, afterName string, found []vcshelpers.SourceChange) {
		for _, change := range found {
			beforePath, afterPath := path.Join(beforeName, change.Path), path.Join(afterName, change.Path)
			name := afterPath
			if change.After == nil {
				name = beforePath
			}
			if rel, err := filepath.Rel("/"+display, "/"+name); err == nil && rel != "." {
				name = filepath.ToSlash(rel)
			} else {
				name = path.Base(name)
			}
			changes[name] = deploymentSourceChange{beforePath: beforePath, afterPath: afterPath, before: change.Before, after: change.After}
		}
	}
	found, truncated, err := vcshelpers.DiffSources(ctx, trees[0], roots[0], trees[1], roots[1], maxDeploymentSourceChanges)
	if err == nil {
		add(roots[0], roots[1], found)
		err = s.diffSourceInputs(ctx, trees, roots, before, after, &truncated, add)
	}
	if err != nil {
		s.l.Warn("unable to compare deployment source trees", zap.Error(err))
		result.Message = "The pinned sources could not be compared."
		return nil
	}
	if truncated {
		result.Partial = true
		result.Reasons = append(result.Reasons, "source_diff_truncated")
	}
	for name, change := range changes {
		file := InstallDeploymentSourceFile{Path: name, Change: "modified"}
		switch {
		case change.before == nil:
			file.Change = "added"
		case change.after == nil:
			file.Change = "removed"
		}
		file.BeforeExecutable = change.before != nil && change.before.Mode == filemode.Executable
		file.AfterExecutable = change.after != nil && change.after.Mode == filemode.Executable
		result.Files = append(result.Files, file)
	}
	slices.SortFunc(result.Files, func(a, b InstallDeploymentSourceFile) int { return strings.Compare(a.Path, b.Path) })
	slices.Sort(result.Reasons)
	result.Reasons = slices.Compact(result.Reasons)
	result.Status = "available"
	return changes
}

func (s *service) diffSourceInputs(ctx context.Context, trees [2]*vcshelpers.SourceTree, roots [2]string, before, after *deploymentSourceSnapshot, truncated *bool, add func(string, string, []vcshelpers.SourceChange)) error {
	inputs := map[string]bool{}
	for _, snapshot := range []*deploymentSourceSnapshot{before, after} {
		if snapshot == nil {
			continue
		}
		for _, input := range snapshot.identity.Inputs {
			if !input.Stat {
				inputs[input.Path] = true
			} else if _, ok := inputs[input.Path]; !ok {
				inputs[input.Path] = false
			}
		}
	}
	names := make([]string, 0, len(inputs))
	for name := range inputs {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		if inside(roots[0], name) || inside(roots[1], name) {
			continue
		}
		if !inputs[name] {
			b, err := trees[0].Lookup(name)
			if err != nil {
				return err
			}
			a, err := trees[1].Lookup(name)
			if err != nil {
				return err
			}
			if entryString(b, true) != entryString(a, true) {
				add(name, name, []vcshelpers.SourceChange{{Before: b, After: a}})
			}
			continue
		}
		found, cut, err := vcshelpers.DiffSources(ctx, trees[0], name, trees[1], name, maxDeploymentSourceChanges)
		if err != nil {
			return err
		}
		*truncated = *truncated || cut
		add(name, name, found)
	}
	return nil
}

func inside(root, name string) bool {
	return root == "" || name == root || strings.HasPrefix(name, root+"/")
}

func entryString(entry *object.TreeEntry, stat bool) string {
	if entry == nil {
		return ""
	}
	return sourceidentity.InputEntry(fmt.Sprintf("git:%06o:%s", uint32(entry.Mode), entry.Hash), stat)
}

func (s *service) deploymentSourceAccess(ctx context.Context, orgID string, snapshot *deploymentSourceSnapshot) (*app.VCSConnection, error) {
	identity := snapshot.identity
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
	if connected != nil && connected.OrgID == orgID && connected.Repo == identity.Repository {
		return &connected.VCSConnection, nil
	}
	if public == nil || public.OrgID != orgID || public.Repo != identity.Repository {
		return nil, fmt.Errorf("The captured repository does not match the build's source configuration.")
	}
	return nil, nil
}

func (s *service) deploymentSourceFile(ctx context.Context, orgID, installID string, snapshot *deploymentSourceSnapshot, filePath string, entry *object.TreeEntry) (*string, error) {
	if snapshot == nil || entry == nil {
		return nil, nil
	}
	if entry.Mode != filemode.Regular && entry.Mode != filemode.Executable {
		return nil, fmt.Errorf("Symlinks and submodules cannot be displayed as text.")
	}
	connection, err := s.deploymentSourceAccess(ctx, orgID, snapshot)
	if err != nil {
		return nil, err
	}
	identity := snapshot.identity
	contents, err := s.appsHelpers.VCSHelpers().GetSourceFile(ctx, orgID, identity.Repository, identity.Commit, filePath, connection)
	if err != nil {
		s.l.Warn("unable to read deployment source file",
			zap.String("org_id", orgID),
			zap.String("install_id", installID),
			zap.String("build_id", snapshot.id),
			zap.String("path", filePath),
			zap.Error(err))
		return nil, errors.New(sourceReadFailureMessage(err, "The pinned source file could not be read. Repository access may have changed, or the file is unavailable or exceeds 512 KiB."))
	}
	if plumbing.ComputeHash(plumbing.BlobObject, contents) != entry.Hash {
		return nil, fmt.Errorf("The retrieved file does not match the pinned commit.")
	}
	if !utf8.Valid(contents) || bytes.IndexByte(contents, 0) >= 0 {
		return nil, fmt.Errorf("This is a binary file; a text diff is unavailable.")
	}
	text := string(contents)
	return &text, nil
}
