package activities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

// applyExternalSourceChanges marks components whose repo is not the app branch
// repo. The latest commit on the component's own branch is diffed against the
// commit stored on the previous build, and only paths under the component
// directory count.
func (a *Activities) applyExternalSourceChanges(ctx context.Context, sources []componentSource, branchRepo string, out *ConfigDiffWithSourceOutput, previousSHAs map[string]string) error {
	if out == nil {
		return nil
	}
	if out.ComponentSourceChanged == nil {
		out.ComponentSourceChanged = map[string]bool{}
	}

	cache := map[string][]string{}
	for _, src := range sources {
		if src.Name == "" || src.Repo == "" || repoURLsEqual(src.Repo, branchRepo) {
			continue
		}
		if src.VCSConfigID == "" {
			return fmt.Errorf("component %s source %s has no vcs config", src.Name, src.Repo)
		}

		latest, err := a.latestVCSCommitSHA(ctx, src.VCSConfigID, src.Branch)
		if err != nil {
			return fmt.Errorf("component %s: %w", src.Name, err)
		}
		previous := previousSHAs[src.Name]
		if previous == "" {
			markComponentSourceChanged(out, src.Name)
			continue
		}
		if previous == latest {
			continue
		}

		key := src.VCSConfigID + "\n" + previous + "\n" + latest
		paths, ok := cache[key]
		if !ok {
			paths, err = a.changedPathsBetween(ctx, src.VCSConfigID, previous, latest)
			if err != nil {
				return fmt.Errorf("component %s: %w", src.Name, err)
			}
			cache[key] = paths
		}
		if anyPathMatchesDirectory(paths, src.Directory) {
			markComponentSourceChanged(out, src.Name)
		}
	}
	return nil
}

func (a *Activities) latestVCSCommitSHA(ctx context.Context, vcsConfigID, branch string) (string, error) {
	if vcsConfigID == "" {
		return "", fmt.Errorf("vcs config id is required")
	}

	var connected app.ConnectedGithubVCSConfig
	err := a.db.WithContext(ctx).
		Preload("VCSConnection").
		First(&connected, "id = ?", vcsConfigID).Error
	if err == nil {
		ref := branch
		if ref == "" {
			ref = connected.Branch
		}
		commit, fetchErr := a.vcsHelpers.GetConnectedGithubVCSConfigCommit(ctx, &connected, ref)
		if fetchErr != nil {
			return "", fmt.Errorf("unable to resolve commit for %s@%s: %w", connected.Repo, ref, fetchErr)
		}
		if commit == nil {
			return "", fmt.Errorf("commit for %s@%s was empty", connected.Repo, ref)
		}
		sha := generics.FromPtrStr(commit.SHA)
		if sha == "" {
			return "", fmt.Errorf("commit for %s@%s has no sha", connected.Repo, ref)
		}
		return sha, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("unable to load vcs config %s: %w", vcsConfigID, err)
	}

	var public app.PublicGitVCSConfig
	if err := a.db.WithContext(ctx).First(&public, "id = ?", vcsConfigID).Error; err != nil {
		return "", fmt.Errorf("vcs config %s not found: %w", vcsConfigID, err)
	}
	ref := branch
	if ref == "" {
		ref = public.Branch
	}
	commit, fetchErr := a.vcsHelpers.GetPublicGitVCSConfigCommit(ctx, &public, ref)
	if fetchErr != nil {
		return "", fmt.Errorf("unable to resolve commit for %s@%s: %w", public.Repo, ref, fetchErr)
	}
	if commit == nil {
		return "", fmt.Errorf("commit for %s@%s was empty", public.Repo, ref)
	}
	sha := generics.FromPtrStr(commit.SHA)
	if sha == "" {
		return "", fmt.Errorf("commit for %s@%s has no sha", public.Repo, ref)
	}
	return sha, nil
}

func (a *Activities) changedPathsBetween(ctx context.Context, vcsConfigID, baseSHA, headSHA string) ([]string, error) {
	if baseSHA == "" || headSHA == "" {
		return nil, fmt.Errorf("base and head commits are required")
	}
	if baseSHA == headSHA {
		return nil, nil
	}

	diff, err := a.computeGitDiffBetweenSHAs(ctx, vcsConfigID, baseSHA, headSHA, fmt.Sprintf("source-diff-%s-%s-%s", vcsConfigID, baseSHA, headSHA))
	if err != nil {
		return nil, err
	}
	if diff == nil {
		return nil, nil
	}
	return diff.ChangedPaths, nil
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

	raw, err := comparison.GitDiff.Get(blobstore.WithBlobService(ctx, a.blobSvc))
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
