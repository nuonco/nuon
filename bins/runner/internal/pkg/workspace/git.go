package workspace

import (
	"context"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// NOTE(jm): this is only for backward compatibility with the existing Waypoint plan functionality.
func (w *workspace) isGit() bool {
	return w.Src.Url != emptyGithubRepoURL
}

func (w *workspace) clone(ctx context.Context) error {
	repo, err := git.PlainCloneContext(ctx, w.rootDir(), false, &git.CloneOptions{
		URL: w.Src.Url,
	})
	if err != nil {
		return CloneErr{
			Url: w.Src.Url,
			Ref: w.Src.Ref,
			Err: err,
		}
	}

	wtree, err := repo.Worktree()
	if err != nil {
		return CloneErr{
			Url: w.Src.Url,
			Ref: w.Src.Ref,
			Err: err,
		}
	}

	// first, attempt to check out as a reference
	coOpts := &git.CheckoutOptions{
		Hash:  plumbing.NewHash(w.Src.Ref),
		Force: true,
	}
	if err := wtree.Checkout(coOpts); err == nil {
		return nil
	}

	// second, attempt to check out as a branch
	coOpts = &git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(w.Src.Ref),
		Force:  true,
	}
	if err := wtree.Checkout(coOpts); err == nil {
		return nil
	}

	return CloneErr{
		Url: w.Src.Url,
		Ref: w.Src.Ref,
		Err: err,
	}
}
