package terraform

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
)

func TestCleanupPartialInitialization(t *testing.T) {
	for _, phase := range []string{"before state", "before archive", "archive only"} {
		t.Run(phase, func(t *testing.T) {
			ctx := context.Background()
			h := &handler{}
			var archivePath string
			if phase != "before state" {
				h.state = &handlerState{}
			}
			if phase == "archive only" {
				arch := ociarchive.New()
				require.NoError(t, arch.Initialize(ctx))
				t.Cleanup(func() { require.NoError(t, arch.Cleanup(ctx)) })
				archivePath = arch.BasePath()
				h.state.arch = arch
			}
			require.NoError(t, h.Cleanup(ctx, nil, nil))
			require.NoError(t, h.Cleanup(ctx, nil, nil))
			if archivePath != "" {
				_, err := os.Stat(archivePath)
				require.True(t, os.IsNotExist(err), "archive must be removed: %v", err)
			}
		})
	}
}
