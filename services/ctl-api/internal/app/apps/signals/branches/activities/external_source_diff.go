package activities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/google/go-github/v50/github"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildsource"
)

func (a *Activities) applyExternalSourceChanges(ctx context.Context, sources []componentSource, branchRepo, branchName string, out *ConfigDiffWithSourceOutput, previousSHAs map[string]string, runID string) error {
	if out == nil {
		return nil
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(8)
	changed := make([]bool, len(sources))
	for i, src := range sources {
		if src.Name == "" || src.Repo == "" || (repoURLsEqual(src.Repo, branchRepo) && src.Branch == branchName) {
			continue
		}
		group.Go(func() error {
			if err := groupCtx.Err(); err != nil {
				return err
			}
			connected, public, err := a.loadSourceVCSConfig(groupCtx, src.VCSConfigID)
			if err != nil {
				return fmt.Errorf("component %s: %w", src.Name, err)
			}
			latest, err := buildsource.ResolveRef(groupCtx, a.db, a.vcsHelpers, connected, public, runID, src.Branch)
			if err != nil {
				return fmt.Errorf("component %s: unable to resolve source commit: %w", src.Name, err)
			}
			previous := previousSHAs[src.Name]
			if previous == "" {
				changed[i] = true
				return nil
			}
			paths, err := a.changedSourcePaths(groupCtx, connected, public, previous, latest)
			if err != nil {
				return fmt.Errorf("component %s: %w", src.Name, err)
			}
			changed[i] = anyPathMatchesDirectory(paths, src.Directory)
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return err
	}
	for i, src := range sources {
		if changed[i] {
			markComponentSourceChanged(out, src.Name)
		}
	}
	return nil
}

func (a *Activities) loadSourceVCSConfig(ctx context.Context, vcsConfigID string) (*app.ConnectedGithubVCSConfig, *app.PublicGitVCSConfig, error) {
	if vcsConfigID == "" {
		return nil, nil, fmt.Errorf("vcs config id is required")
	}

	var connected app.ConnectedGithubVCSConfig
	err := a.db.WithContext(ctx).
		Preload("VCSConnection").
		Where(app.ConnectedGithubVCSConfig{ID: vcsConfigID}).First(&connected).Error
	if err == nil {
		return &connected, nil, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, fmt.Errorf("unable to load vcs config %s: %w", vcsConfigID, err)
	}

	var public app.PublicGitVCSConfig
	if err := a.db.WithContext(ctx).Where(app.PublicGitVCSConfig{ID: vcsConfigID}).First(&public).Error; err != nil {
		return nil, nil, fmt.Errorf("vcs config %s not found: %w", vcsConfigID, err)
	}
	return nil, &public, nil
}

func (a *Activities) changedSourcePaths(ctx context.Context, connected *app.ConnectedGithubVCSConfig, public *app.PublicGitVCSConfig, baseSHA, headSHA string) ([]string, error) {
	if baseSHA == "" || headSHA == "" {
		return nil, fmt.Errorf("base and head commits are required")
	}
	if baseSHA == headSHA {
		return nil, nil
	}

	var orgID, repo string
	var connection *app.VCSConnection
	if connected != nil {
		orgID, repo, connection = connected.OrgID, connected.Repo, &connected.VCSConnection
	} else if public != nil {
		orgID, repo = public.OrgID, public.Repo
	} else {
		return nil, fmt.Errorf("source has no repository config")
	}
	baseTree, err := a.vcsHelpers.GetSourceTree(ctx, orgID, repo, baseSHA, connection)
	if err != nil {
		return nil, err
	}
	headTree, err := a.vcsHelpers.GetSourceTree(ctx, orgID, repo, headSHA, connection)
	if err != nil {
		return nil, err
	}
	return changedPathsFromSourceTrees(baseTree, headTree), nil
}

func changedPathsFromSourceTrees(base, head *github.Tree) []string {
	entries := func(tree *github.Tree) map[string]*github.TreeEntry {
		out := make(map[string]*github.TreeEntry)
		for _, entry := range tree.Entries {
			if entry.GetType() != "tree" {
				out[entry.GetPath()] = entry
			}
		}
		return out
	}
	before, after := entries(base), entries(head)
	var paths []string
	for name, entry := range before {
		other, ok := after[name]
		if !ok || entry.GetSHA() != other.GetSHA() || entry.GetMode() != other.GetMode() || entry.GetType() != other.GetType() {
			paths = append(paths, name)
		}
	}
	for name := range after {
		if _, ok := before[name]; !ok {
			paths = append(paths, name)
		}
	}
	sort.Strings(paths)
	return paths
}

// comparisonChangedPaths reads the git diff already stored for the run.
// ok is false when the run has no stored diff, such as a first run with no baseline.
func (a *Activities) comparisonChangedPaths(ctx context.Context, runID string) ([]string, bool, error) {
	var comparison app.AppBranchRunComparison
	err := a.db.WithContext(ctx).
		Where(app.AppBranchRunComparison{HeadRunID: runID}).
		First(&comparison).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("unable to load comparison: %w", err)
	}
	if comparison.GitDiff == nil || !comparison.GitDiff.IsSet() {
		return nil, false, nil
	}

	raw, err := comparison.GitDiff.Get(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("unable to load git diff: %w", err)
	}
	if raw == "" {
		return nil, false, nil
	}

	var diff GitDiffResult
	if err := json.Unmarshal([]byte(raw), &diff); err != nil {
		return nil, false, fmt.Errorf("unable to parse git diff: %w", err)
	}
	return diff.ChangedPaths, true, nil
}
