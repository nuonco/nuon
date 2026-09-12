package workspace

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nuonco/nuon/pkg/zapwriter"
)

var (
	commitHashRegex    = regexp.MustCompile(`\b[0-9a-f]{5,40}\b`)
	fullCommitSHARegex = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

// IsCommitHash checks if a string matches the pattern of a git commit hash
// (5-40 hexadecimal characters).
func IsCommitHash(s string) bool {
	return commitHashRegex.MatchString(s)
}

// IsFullCommitSHA reports whether s is a complete 40 character git object ID.
// Unlike IsCommitHash this is anchored, so a branch named `deadbeef` is not
// mistaken for a commit.
func IsFullCommitSHA(s string) bool {
	return fullCommitSHARegex.MatchString(s)
}

// clone fetches the source into the workspace root.
//
// A ref that names a branch or tag is cloned shallow (depth 1, single branch,
// no tags), which is the common case and avoids downloading the repo's whole
// history. A full commit SHA cannot take that path: go-git does not implement
// the want-SHA capability, so ReferenceName has to resolve to a real ref and
// the commit is only guaranteed to be present after an unshallow clone.
func (w *Workspace) clone(ctx context.Context) error {
	if !IsFullCommitSHA(w.src.Ref) {
		if err := w.shallowClone(ctx); err == nil {
			return nil
		}
		w.l.Info("shallow clone failed, falling back to full clone",
			zap.String("url", w.src.URL),
			zap.String("ref", w.src.Ref),
		)
	}

	return w.fullClone(ctx)
}

// shallowClone clones just the tip of a single branch or tag. The ref is left
// checked out on success, so the caller has nothing further to do.
func (w *Workspace) shallowClone(ctx context.Context) error {
	pWriter := zapwriter.New(w.l, zapcore.DebugLevel, "")

	refNames := []plumbing.ReferenceName{""}
	if w.src.Ref != "" {
		refNames = []plumbing.ReferenceName{
			plumbing.NewBranchReferenceName(w.src.Ref),
			plumbing.NewTagReferenceName(w.src.Ref),
		}
	}

	var err error
	for _, refName := range refNames {
		w.l.Info("shallow cloning repository",
			zap.String("url", w.src.URL),
			zap.String("ref", w.src.Ref),
			zap.String("ref_name", refName.String()),
		)

		_, err = git.PlainCloneContext(ctx, w.rootDir(), false, &git.CloneOptions{
			URL:           w.src.URL,
			ReferenceName: refName,
			SingleBranch:  true,
			Depth:         1,
			Tags:          git.NoTags,
			Progress:      pWriter,
		})
		if err == nil {
			return nil
		}

		w.l.Debug("shallow clone attempt failed",
			zap.String("url", w.src.URL),
			zap.String("ref_name", refName.String()),
			zap.Error(err),
		)

		// A partially written clone would make the next attempt fail with
		// ErrRepositoryAlreadyExists rather than surfacing the real error.
		if cleanupErr := w.cleanupExistingDir(); cleanupErr != nil {
			return cleanupErr
		}
	}

	return CloneErr{
		Url: w.src.URL,
		Ref: w.src.Ref,
		Err: err,
	}
}

func (w *Workspace) fullClone(ctx context.Context) error {
	pWriter := zapwriter.New(w.l, zapcore.DebugLevel, "")

	w.l.Info("cloning repository", zap.String("url", w.src.URL))
	repo, err := git.PlainCloneContext(ctx, w.rootDir(), false, &git.CloneOptions{
		URL:      w.src.URL,
		Progress: pWriter,
	})
	if err != nil {
		return CloneErr{
			Url: w.src.URL,
			Ref: w.src.Ref,
			Err: err,
		}
	}

	w.l.Info("fetching working tree",
		zap.String("url", w.src.URL),
		zap.String("ref", w.src.Ref),
	)
	wtree, err := repo.Worktree()
	if err != nil {
		return CloneErr{
			Url: w.src.URL,
			Ref: w.src.Ref,
			Err: err,
		}
	}

	coOpts := &git.CheckoutOptions{}

	// first, if it looks like a commit hash, attempt to check out as a reference
	if IsCommitHash(w.src.Ref) {
		hash := plumbing.NewHash(w.src.Ref)
		w.l.Info("checking out as reference",
			zap.String("url", w.src.URL),
			zap.String("ref", w.src.Ref),
			zap.String("hash", hash.String()),
		)
		coOpts = &git.CheckoutOptions{
			Hash:  hash,
			Force: true,
		}
		err = wtree.Checkout(coOpts)
		if err == nil {
			return nil
		}
		w.l.Error("failed to check out as reference",
			zap.String("url", w.src.URL),
			zap.String("ref", w.src.Ref),
			zap.String("hash", hash.String()),
			zap.String("error", err.Error()),
		)
	}

	// fetch remote origin
	w.l.Debug("fetching remote origin",
		zap.String("url", w.src.URL),
		zap.String("ref", w.src.Ref),
	)
	remote, err := repo.Remote("origin")
	if err != nil {
		return CloneErr{
			Url: w.src.URL,
			Ref: w.src.Ref,
			Err: err,
		}
	}
	refSpecStr := fmt.Sprintf("refs/heads/%s:refs/heads/%s", w.src.Ref, w.src.Ref)
	w.l.Info("fetching remote origin",
		zap.String("url", w.src.URL),
		zap.String("ref", w.src.Ref),
		zap.String("ref_spec_str", refSpecStr),
	)
	err = remote.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(refSpecStr)},
	})
	if err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			w.l.Info("failed to fetch remote origin",
				zap.String("url", w.src.URL),
				zap.String("ref", w.src.Ref),
				zap.String("ref_spec_str", refSpecStr),
				zap.String("error", err.Error()),
			)
		}
	}

	// second, attempt to check out as a branch
	branchRefName := plumbing.NewBranchReferenceName(w.src.Ref)
	branch := plumbing.ReferenceName(branchRefName)
	w.l.Info("checking out branch",
		zap.String("url", w.src.URL),
		zap.String("ref", w.src.Ref),
		zap.String("branch_ref_name", branchRefName.String()),
		zap.String("branch", branch.String()),
	)
	coOpts = &git.CheckoutOptions{
		Branch: branch,
		Force:  true,
	}
	err = wtree.Checkout(coOpts)
	if err == nil {
		return nil
	}
	w.l.Error("failed to check out as branch",
		zap.String("url", w.src.URL),
		zap.String("ref", w.src.Ref),
		zap.String("branch_ref_name", branchRefName.String()),
		zap.String("branch", branch.String()),
		zap.String("error", err.Error()),
	)

	// third, attempt to check out as a tag
	tagRefName := plumbing.NewTagReferenceName(w.src.Ref)
	w.l.Info("checking out as a tag",
		zap.String("url", w.src.URL),
		zap.String("ref", w.src.Ref),
		zap.String("tag_ref_name", tagRefName.String()),
	)
	coOpts = &git.CheckoutOptions{
		Branch: tagRefName,
		Force:  true,
	}
	err = wtree.Checkout(coOpts)
	if err == nil {
		return nil
	}
	w.l.Error("failed to check out as a tag",
		zap.String("url", w.src.URL),
		zap.String("ref", w.src.Ref),
		zap.String("tag_ref_name", tagRefName.String()),
		zap.String("error", err.Error()),
	)

	return CloneErr{
		Url: w.src.URL,
		Ref: w.src.Ref,
		Err: err,
	}
}
