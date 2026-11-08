package workspace

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/spf13/afero"

	// errs "github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	github "github.com/nuonco/nuon/pkg/github/repo"
	"github.com/nuonco/nuon/pkg/rootfs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/op"
	"github.com/nuonco/nuon/pkg/zapwriter"
)

// regex to match full git commit hash
// regex to match full git commit hash
var commitHashRegex = regexp.MustCompile(`\b[0-9a-f]{5,40}\b`)

// fetch-by-SHA refspecs only accept a full 40-character hash.
var fullCommitHashRegex = regexp.MustCompile(`^[0-9a-f]{40}$`)

const shallowCheckoutBranch = "nuon-checkout"

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
	return w.Src.URL != emptyGithubRepoURL
}

func (w *workspace) clone(ctx context.Context) (retErr error) {
	opCtx, end := op.Tool(ctx, "git", "clone")
	ctx = opCtx
	defer func() { end(retErr) }()

	l := w.L
	if opL, err := pkgctx.Logger(opCtx); err == nil && opL != nil {
		l = opL
	}

	pWriter := zapwriter.New(l, zapcore.DebugLevel, "")

	entries, err := afero.ReadDir(w.fs, ".")
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("clone requires an empty workspace")
	}
	defer func() {
		if retErr != nil {
			retErr = errors.Join(retErr, w.cleanupDir())
		}
	}()
	cloneURL, auth, err := github.CloneURLAuth(w.Src.URL)
	if err != nil {
		return err
	}

	// go-git's in-process transport for local repositories rejects shallow fetches.
	if isLocalRepo(cloneURL) {
		return w.fullClone(ctx, l, cloneURL, auth, pWriter)
	}

	l.Info("shallow cloning repository", zap.String("url", cloneURL), zap.String("ref", w.Src.Ref))
	shallowErr := w.shallowClone(ctx, l, cloneURL, auth, pWriter)
	if shallowErr == nil {
		return nil
	}
	if !errors.Is(shallowErr, git.ErrExactSHA1NotSupported) {
		return CloneErr{
			Url: cloneURL,
			Ref: w.Src.Ref,
			Err: shallowErr,
		}
	}
	l.Warn("server does not support fetching by commit hash, falling back to full clone",
		zap.String("url", cloneURL),
		zap.String("ref", w.Src.Ref),
	)
	if err := w.emptyFS(); err != nil {
		return errors.Join(shallowErr, err)
	}

	return w.fullClone(ctx, l, cloneURL, auth, pWriter)
}

func isLocalRepo(cloneURL string) bool {
	ep, err := transport.NewEndpoint(cloneURL)
	return err == nil && ep.Protocol == "file"
}

type shallowCandidate struct {
	refSpec  config.RefSpec
	checkout git.CheckoutOptions
}

func (w *workspace) shallowCandidates() []shallowCandidate {
	ref := w.Src.Ref
	if fullCommitHashRegex.MatchString(ref) {
		dst := plumbing.NewBranchReferenceName(shallowCheckoutBranch)
		return []shallowCandidate{{
			refSpec:  config.RefSpec(fmt.Sprintf("%s:%s", ref, dst)),
			checkout: git.CheckoutOptions{Hash: plumbing.NewHash(ref), Force: true},
		}}
	}

	branch := plumbing.NewBranchReferenceName(ref)
	tag := plumbing.NewTagReferenceName(ref)
	return []shallowCandidate{
		{
			refSpec:  config.RefSpec(fmt.Sprintf("+%s:%s", branch, branch)),
			checkout: git.CheckoutOptions{Branch: branch, Force: true},
		},
		{
			refSpec:  config.RefSpec(fmt.Sprintf("+%s:%s", tag, tag)),
			checkout: git.CheckoutOptions{Branch: tag, Force: true},
		},
	}
}

func (w *workspace) shallowClone(ctx context.Context, l *zap.Logger, cloneURL string, auth transport.AuthMethod, progress io.Writer) (retErr error) {
	worktree := rootfs.NewBilly(w.fs)
	defer func() { retErr = errors.Join(retErr, worktree.Close()) }()
	if err := worktree.MkdirAll(".git", 0o700); err != nil {
		return err
	}
	dotgit, err := worktree.Chroot(".git")
	if err != nil {
		return err
	}
	storage := filesystem.NewStorage(dotgit, cache.NewObjectLRUDefault())
	defer func() { retErr = errors.Join(retErr, storage.Close()) }()

	repo, err := git.Init(storage, worktree)
	if err != nil {
		return err
	}
	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{cloneURL},
	})
	if err != nil {
		return err
	}
	wtree, err := repo.Worktree()
	if err != nil {
		return err
	}

	var errs []error
	for _, c := range w.shallowCandidates() {
		err := remote.FetchContext(ctx, &git.FetchOptions{
			Auth:     auth,
			RefSpecs: []config.RefSpec{c.refSpec},
			Depth:    1,
			Tags:     git.NoTags,
			Progress: progress,
		})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			errs = append(errs, fmt.Errorf("fetch %s: %w", c.refSpec, err))
			continue
		}
		if err := wtree.Checkout(&c.checkout); err != nil {
			errs = append(errs, fmt.Errorf("checkout %s: %w", c.refSpec, err))
			continue
		}
		l.Info("shallow clone complete",
			zap.String("url", cloneURL),
			zap.String("ref", w.Src.Ref),
			zap.String("ref_spec", c.refSpec.String()),
		)
		return nil
	}

	return errors.Join(errs...)
}

