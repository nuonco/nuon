package ociarchive

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/pkg/rootfs"
)

func (a *archive) Initialize(ctx context.Context) (retErr error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	if a.workFS != nil || a.store != nil {
		return fmt.Errorf("archive already initialized")
	}

	tmpDir, err := a.createTmpDir()
	if err != nil {
		return err
	}
	a.tmpDir = tmpDir
	defer func() {
		if retErr != nil {
			retErr = errors.Join(retErr, a.Cleanup(ctx))
		}
	}()

	a.workFS, err = rootfs.Open(tmpDir)
	if err != nil {
		return fmt.Errorf("unable to pin archive directory: %w", err)
	}
	if err := a.workFS.Mkdir("store", 0700); err != nil {
		return fmt.Errorf("unable to create output directory: %w", err)
	}
	output, err := a.workFS.Sub("store")
	if err != nil {
		return fmt.Errorf("unable to pin output directory: %w", err)
	}
	a.store = &archiveStore{output: output}
	a.basePath = output.Root()
	if err := a.workFS.Mkdir("layers", 0700); err != nil {
		return fmt.Errorf("unable to create layer directory: %w", err)
	}

	if err := a.workFS.Mkdir("content", 0700); err != nil {
		return fmt.Errorf("unable to create content directory: %w", err)
	}
	contentFS, err := a.workFS.Sub("content")
	if err != nil {
		return fmt.Errorf("unable to pin content directory: %w", err)
	}
	store, err := newContentStore(contentFS)
	if err != nil {
		return errors.Join(fmt.Errorf("unable to create content store: %w", err), contentFS.Close())
	}
	a.store.contentStore = store
	return nil
}
