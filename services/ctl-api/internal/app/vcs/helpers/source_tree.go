package helpers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/sideband"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/storage/memory"

	githubpkg "github.com/nuonco/nuon/pkg/github/repo"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var (
	ErrSourceCommitNotFound   = errors.New("pinned commit not found")
	ErrSourceDirectoryMissing = errors.New("source directory is missing from pinned commit")
)

type SourceTree struct {
	s    storer.EncodedObjectStorer
	root *object.Tree
}

func (h *Helpers) GetSourceTree(ctx context.Context, orgID, repository, commit string, connection *app.VCSConnection) (*SourceTree, error) {
	if !sourceidentity.IsCommit(commit) {
		return nil, fmt.Errorf("source tree requires an immutable commit")
	}
	owner, repo, err := parseOwnerRepo(repository)
	if err != nil {
		return nil, err
	}
	scope := "public"
	if connection != nil {
		if connection.OrgID != orgID {
			return nil, fmt.Errorf("source connection does not belong to org")
		}
		scope = connection.ID
	}
	key := orgID + "\n" + scope + "\n" + owner + "/" + repo + "\n" + commit
	if h.sourceTrees != nil {
		if tree, ok := h.sourceTrees.Get(key); ok {
			return tree, nil
		}
	}
	result, err, _ := h.sourceFetches.Do(key, func() (any, error) {
		if h.sourceTrees != nil {
			if tree, ok := h.sourceTrees.Get(key); ok {
				return tree, nil
			}
		}
		token := ""
		var err error
		if connection != nil {
			token, err = h.CreateInstallationToken(ctx, connection, repo)
		} else {
			token, _, err = h.ResolvePublicRepoCloneToken(ctx, h.l, orgID, owner, repo)
		}
		if err != nil {
			return nil, err
		}
		url, auth := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo), transport.AuthMethod(nil)
		if token != "" {
			if url, auth, err = githubpkg.CloneURLAuth(githubpkg.RepoPath(owner, repo, token)); err != nil {
				return nil, err
			}
		}
		tree, err := fetchSourceTree(ctx, url, auth, commit, githubUploadPackCapabilities())
		if err != nil {
			return nil, err
		}
		if h.sourceTrees != nil {
			h.sourceTrees.Add(key, tree)
		}
		return tree, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*SourceTree), nil
}

func githubUploadPackCapabilities() *capability.List {
	caps := capability.NewList()
	for _, c := range []capability.Capability{capability.OFSDelta, capability.Sideband64k, capability.Shallow, capability.Filter} {
		_ = caps.Set(c)
	}
	return caps
}

func fetchSourceTree(ctx context.Context, url string, auth transport.AuthMethod, commit string, serverCaps *capability.List) (*SourceTree, error) {
	ep, err := transport.NewEndpoint(url)
	if err != nil {
		return nil, err
	}
	cli, err := client.NewClient(ep)
	if err != nil {
		return nil, err
	}
	sess, err := cli.NewUploadPackSession(ep, auth)
	if err != nil {
		return nil, err
	}
	defer sess.Close()
	if serverCaps == nil {
		adv, err := sess.AdvertisedReferencesContext(ctx)
		if errors.Is(err, transport.ErrEmptyRemoteRepository) {
			return nil, ErrSourceCommitNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("unable to read source repository: %w", err)
		}
		serverCaps = adv.Capabilities
	}

	hash := plumbing.NewHash(commit)
	req := packp.NewUploadPackRequestFromCapabilities(serverCaps)
	req.Wants = []plumbing.Hash{hash}
	if serverCaps.Supports(capability.Shallow) {
		if err := req.Capabilities.Set(capability.Shallow); err != nil {
			return nil, err
		}
		req.Depth = packp.DepthCommits(1)
	}
	if serverCaps.Supports(capability.Filter) {
		if err := req.Capabilities.Set(capability.Filter); err != nil {
			return nil, err
		}
		req.Filter = packp.FilterBlobNone()
	}
	resp, err := sess.UploadPack(ctx, req)
	if errors.Is(err, transport.ErrEmptyUploadPackRequest) || (err != nil && strings.Contains(err.Error(), "not our ref")) {
		return nil, ErrSourceCommitNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("unable to fetch pinned commit: %w", err)
	}
	defer resp.Close()

	var pack io.Reader = resp
	switch {
	case req.Capabilities.Supports(capability.Sideband64k):
		pack = sideband.NewDemuxer(sideband.Sideband64k, resp)
	case req.Capabilities.Supports(capability.Sideband):
		pack = sideband.NewDemuxer(sideband.Sideband, resp)
	}
	st := memory.NewStorage()
	if err := packfile.UpdateObjectStorage(st, pack); err != nil {
		return nil, fmt.Errorf("unable to read pinned commit: %w", err)
	}
	c, err := object.GetCommit(st, hash)
	if errors.Is(err, plumbing.ErrObjectNotFound) {
		return nil, ErrSourceCommitNotFound
	}
	if err != nil {
		return nil, err
	}
	root, err := c.Tree()
	if err != nil {
		return nil, fmt.Errorf("pinned commit has no source tree: %w", err)
	}
	return &SourceTree{s: st, root: root}, nil
}

func (t *SourceTree) Scope(root string) (string, error) {
	root = sourceidentity.CleanRoot(root)
	if root == "" {
		return "git:040000:" + t.root.Hash.String(), nil
	}
	entry, err := t.entry(root)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("git:%06o:%s", uint32(entry.Mode), entry.Hash), nil
}

func (t *SourceTree) HasSubmodule(root string) (bool, error) {
	tree := t.root
	if root = sourceidentity.CleanRoot(root); root != "" {
		entry, err := t.entry(root)
		if errors.Is(err, ErrSourceDirectoryMissing) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if entry.Mode == filemode.Submodule {
			return true, nil
		}
		if entry.Mode != filemode.Dir {
			return false, nil
		}
		if tree, err = object.GetTree(t.s, entry.Hash); err != nil {
			return false, err
		}
	}
	walker := object.NewTreeWalker(tree, true, nil)
	defer walker.Close()
	for {
		_, entry, err := walker.Next()
		if err == io.EOF {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if entry.Mode == filemode.Submodule {
			return true, nil
		}
	}
}

func (t *SourceTree) entry(root string) (*object.TreeEntry, error) {
	if path.IsAbs(root) || root == ".." || strings.HasPrefix(root, "../") {
		return nil, fmt.Errorf("source directory escapes repository")
	}
	tree := t.root
	parts := strings.Split(root, "/")
	for i, name := range parts {
		var entry *object.TreeEntry
		for j := range tree.Entries {
			if tree.Entries[j].Name == name {
				entry = &tree.Entries[j]
				break
			}
		}
		if entry == nil {
			return nil, fmt.Errorf("%w: %s", ErrSourceDirectoryMissing, root)
		}
		if i == len(parts)-1 || entry.Mode == filemode.Symlink {
			return entry, nil
		}
		if entry.Mode != filemode.Dir {
			return nil, fmt.Errorf("%w: %s", ErrSourceDirectoryMissing, root)
		}
		next, err := object.GetTree(t.s, entry.Hash)
		if err != nil {
			return nil, err
		}
		tree = next
	}
	return nil, fmt.Errorf("%w: %s", ErrSourceDirectoryMissing, root)
}
