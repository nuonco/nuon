package ociarchive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path"
	"slices"
	"sync"

	"github.com/nuonco/nuon/pkg/rootfs"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/afero"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
)

// contentStore is a private, job-lifetime CAS. References stay in memory: the
// archive never saves or reopens an OCI index, nor needs graph indexing or GC.
// Production callers transfer a root derived from workFS; memory is for tests.
type contentStore struct {
	fs      afero.Fs
	mu      sync.RWMutex
	closed  bool
	indexMu sync.Mutex
	refs    map[string]v1.Descriptor
}

var _ oras.Target = (*contentStore)(nil)

func newContentStore(fs afero.Fs) (*contentStore, error) {
	switch fs.(type) {
	case *rootfs.FS, *afero.MemMapFs:
	default:
		return nil, fmt.Errorf("unsupported archive content filesystem %T", fs)
	}

	s := &contentStore{fs: fs, refs: make(map[string]v1.Descriptor)}
	for _, name := range []string{"blobs", "blobs/sha256", "blobs/sha384", "blobs/sha512", "ingest"} {
		if err := fs.Mkdir(name, 0700); err != nil && !os.IsExist(err) {
			return nil, err
		}
		if err := s.directory(name); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func validateDescriptor(desc v1.Descriptor) error {
	if err := desc.Digest.Validate(); err != nil {
		return fmt.Errorf("%s: %w: %v", desc.Digest, errdef.ErrInvalidDigest, err)
	}
	if desc.Size < 0 {
		return content.ErrInvalidDescriptorSize
	}
	return nil
}

func (s *contentStore) directory(name string) error {
	info, _, err := s.fs.(afero.Lstater).LstatIfPossible(name)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("archive content directory %q is not a directory (symlinks are not allowed)", name)
	}
	return nil
}

func (s *contentStore) blobInfo(dgst digest.Digest) (string, os.FileInfo, error) {
	if err := dgst.Validate(); err != nil {
		return "", nil, fmt.Errorf("%s: %w", dgst, errdef.ErrInvalidDigest)
	}
	dir := path.Join("blobs", dgst.Algorithm().String())
	for _, name := range []string{"blobs", dir} {
		if err := s.directory(name); err != nil {
			return "", nil, err
		}
	}
	name := path.Join(dir, dgst.Encoded())
	info, _, err := s.fs.(afero.Lstater).LstatIfPossible(name)
	if err == nil && !info.Mode().IsRegular() {
		err = fmt.Errorf("archive content %q is not a regular file (symlinks are not allowed)", name)
	}
	return name, info, err
}

func (s *contentStore) exists(desc v1.Descriptor) (bool, error) {
	_, info, err := s.blobInfo(desc.Digest)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Size() != desc.Size {
		return false, content.ErrInvalidDescriptorSize
	}
	return true, nil
}

func (s *contentStore) ready(ctx context.Context, desc v1.Descriptor) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.closed {
		return fmt.Errorf("archive content store closed")
	}
	return validateDescriptor(desc)
}

func (s *contentStore) Exists(ctx context.Context, desc v1.Descriptor) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := s.ready(ctx, desc); err != nil {
		return false, err
	}
	return s.exists(desc)
}

