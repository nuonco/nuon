package activities

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/workspace"
)

// maxGitFilePatchBytes caps the per-file patch text stored in FilePatches so
// a single large connected_repo source file can't blow up the comparison
// blob; the full repo-wide Patch text is unaffected.
const maxGitFilePatchBytes = 32 << 10

// GitDiffResult is the structured git diff between two commits.
type GitDiffResult struct {
	BaseSHA      string   `json:"base_sha"`
	HeadSHA      string   `json:"head_sha"`
	Patch        string   `json:"patch"`
	ChangedPaths []string `json:"changed_paths"`
	FilesChanged int      `json:"files_changed"`
	// FilePatches holds a per-file rendering of the patch, keyed by the
	// repo-relative path, using the same "+ "/"- "/"  " line-prefix
	// convention the dashboard's diff-rendering components expect. This lets
	// callers surface the diff for an individual file (e.g. a file inside a
	// component's connected_repo directory) without re-parsing Patch.
	FilePatches map[string]string `json:"file_patches,omitempty"`
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
		FilePatches:  filePatchesByPath(patch),
	}, nil
}

// filePatchesByPath renders each file in patch individually, keyed by its
// repo-relative path (normalized the same way as changedPathsFromPatch so
// callers can look up a file's patch using a ChangedPaths entry).
func filePatchesByPath(patch *object.Patch) map[string]string {
	if patch == nil {
		return nil
	}
	out := make(map[string]string)
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
		rendered := renderFilePatchLines(fp)
		if rendered == "" {
			continue
		}
		if len(rendered) > maxGitFilePatchBytes {
			rendered = rendered[:maxGitFilePatchBytes]
			if i := strings.LastIndex(rendered, "\n"); i >= 0 {
				rendered = rendered[:i]
			}
		}
		out[p] = rendered
	}
	return out
}

// renderFilePatchLines renders a FilePatch's chunks using the same
// "+ "/"- "/"  " line-prefix convention produced by the dashboard's
// diffLines() helper, so the result can be rendered through the same
// diff-display components without any separate parsing path.
func renderFilePatchLines(fp fdiff.FilePatch) string {
	var b strings.Builder
	for _, chunk := range fp.Chunks() {
		prefix := "  "
		switch chunk.Type() {
		case fdiff.Add:
			prefix = "+ "
		case fdiff.Delete:
			prefix = "- "
		}
		lines := strings.Split(chunk.Content(), "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		for _, line := range lines {
			b.WriteString(prefix)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
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
