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

var commitHashRegex = regexp.MustCompile(`\b[0-9a-f]{5,40}\b`)

// IsCommitHash checks if a string matches the pattern of a git commit hash
// (5-40 hexadecimal characters).
func IsCommitHash(s string) bool {
	return commitHashRegex.MatchString(s)
}

func (w *Workspace) clone(ctx context.Context) error {
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
