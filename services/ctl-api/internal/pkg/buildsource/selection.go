package buildsource

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
)

func SourcesEqual(ctx context.Context, vcs *vcshelpers.Helpers, connected *app.ConnectedGithubVCSConfig, public *app.PublicGitVCSConfig, before, after *sourceidentity.Identity) (bool, error) {
	if before == nil || after == nil {
		return false, nil
	}
	if before.Kind != after.Kind || normalizeRepo(before.Repository) != normalizeRepo(after.Repository) || sourceidentity.CleanRoot(before.Root) != sourceidentity.CleanRoot(after.Root) {
		return false, nil
	}
	if before.ConfigChecksum != after.ConfigChecksum {
		return false, nil
	}
	if strings.HasPrefix(before.Digest, "git:") {
		if before.EligibleChecksum() == "" || after.Validate() != nil || after.Coverage != sourceidentity.Complete || !after.Captured || before.Digest != after.Digest {
			return false, nil
		}
		return inputsUnchanged(ctx, vcs, connected, public, after, before.Inputs)
	}
	if after.Commit != "" && sourceidentity.Equivalent(before, after) {
		return true, nil
	}
	if before.EligibleChecksum() == "" || after.EligibleChecksum() == "" {
		return false, nil
	}
	if after.Commit == "" {
		return before.Digest != "" && before.Digest == after.Digest, nil
	}
	if before.Commit == after.Commit {
		return true, nil
	}
	if !sourceidentity.IsCommit(before.Commit) {
		return false, nil
	}
	orgID, connection, err := sourceAccess(connected, public)
	if err != nil {
		return false, err
	}
	oldTree, err := vcs.GetSourceTree(ctx, orgID, after.Repository, before.Commit, connection)
	if err != nil {
		return false, err
	}
	oldHash, err := oldTree.Scope(after.Root)
	if err != nil {
		return false, err
	}
	return oldHash == after.Digest, nil
}

func inputsUnchanged(ctx context.Context, vcs *vcshelpers.Helpers, connected *app.ConnectedGithubVCSConfig, public *app.PublicGitVCSConfig, after *sourceidentity.Identity, inputs []sourceidentity.Input) (bool, error) {
	if len(inputs) == 0 {
		return true, nil
	}
	orgID, connection, err := sourceAccess(connected, public)
	if err != nil {
		return false, err
	}
	tree, err := vcs.GetSourceTree(ctx, orgID, after.Repository, after.Commit, connection)
	if err != nil {
		return false, err
	}
	for _, input := range inputs {
		entry, err := inputEntry(tree, input.Path, input.Stat)
		if err != nil {
			return false, err
		}
		if entry != input.Entry {
			return false, nil
		}
	}
	return true, nil
}

func inputEntry(tree *vcshelpers.SourceTree, name string, stat bool) (string, error) {
	entry, err := tree.Scope(name)
	if errors.Is(err, vcshelpers.ErrSourceDirectoryMissing) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return sourceidentity.InputEntry(entry, stat), nil
}

func ReportedSource(ctx context.Context, db *gorm.DB, vcs *vcshelpers.Helpers, l *zap.Logger, job *app.RunnerJob, success bool, i *sourceidentity.Identity, reads *sourceidentity.Reads) *sourceidentity.Identity {
	if i == nil {
		return nil
	}
	i = i.WithoutInputs()
	if !success || reads == nil {
		return i
	}
	recorded, err := recordInputs(ctx, db, vcs, job.OwnerType, job.OwnerID, i, reads)
	if err != nil {
		l.Warn("unable to record build source inputs; the build will not be reused", zap.String("runner_job_id", job.ID), zap.Error(err))
		return i
	}
	return recorded
}

