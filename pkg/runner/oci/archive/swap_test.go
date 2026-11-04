package ociarchive

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractionSymlinkSwapStressAndControl(t *testing.T) {
	ctx := context.Background()
	entries := []tarEntry{regularEntry("swap/canary", "must stay in archive")}
	for i := 0; i < 80; i++ {
		entries = append(entries, regularEntry(fmt.Sprintf("swap/file-%03d", i), "contents"))
	}
	source := localFixture(t, fixtureLayer{title: ".", unpack: true, data: fixtureTar(t, entries...)})
	control := newTestArchive(t, ctx)
	require.NoError(t, copyFixture(ctx, source, control))
	for attempt := 0; attempt < 12; attempt++ {
		dst := newTestArchive(t, ctx)
		outside := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(outside, "canary"), []byte("neighbor"), 0640))
		swap, held := filepath.Join(dst.BasePath(), "swap"), filepath.Join(dst.BasePath(), "held")
		require.NoError(t, os.Mkdir(swap, 0755))
		stop, started, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
		var swaps atomic.Int64
		go func() {
			for {
				if err := os.Rename(swap, held); err != nil {
					done <- err
					return
				}
				if err := os.Symlink(outside, swap); err != nil {
					if os.IsExist(err) && swaps.Load() > 0 {
						done <- nil
						return
					}
					done <- err
					return
				}
				if swaps.Add(1) == 1 {
					close(started)
				}
				runtime.Gosched()
				if err := os.Remove(swap); err != nil {
					done <- err
					return
				}
				if err := os.Rename(held, swap); err != nil {
					// Extraction may create a safe directory in the gap.
					if errors.Is(err, syscall.ENOTEMPTY) || errors.Is(err, os.ErrExist) {
						done <- nil
						return
					}
					done <- err
					return
				}
				select {
				case <-stop:
					done <- nil
					return
				default:
					runtime.Gosched()
				}
			}
		}()
		<-started
		// A confined success or a detected race are both valid outcomes.
		_ = copyFixture(ctx, source, dst)
		close(stop)
		require.NoError(t, <-done)
		require.Positive(t, swaps.Load())
		contents, err := os.ReadFile(filepath.Join(outside, "canary"))
		require.NoError(t, err)
		require.Equal(t, "neighbor", string(contents))
		info, err := os.Stat(filepath.Join(outside, "canary"))
		require.NoError(t, err)
		require.EqualValues(t, 0640, info.Mode().Perm())
		neighbors, err := os.ReadDir(outside)
		require.NoError(t, err)
		require.Len(t, neighbors, 1, "no files may be created outside the pinned output")
	}
}
