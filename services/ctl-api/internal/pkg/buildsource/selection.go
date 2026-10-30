package buildsource

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/google/go-github/v50/github"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
)

func SourcesEqual(ctx context.Context, vcs *vcshelpers.Helpers, connected *app.ConnectedGithubVCSConfig, public *app.PublicGitVCSConfig, before, after *sourceidentity.Identity) (bool, error) {
	if before.EligibleChecksum() == "" || after == nil || after.Validate() != nil {
		return false, nil
	}
	if before.Kind != after.Kind || normalizeRepo(before.Repository) != normalizeRepo(after.Repository) || path.Clean(before.Root) != path.Clean(after.Root) {
		return false, nil
	}
	if after.Commit == "" {
		return before.Digest != "" && before.Digest == after.Digest && after.Captured, nil
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
	newTree, err := vcs.GetSourceTree(ctx, orgID, after.Repository, after.Commit, connection)
	if err != nil {
		return false, err
	}
	root := path.Clean(after.Root)
	oldHash, err := treeScope(oldTree, root)
	if err != nil {
		return false, err
	}
	newHash, err := treeScope(newTree, root)
	if err != nil || oldHash != newHash {
		return false, err
	}
	return true, nil
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
		return tree.GetSHA(), nil
	}
	if path.IsAbs(root) || root == ".." || strings.HasPrefix(root, "../") {
		return "", fmt.Errorf("source directory escapes repository")
	}
	for _, entry := range tree.Entries {
		name := entry.GetPath()
		if entry.GetMode() == "120000" && (name == root || strings.HasPrefix(root, name+"/") || strings.HasPrefix(name, root+"/")) {
			return tree.GetSHA(), nil
		}
	}
	for _, entry := range tree.Entries {
		if entry.GetPath() == root {
			if entry.GetType() == "blob" {
				return entry.GetMode() + ":" + entry.GetSHA(), nil
			}
			if entry.GetType() != "tree" {
				return "", fmt.Errorf("source root is not a Git directory or file")
			}
			return entry.GetSHA(), nil
		}
	}
	return "", fmt.Errorf("source directory %s is missing from pinned commit", root)
}
