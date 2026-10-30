/*
Package buildsource prepares pinned build inputs and checks build reuse.

Git-backed component and sandbox flow:

	Ref -> pinned commit -> configured directory's tree/blob ID
	                                  |
	                    Compare source + build config
	                   with a verified successful build
	                                  |
	             +--------------------+--------------------+
	             |                                         |
	          Unchanged                     Changed / insufficient evidence
	             |                                         |
	         Reuse build                     Runner checkout at pinned commit
	                                                       |
	                                            sourceidentity.Verify
	                                                       |
	                                                  Build tools

Selection uses cached Git metadata, not checkouts or blob downloads. The configured
directory is the change-detection boundary; dependencies outside it are not tracked.
Runner verification hashes the already-required checkout before build tools modify it.
*/
package buildsource

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/oci/imageref"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
)

func Component(ctx context.Context, db *gorm.DB, vcs *vcshelpers.Helpers, b *app.ComponentBuild) (*sourceidentity.Identity, error) {
	c := &b.ComponentConfigConnection
	kind := string(c.Type)
	if c.TerraformModuleComponentConfig != nil {
		kind = "terraform"
	}
	if c.HelmComponentConfig != nil {
		kind = "helm"
	}
	if c.PulumiComponentConfig != nil {
		kind = "pulumi"
	}
	if c.KubernetesManifestComponentConfig != nil {
		kind = "inline"
		if cfg := c.KubernetesManifestComponentConfig.Kustomize; cfg != nil && cfg.Path != "" {
			kind = "kustomize"
		}
	}
	checksum, err := ComponentChecksum(c)
	if err != nil {
		return nil, err
	}
	i := sourceidentity.New(kind, checksum)
	if b.SourceIdentity != nil {
		return b.SourceIdentity, b.SourceIdentity.Validate()
	}
	if c.DockerBuildComponentConfig != nil {
		i.MarkUnknown("deprecated_docker_build_unsupported")
	}
	if cfg := c.ExternalImageComponentConfig; cfg != nil {
		i.Kind = "image"
		ref := (imageref.Spec{Image: cfg.ImageURL, Tag: cfg.Tag, UpdatePolicy: cfg.UpdatePolicy}).PullRef("")
		if strings.HasPrefix(ref, "sha256:") {
			i.Digest = ref
			i.Captured = true
		} else {
			i.MarkUnknown("source_requires_runner_resolution")
		}
		i.Seal()
		return i, nil
	}
	if kind == "inline" {
		i.Digest = sourceidentity.Hash(c.KubernetesManifestComponentConfig.Manifest)
		i.Captured = true
		i.Seal()
		return i, nil
	}
	connected, public := c.ConnectedGithubVCSConfig, c.PublicGitVCSConfig
	if connected == nil && public == nil {
		i.MarkUnknown("source_requires_runner_resolution")
		i.Seal()
		return i, nil
	}
	ref := generics.FromPtrStr(b.GitRef)
	if b.VCSConnectionCommit != nil && b.VCSConnectionCommit.SHA != "" {
		ref = b.VCSConnectionCommit.SHA
	}
	return prepareGit(ctx, db, vcs, i, connected, public, b.AppBranchRunID, ref)
}

func Sandbox(ctx context.Context, db *gorm.DB, vcs *vcshelpers.Helpers, b *app.AppSandboxBuild) (*sourceidentity.Identity, error) {
	if b.SourceIdentity != nil {
		return b.SourceIdentity, b.SourceIdentity.Validate()
	}
	c := &b.AppSandboxConfig
	i := sourceidentity.New("terraform", SandboxChecksum(c))
	if c.Type == "pulumi" {
		i.Kind = "pulumi"
	}
	ref := ""
	if b.VCSConnectionCommit != nil {
		ref = b.VCSConnectionCommit.SHA
	}
	return prepareGit(ctx, db, vcs, i, c.ConnectedGithubVCSConfig, c.PublicGitVCSConfig, b.AppBranchRunID, ref)
}

