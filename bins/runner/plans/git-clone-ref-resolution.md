# Runner: resolve git refs as branch, tag, or commit hash

## Goals

1. Support pinning a `GitSource.Ref` to a **branch** (current behavior).
2. Support pinning to a **commit hash** (currently only works if reachable from the default branch).
3. Support pinning to a **tag** (not currently supported).

## Motivation

The charts repo (`nuonco/charts`) is moving to a tag-on-release model — each chart bump produces a git tag like
`ctl-api-v0.2.0`. The byoc component TOMLs need to point to those tags instead of `main` so deployments pin to a
specific released version. Today `internal/pkg/git/clone.go` only fetches `refs/heads/<ref>:refs/heads/<ref>`, so a
tag name passed as `src.Ref` fails resolution.

## Current behavior (`internal/pkg/git/clone.go`)

1. `git.PlainCloneContext` clones the default branch.
2. Try `wtree.Checkout(Hash: NewHash(src.Ref))` — succeeds only if `src.Ref` is a SHA reachable from the cloned
   default branch.
3. If that fails, fetch `refs/heads/<ref>:refs/heads/<ref>` from origin.
4. Try `wtree.Checkout(Branch: NewBranchReferenceName(src.Ref))`.

No tag fetch, no tag checkout. Hashes outside the default branch's history don't resolve.

## Proposed approach

Restructure the resolution into three stages, each a fallback for the previous:

1. **Direct hash** — if `src.Ref` parses as a non-zero hash, try checking it out against the clone we already have.
2. **Best-effort fetch** of both `refs/heads/<ref>` AND `refs/tags/<ref>`. Only one will match; the other is a cheap
   no-op. Errors here are non-fatal — we move on to checkout attempts.
3. **Checkout candidates** — try `NewBranchReferenceName`, then `NewTagReferenceName`. First success wins.

Branch takes precedence over tag on name collision, matching `git checkout` semantics.

### Sketch

```go
func (w *workspace) clone(ctx context.Context, rootDir string, src *plantypes.GitSource, l *zap.Logger) error {
    pWriter := zapwriter.New(l, zapcore.DebugLevel, "")

    l.Info("cloning repository", zap.String("url", src.URL))
    repo, err := git.PlainCloneContext(ctx, rootDir, false, &git.CloneOptions{
        URL:      src.URL,
        Progress: pWriter,
    })
    if err != nil {
        return cloneErr(src, err)
    }

    wtree, err := repo.Worktree()
    if err != nil {
        return cloneErr(src, err)
    }

    // 1) If the ref parses as a full commit hash, try a direct checkout.
    //    Works when the commit is reachable from the default branch we just cloned.
    if h := plumbing.NewHash(src.Ref); !h.IsZero() {
        if err := wtree.Checkout(&git.CheckoutOptions{Hash: h, Force: true}); err == nil {
            l.Info("checked out commit", zap.String("ref", src.Ref))
            return nil
        }
    }

    // 2) Fetch as branch AND as tag — only one will match, the other is a cheap no-op.
    remote, err := repo.Remote("origin")
    if err != nil {
        return cloneErr(src, err)
    }
    refspecs := []config.RefSpec{
        config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", src.Ref, src.Ref)),
        config.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", src.Ref, src.Ref)),
    }
    for _, rs := range refspecs {
        if err := remote.Fetch(&git.FetchOptions{RefSpecs: []config.RefSpec{rs}}); err != nil &&
            !errors.Is(err, git.NoErrAlreadyUpToDate) {
            // Non-fatal — the other refspec may still resolve.
            l.Debug("fetch attempt failed", zap.String("refspec", string(rs)), zap.Error(err))
        }
    }

    // 3) Try as branch, then as tag. go-git resolves through the tag object to the commit.
    candidates := []plumbing.ReferenceName{
        plumbing.NewBranchReferenceName(src.Ref),
        plumbing.NewTagReferenceName(src.Ref),
    }
    for _, name := range candidates {
        if err := wtree.Checkout(&git.CheckoutOptions{Branch: name, Force: true}); err == nil {
            l.Info("checked out ref", zap.String("ref", string(name)))
            return nil
        }
    }

    l.Error("unable to resolve ref as branch, tag, or commit",
        zap.String("ref", src.Ref),
        zap.String("url", src.URL),
    )
    return CloneErr{
        Url: src.URL,
        Ref: src.Ref,
        Err: fmt.Errorf("ref %q not found as commit, branch, or tag", src.Ref),
    }
}

func cloneErr(src *plantypes.GitSource, err error) error {
    return CloneErr{Url: src.URL, Ref: src.Ref, Err: err}
}
```

## Tradeoffs & things to call out

- **Branch beats tag on collision.** If someone names a branch and tag identically, the branch wins. Matches `git`
  itself; release-tag naming (`ctl-api-vX.Y.Z`) makes this unlikely in practice.
- **Lightweight vs annotated tags.** `Checkout{Branch: NewTagReferenceName(...)}` handles both — go-git follows the
  tag object to the commit.
- **No auth changes.** Same credentials/transport as today.
- **Two fetches per non-hash ref.** One extra round trip vs. today (we attempt both `refs/heads` and `refs/tags`).
  Cost is small; clarity is the win.

## Out of scope

- **Fetching arbitrary SHAs.** A hash that isn't reachable from the default branch still won't resolve, because we
  don't `fetch origin <sha>`. Doing so requires the server to honor `uploadpack.allowReachableSHA1InWant` /
  `allowAnySHA1InWant` (GitHub does for reachable commits). Not needed for the chart-release use case — if/when
  someone wants to pin to a non-tip SHA on a feature branch, add a third fetch attempt that does `git fetch origin
  <sha>` and retries the direct hash checkout.

## Validation

- Unit test in `internal/pkg/git/` exercising:
  - branch ref → checks out HEAD of that branch
  - tag ref (both lightweight and annotated) → checks out the tagged commit
  - commit hash reachable from default branch → checks out that commit
  - commit hash NOT reachable → returns `CloneErr` (documents current limitation)
  - unknown ref → returns `CloneErr`
- Integration check: deploy a byoc component with `branch = "ctl-api-v0.2.0"` against the tag created by the
  charts repo's release workflow.
