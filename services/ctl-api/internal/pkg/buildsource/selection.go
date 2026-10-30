package buildsource

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/google/go-github/v50/github"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
)

var errSourceDirectoryMissing = errors.New("source directory is missing from pinned commit")

func SourcesEqual(ctx context.Context, vcs *vcshelpers.Helpers, connected *app.ConnectedGithubVCSConfig, public *app.PublicGitVCSConfig, before, after *sourceidentity.Identity) (bool, error) {
	if before == nil || after == nil {
		return false, nil
	}
	if before.Kind != after.Kind || normalizeRepo(before.Repository) != normalizeRepo(after.Repository) || path.Clean(before.Root) != path.Clean(after.Root) {
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
	oldHash, err := treeScope(oldTree, after.Root)
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
	if err != nil {
		return false, err
	}
	head, err := vcs.GetSourceTree(ctx, orgID, repository, headSHA, connection)
	if err != nil {
		return false, err
	}
	before, err := treeScope(base, root)
	if err != nil && !errors.Is(err, errSourceDirectoryMissing) {
		return false, err
	}
	after, err := treeScope(head, root)
	if err != nil && !errors.Is(err, errSourceDirectoryMissing) {
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

func treeScope(tree *github.Tree, root string) (string, error) {
	root = path.Clean(root)
	if root == "." {
		return "git:040000:" + tree.GetSHA(), nil
	}
	if path.IsAbs(root) || root == ".." || strings.HasPrefix(root, "../") {
		return "", fmt.Errorf("source directory escapes repository")
	}
	for _, entry := range tree.Entries {
		name := entry.GetPath()
		if entry.GetMode() == "120000" && strings.HasPrefix(root, name+"/") {
			return "git:120000:" + entry.GetSHA(), nil
		}
	}
	for _, entry := range tree.Entries {
		if entry.GetPath() == root {
			if entry.GetType() != "tree" && entry.GetType() != "blob" && entry.GetType() != "commit" {
				return "", fmt.Errorf("source root is not a Git directory or file")
			}
			return "git:" + entry.GetMode() + ":" + entry.GetSHA(), nil
		}
	}
	return "", fmt.Errorf("%w: %s", errSourceDirectoryMissing, root)
}