func prepareGit(ctx context.Context, db *gorm.DB, vcs *vcshelpers.Helpers, i *sourceidentity.Identity, connected *app.ConnectedGithubVCSConfig, public *app.PublicGitVCSConfig, runID *string, ref string) (*sourceidentity.Identity, error) {
	branch := ""
	if connected != nil {
		i.Repository = connected.Repo
		i.Root = connected.Directory
		branch = connected.Branch
	}
	if public != nil {
		i.Repository = public.Repo
		i.Root = public.Directory
		branch = public.Branch
	}
	if connected == nil && public == nil {
		i.MarkUnknown("git_source_missing")
		i.Seal()
		return i, nil
	}
	if ref == "" {
		ref = branch
	}
	if runID != nil && *runID != "" && (ref == branch || ref == "") {
		var run app.AppBranchRun
		if err := db.WithContext(ctx).Preload("VCSConnectionCommit").Preload("AppBranchConfig.ConnectedGithubVCSConfig").Preload("AppBranchConfig.PublicGitVCSConfig").Where(app.AppBranchRun{ID: *runID}).First(&run).Error; err != nil {
			return nil, err
		}
		repo, runBranch := "", ""
		if cfg := run.AppBranchConfig.ConnectedGithubVCSConfig; cfg != nil {
			repo, runBranch = cfg.Repo, cfg.Branch
		}
		if cfg := run.AppBranchConfig.PublicGitVCSConfig; cfg != nil {
			repo, runBranch = cfg.Repo, cfg.Branch
		}
		if normalizeRepo(repo) == normalizeRepo(i.Repository) && branch == runBranch {
			if run.VCSConnectionCommit != nil {
				ref = run.VCSConnectionCommit.SHA
			} else if run.HeadSHA != "" {
				ref = run.HeadSHA
			}
		}
	}
	if !sourceidentity.IsCommit(ref) {
		var err error
		ref, err = ResolveRef(ctx, db, vcs, connected, public, generics.FromPtrStr(runID), ref)
		if err != nil {
			return nil, err
		}
	}
	if !sourceidentity.IsCommit(ref) {
		return nil, fmt.Errorf("source did not resolve to a full commit SHA")
	}
	i.Commit = ref
	i.Root = filepath.ToSlash(filepath.Clean(i.Root))
	if i.Root == "." {
		i.Root = ""
	}
	orgID, connection, err := sourceAccess(connected, public)
	if err != nil {
		return nil, err
	}
	tree, err := vcs.GetSourceTree(ctx, orgID, i.Repository, i.Commit, connection)
	if errors.Is(err, vcshelpers.ErrSourceTreeTruncated) {
		i.MarkUnknown("git_tree_metadata_truncated")
		i.Seal()
		return i, i.Validate()
	}
	if err != nil {
		return nil, err
	}
	for _, entry := range tree.Entries {
		if entry.GetMode() == "160000" && (i.Root == "" || entry.GetPath() == i.Root || strings.HasPrefix(entry.GetPath(), i.Root+"/")) {
			i.MarkUnknown("git_submodules_not_materialized")
			i.Seal()
			return i, i.Validate()
		}
	}
	i.Digest, err = treeScope(tree, i.Root)
	if err != nil {
		return nil, err
	}
	i.Captured = true
	i.Seal()
	return i, i.Validate()
}

func GitSource(ctx context.Context, vcs *vcshelpers.Helpers, connected *app.ConnectedGithubVCSConfig, public *app.PublicGitVCSConfig, i *sourceidentity.Identity) (*plantypes.GitSource, error) {
	if i == nil || i.Commit == "" {
		return nil, nil
	}
	if !sourceidentity.IsCommit(i.Commit) {
		return nil, fmt.Errorf("invalid pinned source commit")
	}
	if connected != nil {
		return vcs.GetGitSourceAtCommit(ctx, connected, i.Commit)
	}
	if public != nil {
		return vcs.GetPublicGitSourceAtCommit(ctx, public, i.Commit)
	}
	return nil, fmt.Errorf("pinned source has no repository config")
}

func SaveComponent(ctx context.Context, db *gorm.DB, id string, identity *sourceidentity.Identity) (*sourceidentity.Identity, error) {
	if err := identity.Validate(); err != nil {
		return nil, err
	}
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current app.ComponentBuild
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.ComponentBuild{ID: id}).First(&current).Error; err != nil {
			return err
		}
		if current.SourceIdentity != nil {
			if !sourceidentity.Matches(current.SourceIdentity, identity) {
				return fmt.Errorf("component build source was already pinned to different inputs")
			}
			return nil
		}
		update := app.ComponentBuild{SourceIdentity: identity, Checksum: identity.ConfigChecksum}
		if identity.Commit != "" {
			update.GitRef = &identity.Commit
		}
		return tx.Model(&current).Select("source_identity", "source_checksum", "checksum", "git_ref").Updates(update).Error
	})
	return identity, err
}

func SaveSandbox(ctx context.Context, db *gorm.DB, id string, identity *sourceidentity.Identity) (*sourceidentity.Identity, error) {
	if err := identity.Validate(); err != nil {
		return nil, err
	}
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current app.AppSandboxBuild
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.AppSandboxBuild{ID: id}).First(&current).Error; err != nil {
			return err
		}
		if current.SourceIdentity != nil {
			if !sourceidentity.Matches(current.SourceIdentity, identity) {
				return fmt.Errorf("sandbox build source was already pinned to different inputs")
			}
			return nil
		}
		return tx.Model(&current).Select("source_identity", "source_checksum", "checksum").Updates(app.AppSandboxBuild{SourceIdentity: identity, Checksum: identity.ConfigChecksum}).Error
	})
	return identity, err
}

func normalizeRepo(repo string) string {
	repo = strings.TrimSuffix(strings.TrimSuffix(repo, "/"), ".git")
	for _, prefix := range []string{"https://github.com/", "http://github.com/", "git@github.com:"} {
		repo = strings.TrimPrefix(repo, prefix)
	}
	return strings.ToLower(repo)
}

func PreloadConnection(db *gorm.DB) *gorm.DB {
	for _, kind := range []string{"TerraformModuleComponentConfig", "HelmComponentConfig", "KubernetesManifestComponentConfig", "PulumiComponentConfig"} {
		db = db.Preload(kind).Preload(kind + ".PublicGitVCSConfig").Preload(kind + ".ConnectedGithubVCSConfig").Preload(kind + ".ConnectedGithubVCSConfig.VCSConnection")
	}
	return db.Preload("ExternalImageComponentConfig")
}
