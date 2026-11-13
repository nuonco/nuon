package helpers

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"sync"
	"time"

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
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var (
	ErrSourceCommitNotFound   = errors.New("pinned commit not found")
	ErrSourceDirectoryMissing = errors.New("source directory is missing from pinned commit")
	ErrSourceTreeTooLarge     = errors.New("source tree exceeds the size limit")
)

const (
	maxSourceTreeBytes         = 128 << 20
	maxCachedSourceTreeBytes   = 256 << 20
	maxConcurrentSourceFetches = 4
	sourceFetchTimeout         = 3 * time.Minute
)

type limitedStorage struct {
	*memory.Storage
	remaining int64
}

func (s *limitedStorage) SetEncodedObject(o plumbing.EncodedObject) (plumbing.Hash, error) {
	if s.remaining -= o.Size(); s.remaining < 0 {
		return plumbing.ZeroHash, ErrSourceTreeTooLarge
	}
	return s.Storage.SetEncodedObject(o)
}

type SourceTree struct {
	s     storer.EncodedObjectStorer
	root  *object.Tree
	size  int64
	mu    sync.Mutex
	trees map[plumbing.Hash]*object.Tree
}

func (t *SourceTree) tree(hash plumbing.Hash) (*object.Tree, error) {
	t.mu.Lock()
	tree, ok := t.trees[hash]
	t.mu.Unlock()
	if ok {
		return tree, nil
	}
	tree, err := object.GetTree(t.s, hash)
	if err != nil {
		return nil, err
	}
	t.mu.Lock()
	if t.trees == nil {
		t.trees = make(map[plumbing.Hash]*object.Tree)
	}
	t.trees[hash] = tree
	t.mu.Unlock()
	return tree, nil
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
	if h.sourceTreesTooLarge != nil {
		if _, ok := h.sourceTreesTooLarge.Get(key); ok {
			return nil, ErrSourceTreeTooLarge
		}
	}
	if h.sourceTrees != nil {
		if tree, ok := h.sourceTrees.Get(key); ok {
			return tree, nil
		}
	}
	fetches := h.sourceFetches.DoChan(key, func() (tree any, err error) {
		defer func() {
			if r := recover(); r != nil {
				tree, err = nil, fmt.Errorf("source tree fetch panicked: %v", r)
			}
		}()
		return h.loadSourceTree(ctx, key, orgID, owner, repo, commit, connection)
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-fetches:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.(*SourceTree), nil
	}
}

func (h *Helpers) loadSourceTree(ctx context.Context, key, orgID, owner, repo, commit string, connection *app.VCSConnection) (*SourceTree, error) {
	if h.sourceTrees != nil {
		if tree, ok := h.sourceTrees.Get(key); ok {
			return tree, nil
		}
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), sourceFetchTimeout)
	defer cancel()
	if h.sourceFetchSlots != nil {
		select {
		case h.sourceFetchSlots <- struct{}{}:
			defer func() { <-h.sourceFetchSlots }()
		case <-ctx.Done():
			return nil, ctx.Err()
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
	var auth transport.AuthMethod
	if token != "" {
		auth = &githttp.BasicAuth{Username: "x-access-token", Password: token}
	}
	tree, err := fetchSourceTree(ctx, fmt.Sprintf("https://github.com/%s/%s.git", owner, repo), auth, commit, githubUploadPackCapabilities(), maxSourceTreeBytes)
	if errors.Is(err, ErrSourceTreeTooLarge) && h.sourceTreesTooLarge != nil {
		h.sourceTreesTooLarge.Add(key, struct{}{})
	}
	if err != nil {
		return nil, err
	}
	h.cacheSourceTree(key, tree)
	return tree, nil
}

func (h *Helpers) cacheSourceTree(key string, tree *SourceTree) {
	if h.sourceTrees == nil {
		return
	}
	h.sourceTrees.Add(key, tree)
	for {
		var total int64
		for _, cached := range h.sourceTrees.Values() {
			total += cached.size
		}
		if total <= maxCachedSourceTreeBytes {
			return
		}
		if _, _, ok := h.sourceTrees.RemoveOldest(); !ok {
			return
		}
	}
}

func githubUploadPackCapabilities() *capability.List {
	caps := capability.NewList()
	for _, c := range []capability.Capability{capability.OFSDelta, capability.Sideband64k, capability.Shallow, capability.Filter} {
		_ = caps.Set(c)
	}
	return caps
}

func fetchSourceTree(ctx context.Context, url string, auth transport.AuthMethod, commit string, serverCaps *capability.List, maxBytes int64) (*SourceTree, error) {
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
	chunks, err := readPack(pack, maxBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to read pinned commit: %w", err)
	}
	readers := make([]io.Reader, len(chunks))
	for i, chunk := range chunks {
		readers[i] = bytes.NewReader(chunk)
	}
	if err := checkPackSize(io.MultiReader(readers...), maxBytes); err != nil {
		return nil, fmt.Errorf("unable to read pinned commit: %w", err)
	}
	parse := &releasingReader{chunks: chunks}
	chunks, readers = nil, nil
	st := &limitedStorage{Storage: memory.NewStorage(), remaining: maxBytes}
	if err := packfile.UpdateObjectStorage(st, parse); err != nil {
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
	return &SourceTree{s: st.Storage, root: root, size: maxBytes - st.remaining}, nil
}

func readPack(pack io.Reader, maxBytes int64) ([][]byte, error) {
	var chunks [][]byte
	var total int64
	for {
		chunk := make([]byte, 1<<20)
		n := 0
		var err error
		for n < len(chunk) && err == nil {
			var read int
			read, err = pack.Read(chunk[n:])
			n += read
		}
		chunks = append(chunks, chunk[:n])
		if total += int64(n); total > maxBytes {
			return nil, ErrSourceTreeTooLarge
		}
		if errors.Is(err, io.EOF) {
			return chunks, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

type releasingReader struct {
	chunks [][]byte
}

func (r *releasingReader) Read(p []byte) (int, error) {
	for len(r.chunks) > 0 && len(r.chunks[0]) == 0 {
		r.chunks[0] = nil
		r.chunks = r.chunks[1:]
	}
	if len(r.chunks) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.chunks[0])
	r.chunks[0] = r.chunks[0][n:]
	return n, nil
}

func checkPackSize(pack io.Reader, maxBytes int64) error {
	s := packfile.NewScanner(pack)
	_, count, err := s.Header()
	if err != nil {
		return err
	}
	var total int64
	for range count {
		h, err := s.NextObjectHeader()
		if err != nil {
			return err
		}
		if h.Length > maxBytes-total {
			return ErrSourceTreeTooLarge
		}
		size := h.Length
		if h.Type.IsDelta() {
			var header deltaHeader
			if _, _, err := s.NextObject(&header); err != nil {
				return err
			}
			_, n := binary.Uvarint(header)
			if n <= 0 {
				return packfile.ErrMalformedPackFile
			}
			target, m := binary.Uvarint(header[n:])
			if m <= 0 {
				return packfile.ErrMalformedPackFile
			}
			if target > uint64(maxBytes) {
				return ErrSourceTreeTooLarge
			}
			size += int64(target)
		}
		if total += size; total > maxBytes {
			return ErrSourceTreeTooLarge
		}
	}
	return nil
}

type deltaHeader []byte

func (h *deltaHeader) Write(p []byte) (int, error) {
	if n := 2*binary.MaxVarintLen64 - len(*h); n > 0 {
		*h = append(*h, p[:min(n, len(p))]...)
	}
	return len(p), nil
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

const (
	UnsupportedSubmodule    = "git_submodules_not_materialized"
	UnsupportedNonCanonical = "git_noncanonical_tree"
)

func (t *SourceTree) Unsupported(ctx context.Context, root string) (string, error) {
	tree := t.root
	if root = sourceidentity.CleanRoot(root); root != "" {
		entry, err := t.entry(root)
		if errors.Is(err, ErrSourceDirectoryMissing) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		if entry.Mode != filemode.Dir {
			return unsupportedMode(entry.Mode), nil
		}
		if tree, err = t.tree(entry.Hash); err != nil {
			return "", err
		}
	}
	seen := map[plumbing.Hash]bool{tree.Hash: true}
	var pending []plumbing.Hash
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		canonical, err := t.canonical(tree)
		if err != nil {
			return "", err
		}
		if !canonical {
			return UnsupportedNonCanonical, nil
		}
		for _, entry := range tree.Entries {
			if reason := unsupportedMode(entry.Mode); reason != "" {
				return reason, nil
			}
			if entry.Mode != filemode.Dir || seen[entry.Hash] {
				continue
			}
			seen[entry.Hash] = true
			pending = append(pending, entry.Hash)
		}
		if len(pending) == 0 {
			return "", nil
		}
		hash := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if tree, err = object.GetTree(t.s, hash); err != nil {
			return "", err
		}
	}
}

func unsupportedMode(mode filemode.FileMode) string {
	if mode == filemode.Submodule {
		return UnsupportedSubmodule
	}
	return ""
}

func (t *SourceTree) canonical(tree *object.Tree) (bool, error) {
	obj := t.s.NewEncodedObject()
	if err := tree.Encode(obj); err != nil {
		return false, err
	}
	return obj.Hash() == tree.Hash, nil
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
		next, err := t.tree(entry.Hash)
		if err != nil {
			return nil, err
		}
		tree = next
	}
	return nil, fmt.Errorf("%w: %s", ErrSourceDirectoryMissing, root)
}
