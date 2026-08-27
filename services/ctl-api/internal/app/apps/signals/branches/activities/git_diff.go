package activities

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/workspace"
)

// GitDiffResult is the structured git diff between two commits.
type GitDiffResult struct {
	BaseSHA      string   `json:"base_sha"`
	HeadSHA      string   `json:"head_sha"`
	Patch        string   `json:"patch"`
	ChangedPaths []string `json:"changed_paths"`
	FilesChanged int      `json:"files_changed"`
}

// computeGitDiffBetweenSHAs clones the repo at headSHA, ensures baseSHA is
// present, and returns a unified patch equivalent to `git diff base..head`.
func (a *Activities) computeGitDiffBetweenSHAs(ctx context.Context, vcsConfigID, baseSHA, headSHA, workspaceID string) (*GitDiffResult, error) {
	gitSource, err := a.resolveGitSource(ctx, vcsConfigID, headSHA)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve git source: %w", err)
	}

	ws, err := workspace.New(a.v,
		workspace.WithGitSource(gitSource),
		workspace.WithID(workspaceID),
		workspace.WithCleanup(true),
		workspace.WithLogger(a.l),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}
	defer func() {
		_ = ws.Cleanup(ctx)
	}()

	if err := ws.Init(ctx); err != nil {
		return nil, fmt.Errorf("unable to init workspace: %w", err)
	}

	repo, err := git.PlainOpen(ws.Root())
	if err != nil {
		return nil, fmt.Errorf("unable to open git repo: %w", err)
	}

	if err := ensureCommitPresent(ctx, repo, baseSHA, a.l); err != nil {
		return nil, fmt.Errorf("unable to resolve base commit %s: %w", baseSHA, err)
	}
	if err := ensureCommitPresent(ctx, repo, headSHA, a.l); err != nil {
		return nil, fmt.Errorf("unable to resolve head commit %s: %w", headSHA, err)
	}

	return diffCommits(ctx, repo, baseSHA, headSHA)
}

func ensureCommitPresent(ctx context.Context, repo *git.Repository, sha string, l *zap.Logger) error {
	hash := plumbing.NewHash(sha)
	_, err := repo.CommitObject(hash)
	if err == nil {
		return nil
	}

	remote, remErr := repo.Remote("origin")
	if remErr != nil {
		return fmt.Errorf("commit not found and no origin remote: %w", err)
	}

	l.Info("fetching origin to resolve missing commit", zap.String("sha", sha))
	fetchErr := remote.FetchContext(ctx, &git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec("+refs/heads/*:refs/heads/*"),
			config.RefSpec("+refs/tags/*:refs/tags/*"),
		},
		Tags: git.AllTags,
	})
	if fetchErr != nil && fetchErr != git.NoErrAlreadyUpToDate {
		l.Warn("fetch origin failed", zap.Error(fetchErr))
	}

	_, err = repo.CommitObject(hash)
	return err
}

func diffCommits(ctx context.Context, repo *git.Repository, baseSHA, headSHA string) (*GitDiffResult, error) {
	baseCommit, err := repo.CommitObject(plumbing.NewHash(baseSHA))
	if err != nil {
		return nil, fmt.Errorf("base commit: %w", err)
	}
	headCommit, err := repo.CommitObject(plumbing.NewHash(headSHA))
	if err != nil {
		return nil, fmt.Errorf("head commit: %w", err)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("base tree: %w", err)
	}
	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("head tree: %w", err)
	}

	patch, err := baseTree.PatchContext(ctx, headTree)
	if err != nil {
		return nil, fmt.Errorf("tree patch: %w", err)
	}

	var buf strings.Builder
	if err := patch.Encode(&buf); err != nil {
		return nil, fmt.Errorf("encode patch: %w", err)
	}

	paths := changedPathsFromPatch(patch)
	return &GitDiffResult{
		BaseSHA:      baseSHA,
		HeadSHA:      headSHA,
		Patch:        buf.String(),
		ChangedPaths: paths,
		FilesChanged: len(paths),
	}, nil
}

func changedPathsFromPatch(patch *object.Patch) []string {
	if patch == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var paths []string
	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		var p string
		switch {
		case to != nil:
			p = to.Path()
		case from != nil:
			p = from.Path()
		}
		p = normalizeRepoPath(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}
	return paths
}
