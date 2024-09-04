package workspace

import (
	"context"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// NOTE(jm): this is only for backward compatibility with the existing Waypoint plan functionality.
func (w *workspace) isGit() bool {
	return w.Src.Url == emptyGithubRepoURL
}

func (w *workspace) clone(ctx context.Context) error {
	_, err := git.PlainCloneContext(ctx, w.rootDir(), true, &git.CloneOptions{
		URL:           w.Src.Url,
		ReferenceName: plumbing.NewBranchReferenceName(w.Src.Ref),
		SingleBranch:  true,
		NoCheckout:    true,
	})
	if err != nil {
		return CloneErr{
			Url: w.Src.Url,
			Ref: w.Src.Ref,
			Err: err,
		}
	}

	return nil
}