func recordInputs(ctx context.Context, db *gorm.DB, vcs *vcshelpers.Helpers, ownerType, ownerID string, i *sourceidentity.Identity, reads *sourceidentity.Reads) (*sourceidentity.Identity, error) {
	if !strings.HasPrefix(i.Digest, "git:") || i.Coverage != sourceidentity.Complete {
		return i, nil
	}
	reads = sourceidentity.ExternalReads(i.Root, reads.Read, reads.Stat)
	if len(reads.Read)+len(reads.Stat) > sourceidentity.MaxInputs {
		return i.OverInputLimit()
	}
	var connected *app.ConnectedGithubVCSConfig
	var public *app.PublicGitVCSConfig
	switch ownerType {
	case "component_builds":
		var b app.ComponentBuild
		if err := db.WithContext(ctx).Preload("ComponentConfigConnection", PreloadConnection).Where(app.ComponentBuild{ID: ownerID}).First(&b).Error; err != nil {
			return nil, err
		}
		connected, public = b.ComponentConfigConnection.ConnectedGithubVCSConfig, b.ComponentConfigConnection.PublicGitVCSConfig
	case "app_sandbox_builds":
		var b app.AppSandboxBuild
		if err := db.WithContext(ctx).Preload("AppSandboxConfig.ConnectedGithubVCSConfig.VCSConnection").Preload("AppSandboxConfig.PublicGitVCSConfig").Where(app.AppSandboxBuild{ID: ownerID}).First(&b).Error; err != nil {
			return nil, err
		}
		connected, public = b.AppSandboxConfig.ConnectedGithubVCSConfig, b.AppSandboxConfig.PublicGitVCSConfig
	default:
		return nil, fmt.Errorf("source inputs require a build owner")
	}
	orgID, connection, err := sourceAccess(connected, public)
	if err != nil {
		return nil, err
	}
	tree, err := vcs.GetSourceTree(ctx, orgID, i.Repository, i.Commit, connection)
	if err != nil {
		return nil, err
	}
	inputs := make([]sourceidentity.Input, 0, len(reads.Read)+len(reads.Stat))
	for _, stat := range []bool{false, true} {
		names := reads.Read
		if stat {
			names = reads.Stat
		}
		for _, name := range names {
			if !sourceidentity.ValidInputPath(name) {
				return nil, fmt.Errorf("invalid source input path %q", name)
			}
			entry, err := inputEntry(tree, name, stat)
			if err != nil {
				return nil, err
			}
			inputs = append(inputs, sourceidentity.Input{Path: name, Stat: stat, Entry: entry})
		}
	}
	return i.WithInputs(inputs)
}

func DirectoryChanged(ctx context.Context, vcs *vcshelpers.Helpers, connected *app.ConnectedGithubVCSConfig, public *app.PublicGitVCSConfig, root, baseSHA, headSHA string) (bool, error) {
	if baseSHA == "" || headSHA == "" {
		return false, fmt.Errorf("base and head commits are required")
	}
	if baseSHA == headSHA {
		return false, nil
	}
	orgID, connection, err := sourceAccess(connected, public)
	if err != nil {
		return false, err
	}
	repository := ""
	if connected != nil {
		repository = connected.Repo
	} else {
		repository = public.Repo
	}
	base, err := vcs.GetSourceTree(ctx, orgID, repository, baseSHA, connection)
	if errors.Is(err, vcshelpers.ErrSourceCommitNotFound) || errors.Is(err, vcshelpers.ErrSourceTreeTooLarge) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	head, err := vcs.GetSourceTree(ctx, orgID, repository, headSHA, connection)
	if errors.Is(err, vcshelpers.ErrSourceTreeTooLarge) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	before, err := base.Scope(root)
	if err != nil && !errors.Is(err, vcshelpers.ErrSourceDirectoryMissing) {
		return false, err
	}
	after, err := head.Scope(root)
	if err != nil && !errors.Is(err, vcshelpers.ErrSourceDirectoryMissing) {
		return false, err
	}
	return before != after, nil
}

func sourceAccess(connected *app.ConnectedGithubVCSConfig, public *app.PublicGitVCSConfig) (string, *app.VCSConnection, error) {
	if connected != nil {
		return connected.OrgID, &connected.VCSConnection, nil
	}
	if public != nil {
		return public.OrgID, nil, nil
	}
	return "", nil, fmt.Errorf("git source has no repository config")
}
