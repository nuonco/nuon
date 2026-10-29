package ociarchive

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/nuonco/nuon/pkg/rootfs"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
)

// Blobs never share a directory with extracted names. Tag is the publication
// boundary: all blobs have been verified by the CAS before any are extracted,
// and manifest order (including duplicate-content layers) survives parallel copy.
type archiveStore struct {
	*contentStore
	output       *rootfs.FS
	mu           sync.Mutex
	closed       bool
	failed       error
	materialized digest.Digest
	directories  map[string]directoryMode
}

var _ oras.Target = (*archiveStore)(nil)

func (s *archiveStore) Tag(ctx context.Context, desc v1.Descriptor, reference string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failed != nil {
		return s.failed
	}
	if s.closed {
		return fmt.Errorf("archive store closed")
	}
	if err := validateDescriptor(desc); err != nil {
		return err
	}
	if reference == "" {
		return errdef.ErrMissingReference
	}
	if s.materialized != "" && s.materialized != desc.Digest {
		return fmt.Errorf("archive output already materialized")
	}
	if s.materialized != desc.Digest {
		s.directories = make(map[string]directoryMode)
		extraction := &extraction{fs: s.output, directories: s.directories}
		if err := s.extractGraph(ctx, desc, extraction, 0); err != nil {
			return s.reject(err)
		}
		if err := extraction.validateSymlinks(ctx); err != nil {
			return s.reject(err)
		}
		if err := extraction.restoreDirectoryModes(); err != nil {
			return s.reject(err)
		}
		s.materialized = desc.Digest
	}
	return s.contentStore.Tag(ctx, desc, reference)
}

func (s *archiveStore) extractGraph(ctx context.Context, desc v1.Descriptor, extraction *extraction, depth int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateDescriptor(desc); err != nil {
		return err
	}
	if depth > 128 {
		return fmt.Errorf("archive manifest nesting too deep")
	}
	if name := desc.Annotations[v1.AnnotationTitle]; name != "" {
		blob, err := s.contentStore.Fetch(ctx, desc)
		if err != nil {
			return err
		}
		err = extraction.layer(ctx, desc, blob)
		closeErr := blob.Close()
		if err != nil {
			return fmt.Errorf("layer %q: %w", name, err)
		}
		return closeErr
	}
	children, err := content.Successors(ctx, s.contentStore, desc)
	if err != nil {
		return err
	}
	for _, child := range children {
		if err := s.extractGraph(ctx, child, extraction, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func (s *archiveStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.closeOutput()
	if s.contentStore != nil {
		err = errors.Join(err, s.contentStore.Close())
	}
	return err
}

func (s *archiveStore) checkOpen() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("archive store closed")
	}
	return nil
}

func (s *archiveStore) reject(err error) error {
	s.failed = errors.Join(fmt.Errorf("unable to materialize archive: %w", err), s.closeOutput())
	return s.failed
}

func (s *archiveStore) closeOutput() error {
	if s.closed {
		return nil
	}
	s.closed = true
	// Read-only archive directories must remain removable during job cleanup.
	names := make([]string, 0, len(s.directories))
	for name := range s.directories {
		names = append(names, name)
	}
	sort.Strings(names)
	var errs []error
	for _, name := range names {
		if retained := s.directories[name].cleanup; retained != nil {
			errs = append(errs, setFileMode(retained, 0700), retained.Close())
			continue
		}
		dir, err := s.output.Sub(name)
		if err != nil {
			if !os.IsNotExist(err) {
				errs = append(errs, err)
			}
			continue
		}
		errs = append(errs, dir.Chmod(".", 0700), dir.Close())
	}
	errs = append(errs, s.output.Close())
	return errors.Join(errs...)
}
