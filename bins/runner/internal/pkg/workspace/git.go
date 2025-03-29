package workspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	errs "github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/zapwriter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NOTE(jm): this is only for backward compatibility with the existing Waypoint plan functionality.
func (w *workspace) isGit() bool {
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

	w.L.Info("fetching working tree")
	wtree, err := repo.Worktree()
	if err != nil {
		return CloneErr{
			Url: w.Src.Url,
			Ref: w.Src.Ref,
			Err: err,
		}
	}

	// first, attempt to check out as a reference
	w.L.Info("checking out as reference")
	coOpts := &git.CheckoutOptions{
		Hash:  plumbing.NewHash(w.Src.Ref),
		Force: true,
	}
	err = wtree.Checkout(coOpts)
	if err == nil {
		return nil
	}

	w.L.Info("fetching remote branch")
	remote, err := repo.Remote("origin")
	if err != nil {
		return CloneErr{
			Url: w.Src.Url,
			Ref: w.Src.Ref,
			Err: err,
		}
	}

	refSpecStr := fmt.Sprintf("refs/heads/%s:refs/heads/%s", w.Src.Ref, w.Src.Ref)
	if err = remote.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(refSpecStr)},
	}); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return CloneErr{
				Url: w.Src.Url,
				Ref: w.Src.Ref,
				Err: errs.Wrap(err, "error fetching origin"),
			}
		}
	}

	// second, attempt to check out as a branch
	w.L.Info("checking out branch")
	branchRefName := plumbing.NewBranchReferenceName(w.Src.Ref)
	coOpts = &git.CheckoutOptions{
		Branch: plumbing.ReferenceName(branchRefName),
		Force:  true,
	}
	err = wtree.Checkout(coOpts)
	if err == nil {
		return nil
	}

	// third, attempt to check out as a tag
	w.L.Info("checking out as a tag")
	tagRefName := plumbing.NewTagReferenceName(w.Src.Ref)
	coOpts = &git.CheckoutOptions{
		Branch: tagRefName,
		Force:  true,
	}
	err = wtree.Checkout(coOpts)
	if err == nil {
		return nil
	}

	return CloneErr{
		Url: w.Src.Url,
		Ref: w.Src.Ref,
		Err: err,
	}
}