func (w *workspace) emptyFS() error {
	entries, err := afero.ReadDir(w.fs, ".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := w.fs.RemoveAll(entry.Name()); err != nil {
			return err
		}
	}
	return nil
}

func (w *workspace) fullClone(ctx context.Context, l *zap.Logger, cloneURL string, auth transport.AuthMethod, pWriter io.Writer) (retErr error) {
	var err error
	l.Info("cloning repository", zap.String("url", cloneURL))
	worktree := rootfs.NewBilly(w.fs)
	defer func() { retErr = errors.Join(retErr, worktree.Close()) }()
	if err := worktree.MkdirAll(".git", 0o700); err != nil {
		return err
	}
	dotgit, err := worktree.Chroot(".git")
	if err != nil {
		return err
	}
	storage := filesystem.NewStorage(dotgit, cache.NewObjectLRUDefault())
	defer func() { retErr = errors.Join(retErr, storage.Close()) }()
	repo, err := git.CloneContext(ctx, storage, worktree, &git.CloneOptions{
		URL:        cloneURL,
		Auth:       auth,
		Progress:   pWriter,
		NoCheckout: true,
	})
	if err != nil {
		return CloneErr{
			Url: cloneURL,
			Ref: w.Src.Ref,
			Err: err,
		}
	}

	l.Info("fetching working tree",
		zap.String("url", cloneURL),
		zap.String("ref", w.Src.Ref),
	)
	wtree, err := repo.Worktree()
	if err != nil {
		return CloneErr{
			Url: cloneURL,
			Ref: w.Src.Ref,
			Err: err,
		}
	}

	// hoist this var, like a savage
	coOpts := &git.CheckoutOptions{}

	// first, if it looks like a 40 char regex, attempt to check out as a reference w/ the hash
	if IsCommitHash(w.Src.Ref) {
		hash := plumbing.NewHash(w.Src.Ref)
		l.Info("checking out as reference",
			zap.String("url", cloneURL),
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
			l.Error("failed to check out as reference",
				zap.String("url", cloneURL),
				zap.String("ref", w.Src.Ref),
				zap.String("hash", hash.String()),
				zap.String("error", err.Error()),
			)
		}
	}

	// fetch remote origin
	l.Debug("fetching remote origin",
		zap.String("url", cloneURL),
		zap.String("ref", w.Src.Ref),
	)
	remote, err := repo.Remote("origin")
	if err != nil {
		return CloneErr{
			Url: cloneURL,
			Ref: w.Src.Ref,
			Err: err,
		}
	}
	refSpecStr := fmt.Sprintf("refs/heads/%s:refs/heads/%s", w.Src.Ref, w.Src.Ref)
	l.Info("fetching remote origin",
		zap.String("url", cloneURL),
		zap.String("ref", w.Src.Ref),
		zap.String("ref_spec_str", refSpecStr),
	)
	err = remote.Fetch(&git.FetchOptions{
		Auth:     auth,
		RefSpecs: []config.RefSpec{config.RefSpec(refSpecStr)},
	})
	if err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			l.Info("failed to fetch remote origin",
				zap.String("url", cloneURL),
				zap.String("ref", w.Src.Ref),
				zap.String("ref_spec_str", refSpecStr),
				zap.String("error", err.Error()),
			)
			// return CloneErr{
			// 	Url: w.Src.URL,
			// 	Ref: w.Src.Ref,
			// 	Err: errs.Wrap(err, "error fetching origin"),
			// }
		}
	}

	// second, attempt to check out as a branch
	branchRefName := plumbing.NewBranchReferenceName(w.Src.Ref)
	branch := plumbing.ReferenceName(branchRefName)
	l.Info("checking out branch",
		zap.String("url", cloneURL),
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
		l.Error("failed to check out as branch",
			zap.String("url", cloneURL),
			zap.String("ref", w.Src.Ref),
			zap.String("branch_ref_name", branchRefName.String()),
			zap.String("branch", branch.String()),
			zap.String("error", err.Error()),
		)
	}

	// third, attempt to check out as a tag
	tagRefName := plumbing.NewTagReferenceName(w.Src.Ref)
	l.Info("checking out as a tag",
		zap.String("url", cloneURL),
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
		l.Error("failed to check out as a tag",
			zap.String("url", cloneURL),
			zap.String("ref", w.Src.Ref),
			zap.String("tag_ref_name", tagRefName.String()),
			zap.String("error", err.Error()),
		)
	}

	return CloneErr{
		Url: cloneURL,
		Ref: w.Src.Ref,
		Err: err,
	}
}
