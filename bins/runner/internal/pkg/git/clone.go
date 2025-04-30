package git

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	errs "github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/zapwriter"
)

func Clone(ctx context.Context, rootDir string, src *plantypes.GitSource, l *zap.Logger) error {
	cl := &workspace{}
	return cl.clone(ctx, rootDir, src, l)
}

type workspace struct{}

func (w *workspace) clone(ctx context.Context, rootDir string, src *plantypes.GitSource, l *zap.Logger) error {
	pWriter := zapwriter.New(l, zapcore.DebugLevel, "")

	l.Info("cloning repository", zap.String("url", src.URL))
	repo, err := git.PlainCloneContext(ctx, rootDir, false, &git.CloneOptions{
		URL:      src.URL,
		Progress: pWriter,
	})
	if err != nil {
		l.Error("error cloning repository",
			zap.String("url", src.URL),
			zap.Error(err),
		)
		return CloneErr{
			Url: src.URL,
			Ref: src.Ref,
			Err: err,
		}
	}

	l.Info("fetching working tree")
	wtree, err := repo.Worktree()
	if err != nil {
		l.Error("error fetching working tree",
			zap.String("url", src.URL),
			zap.Error(err),
		)
		return CloneErr{
			Url: src.URL,
			Ref: src.Ref,
			Err: err,
		}
	}

	// first, attempt to check out as a reference
	l.Info("checking out as reference")
	coOpts := &git.CheckoutOptions{
		Hash:  plumbing.NewHash(src.Ref),
		Force: true,
	}
	err = wtree.Checkout(coOpts)
	if err == nil {
		return nil
	}

	l.Info("fetching remote origin")
	remote, err := repo.Remote("origin")
	if err != nil {
		l.Error("error fetching remote origin",
			zap.String("url", src.URL),
			zap.Error(err),
		)
		return CloneErr{
			Url: src.URL,
			Ref: src.Ref,
			Err: err,
		}
	}

	refSpecStr := fmt.Sprintf("refs/heads/%s:refs/heads/%s", src.Ref, src.Ref)
	if err = remote.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(refSpecStr)},
	}); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			l.Error("Unable to fetch reference",
				zap.String("ref", src.Ref),
				zap.String("url", src.URL),
				zap.Error(err),
			)
			return CloneErr{
				Url: src.URL,
				Ref: src.Ref,
				Err: errs.Wrap(err, "error fetching reference"),
			}
		}
	}

	// second, attempt to check out as a branch
	l.Info("checking out branch")
	branchRefName := plumbing.NewBranchReferenceName(src.Ref)
	coOpts = &git.CheckoutOptions{
		Branch: plumbing.ReferenceName(branchRefName),
		Force:  true,
	}
	err = wtree.Checkout(coOpts)
	if err == nil {
		return nil
	}

	l.Error("Unable to fetch reference as branch, tag or commit",
		zap.String("ref", src.Ref),
		zap.String("url", src.URL),
		zap.Error(err),
	)
	return CloneErr{
		Url: src.URL,
		Ref: src.Ref,
		Err: err,
	}
}