func (s *contentStore) Fetch(ctx context.Context, desc v1.Descriptor) (io.ReadCloser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := s.ready(ctx, desc); err != nil {
		return nil, err
	}
	name, _, err := s.blobInfo(desc.Digest)
	if os.IsNotExist(err) {
		return nil, errdef.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	f, err := s.fs.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err == nil && (!info.Mode().IsRegular() || info.Size() != desc.Size) {
		err = content.ErrInvalidDescriptorSize
	}
	if err != nil {
		return nil, errors.Join(err, f.Close())
	}
	return &verifiedBlob{ctx: ctx, file: f, reader: content.NewVerifyReader(f, desc)}, nil
}

// Verify on EOF, including direct extraction reads (not just FetchAll callers).
// Closing a partial read deliberately does not drain a potentially large blob.
type verifiedBlob struct {
	ctx    context.Context
	file   io.Closer
	reader *content.VerifyReader
}

func (b *verifiedBlob) Read(p []byte) (int, error) {
	if err := b.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := b.reader.Read(p)
	if err == io.EOF {
		if verifyErr := b.reader.Verify(); verifyErr != nil {
			return n, verifyErr
		}
	}
	return n, err
}

func (b *verifiedBlob) Close() error { return b.file.Close() }

func (s *contentStore) Push(ctx context.Context, desc v1.Descriptor, reader io.Reader) (retErr error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := s.ready(ctx, desc); err != nil {
		return err
	}
	if exists, err := s.exists(desc); err != nil {
		return err
	} else if exists {
		return errdef.ErrAlreadyExists
	}
	if err := s.directory("ingest"); err != nil {
		return err
	}
	f, err := afero.TempFile(s.fs, "ingest", "blob-")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer func() {
		if f != nil {
			retErr = errors.Join(retErr, f.Close())
		}
		if err := s.fs.Remove(tmp); err != nil && !os.IsNotExist(err) {
			retErr = errors.Join(retErr, err)
		}
	}()
	vr := content.NewVerifyReader(&contextReader{ctx: ctx, reader: reader}, desc)
	if _, err := io.Copy(f, vr); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := vr.Verify(); err != nil {
		return err
	}
	// Only manifests are decoded. Blob bytes are never buffered in memory.
	children, err := content.Successors(ctx, content.FetcherFunc(func(context.Context, v1.Descriptor) (io.ReadCloser, error) {
		return io.NopCloser(io.NewSectionReader(f, 0, desc.Size)), nil
	}), desc)
	if err != nil {
		return err
	}
	for _, child := range children {
		if err := validateDescriptor(child); err != nil {
			return err
		}
	}
	closeErr := f.Close()
	f = nil
	if closeErr != nil {
		return closeErr
	}
	// Serializing only publication permits parallel streaming and ensures that
	// duplicate pushes never replace a blob another copy has already published.
	s.indexMu.Lock()
	defer s.indexMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	name, info, err := s.blobInfo(desc.Digest)
	if err == nil {
		if info.Size() != desc.Size {
			return content.ErrInvalidDescriptorSize
		}
		return errdef.ErrAlreadyExists
	}
	if !os.IsNotExist(err) || name == "" {
		return err
	}
	if err := s.fs.Rename(tmp, name); err != nil {
		return err
	}
	if isManifest(desc) {
		s.refs[desc.Digest.String()] = plainDescriptor(desc)
	}
	return nil
}

func isManifest(desc v1.Descriptor) bool {
	switch desc.MediaType {
	case v1.MediaTypeImageManifest, v1.MediaTypeImageIndex,
		"application/vnd.oci.artifact.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.docker.distribution.manifest.list.v2+json":
		return true
	}
	return false
}

func plainDescriptor(desc v1.Descriptor) v1.Descriptor {
	return v1.Descriptor{MediaType: desc.MediaType, Digest: desc.Digest, Size: desc.Size}
}

func cloneDescriptor(desc v1.Descriptor) v1.Descriptor {
	desc.Annotations = maps.Clone(desc.Annotations)
	desc.URLs = slices.Clone(desc.URLs)
	desc.Data = slices.Clone(desc.Data)
	if desc.Platform != nil {
		platform := *desc.Platform
		platform.OSFeatures = slices.Clone(platform.OSFeatures)
		desc.Platform = &platform
	}
	return desc
}

func (s *contentStore) Tag(ctx context.Context, desc v1.Descriptor, reference string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := s.ready(ctx, desc); err != nil {
		return err
	}
	if reference == "" {
		return errdef.ErrMissingReference
	}
	if exists, err := s.exists(desc); err != nil {
		return err
	} else if !exists {
		return errdef.ErrNotFound
	}
	s.indexMu.Lock()
	defer s.indexMu.Unlock()
	s.refs[desc.Digest.String()] = plainDescriptor(desc)
	s.refs[reference] = cloneDescriptor(desc)
	return nil
}

func (s *contentStore) Resolve(ctx context.Context, reference string) (v1.Descriptor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return v1.Descriptor{}, err
	}
	if s.closed {
		return v1.Descriptor{}, fmt.Errorf("archive content store closed")
	}
	if reference == "" {
		return v1.Descriptor{}, errdef.ErrMissingReference
	}
	s.indexMu.Lock()
	desc, ok := s.refs[reference]
	s.indexMu.Unlock()
	if ok {
		if reference == desc.Digest.String() {
			return plainDescriptor(desc), nil
		}
		return cloneDescriptor(desc), nil
	}
	_, info, err := s.blobInfo(digest.Digest(reference))
	if os.IsNotExist(err) || errors.Is(err, errdef.ErrInvalidDigest) {
		return v1.Descriptor{}, errdef.ErrNotFound
	}
	if err != nil {
		return v1.Descriptor{}, err
	}
	return v1.Descriptor{MediaType: "application/octet-stream", Digest: digest.Digest(reference), Size: info.Size()}, nil
}

func (s *contentStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	entries, err := afero.ReadDir(s.fs, ".")
	errs := []error{err}
	for _, entry := range entries {
		errs = append(errs, s.fs.RemoveAll(entry.Name()))
	}
	if closer, ok := s.fs.(io.Closer); ok {
		errs = append(errs, closer.Close())
	}
	return errors.Join(errs...)
}
