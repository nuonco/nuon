package sourceidentity

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

func Checkout(ctx context.Context, repo *git.Repository, ref string, auth transport.AuthMethod) error {
	if !IsCommit(ref) {
		return fmt.Errorf("source requires a full immutable commit")
	}
	hash := plumbing.NewHash(ref)
	if _, err := repo.CommitObject(hash); err != nil {
		fetchErr := repo.FetchContext(ctx, &git.FetchOptions{Auth: auth, RefSpecs: []config.RefSpec{config.RefSpec("+" + ref + ":refs/nuon/captured")}})
		if fetchErr != nil && fetchErr != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("fetch captured commit: %w", fetchErr)
		}
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	if err := wt.Checkout(&git.CheckoutOptions{Hash: hash, Force: true}); err != nil {
		return fmt.Errorf("checkout captured commit: %w", err)
	}
	head, err := repo.Head()
	if err != nil {
		return err
	}
	if head.Hash() != hash {
		return fmt.Errorf("checkout does not match captured commit")
	}
	return nil
}
