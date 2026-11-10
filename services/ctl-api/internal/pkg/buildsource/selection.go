package buildsource

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	if after.Commit != "" && sourceidentity.Equivalent(before, after) {
		return true, nil
	}
	if before.EligibleChecksum() == "" || after.EligibleChecksum() == "" {
		return false, nil
	}
	if after.Commit == "" {
		return before.Digest != "" && before.Digest == after.Digest, nil
	}
	if strings.HasPrefix(before.Digest, "git:") {
		return before.Digest == after.Digest, nil
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
	if errors.Is(err, vcshelpers.ErrSourceCommitNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	head, err := vcs.GetSourceTree(ctx, orgID, repository, headSHA, connection)
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
