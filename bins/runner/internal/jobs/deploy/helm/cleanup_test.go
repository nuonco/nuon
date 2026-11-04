package helm

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/errs"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
)

type failingCleanupArchive struct {
	ociarchive.Archive
	err error
}

func (a *failingCleanupArchive) Cleanup(context.Context) error { return a.err }

func TestCleanupReturnsArchiveError(t *testing.T) {
	want := errors.New("unable to remove archive")
	h := &handler{
		errRecorder: errs.NewRecorder(errs.Params{L: zap.NewNop(), LC: fxtest.NewLifecycle(t)}),
		state:       &handlerState{arch: &failingCleanupArchive{err: want}},
	}
	require.ErrorIs(t, h.Cleanup(context.Background(), nil, nil), want)
	require.NoError(t, h.Cleanup(context.Background(), nil, nil))
}

func TestCleanupRemovesArchive(t *testing.T) {
	arch := ociarchive.New()
	require.NoError(t, arch.Initialize(context.Background()))
	t.Cleanup(func() { _ = arch.Cleanup(context.Background()) })
	path := arch.TmpDir()
	require.DirExists(t, path)
	h := &handler{
		errRecorder: errs.NewRecorder(errs.Params{L: zap.NewNop(), LC: fxtest.NewLifecycle(t)}),
		state:       &handlerState{arch: arch},
	}
	require.NoError(t, h.Cleanup(context.Background(), nil, nil))
	require.NoDirExists(t, path)
	require.NoError(t, h.Cleanup(context.Background(), nil, nil))
}
