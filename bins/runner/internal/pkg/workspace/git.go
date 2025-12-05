package workspace

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"

	// errs "github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/powertoolsdev/mono/pkg/zapwriter"
)

// regex to match full git commit hash
// regex to match full git commit hash
var commitHashRegex = regexp.MustCompile(`\b[0-9a-f]{5,40}\b`)

// IsCommitHash checks if a string matches the pattern of a git commit hash
// (5-40 hexadecimal characters).
func IsCommitHash(s string) bool {
	return commitHashRegex.MatchString(s)
}

// NOTE(jm): this is only for backward compatibility with the existing Waypoint plan functionality.
func (w *workspace) isGit() bool {
	if w.Src == nil {
		return false
	}
	return w.Src.Url != emptyGithubRepoURL
}

func (w *workspace) clone(ctx context.Context) error {
	pWriter := zapwriter.New(w.L, zapcore.DebugLevel, "")

	w.L.Info("cloning repository", zap.String("url", w.Src.Url))
	repo, err := git.PlainCloneContext(ctx, w.rootDir(), false, &git.CloneOptions{
		URL:      w.Src.Url,
		Progress: pWriter,
	})
	if err != nil {
		return CloneErr{
			Url: w.Src.Url,
			Ref: w.Src.Ref,
			Err: err,
		}
	}

	w.L.Info("fetching working tree",
		zap.String("url", w.Src.Url),
		zap.String("ref", w.Src.Ref),
	)
	wtree, err := repo.Worktree()
	if err != nil {
		return CloneErr{
			Url: w.Src.Url,
			Ref: w.Src.Ref,
			Err: err,
		}
	}

	// hoist this var, like a savage
	coOpts := &git.CheckoutOptions{}

	// first, if it looks like a 40 char regex, attempt to check out as a reference w/ the hash
	if IsCommitHash(w.Src.Ref) {
		hash := plumbing.NewHash(w.Src.Ref)
		w.L.Info("checking out as reference",
			zap.String("url", w.Src.Url),
			zap.String("ref", w.Src.Ref),
			zap.String("hash", hash.String()),
		)
		coOpts = &git.CheckoutOptions{
			Hash:  hash,
			Force: true,
		}
		err = wtree.Checkout(coOpts)
		if err == nil {
			return nil
		} else {
			w.L.Error("failed to check out as reference",
				zap.String("url", w.Src.Url),
				zap.String("ref", w.Src.Ref),
				zap.String("hash", hash.String()),
				zap.String("error", err.Error()),
			)
		}
	}

	// fetch remote origin
	w.L.Debug("fetching remote origin",
		zap.String("url", w.Src.Url),
		zap.String("ref", w.Src.Ref),
	)
	remote, err := repo.Remote("origin")
	if err != nil {
		return CloneErr{
			Url: w.Src.Url,
			Ref: w.Src.Ref,
			Err: err,
		}
	}
	refSpecStr := fmt.Sprintf("refs/heads/%s:refs/heads/%s", w.Src.Ref, w.Src.Ref)
	w.L.Info("fetching remote origin",
		zap.String("url", w.Src.Url),
		zap.String("ref", w.Src.Ref),
		zap.String("ref_spec_str", refSpecStr),
	)
	err = remote.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(refSpecStr)},
	})
	if err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			w.L.Info("failed to fetch remote origin",
				zap.String("url", w.Src.Url),
				zap.String("ref", w.Src.Ref),
				zap.String("ref_spec_str", refSpecStr),
				zap.String("error", err.Error()),
			)
			// return CloneErr{
			// 	Url: w.Src.Url,
			// 	Ref: w.Src.Ref,
			// 	Err: errs.Wrap(err, "error fetching origin"),
			// }
		}
	}

	// second, attempt to check out as a branch
	branchRefName := plumbing.NewBranchReferenceName(w.Src.Ref)
	branch := plumbing.ReferenceName(branchRefName)
	w.L.Info("checking out branch",
		zap.String("url", w.Src.Url),
		zap.String("ref", w.Src.Ref),
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
	} else {
		w.L.Error("failed to check out as branch",
			zap.String("url", w.Src.Url),
			zap.String("ref", w.Src.Ref),
			zap.String("branch_ref_name", branchRefName.String()),
			zap.String("branch", branch.String()),
			zap.String("error", err.Error()),
		)
	}

	// third, attempt to check out as a tag
	tagRefName := plumbing.NewTagReferenceName(w.Src.Ref)
	w.L.Info("checking out as a tag",
		zap.String("url", w.Src.Url),
		zap.String("ref", w.Src.Ref),
		zap.String("tag_ref_name", tagRefName.String()),
	)
	coOpts = &git.CheckoutOptions{
		Branch: tagRefName,
		Force:  true,
	}
	err = wtree.Checkout(coOpts)
	if err == nil {
		return nil
	} else {
		w.L.Error("failed to check out as a tag",
			zap.String("url", w.Src.Url),
			zap.String("ref", w.Src.Ref),
			zap.String("tag_ref_name", tagRefName.String()),
			zap.String("error", err.Error()),
		)
	}

	return CloneErr{
		Url: w.Src.Url,
		Ref: w.Src.Ref,
		Err: err,
	}
}
