package terraform

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/pkg/sourcefs"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func TestValidateSourceFilesMemory(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll("/source", 0755))
	source, err := sourcefs.New(fs, "/source")
	require.NoError(t, err)
	h := &handler{}
	files := []ociarchive.FileRef{{AbsPath: "/source/main.tf", RelPath: "main.tf"}}
	require.NoError(t, afero.WriteFile(fs, files[0].AbsPath, []byte("terraform { backend s3 {} }"), 0644))
	require.NoError(t, h.validateSourceFiles(context.Background(), source, files))
	require.NoError(t, afero.WriteFile(fs, files[0].AbsPath, []byte("terraform {}"), 0644))
	require.ErrorContains(t, h.validateSourceFiles(context.Background(), source, files), "no backend s3 {} block found")
	files[0].AbsPath = "/outside/main.tf"
	require.ErrorContains(t, h.validateSourceFiles(context.Background(), source, files), "escapes repository")
}

func TestInitializeUsesJobFilesystem(t *testing.T) {
	for name, fs := range map[string]afero.Fs{"memory": afero.NewMemMapFs(), "os": afero.NewOsFs()} {
		t.Run(name, func(t *testing.T) {
			ctx := pkgctx.SetLogger(context.Background(), zap.NewNop())
			h := &handler{v: validator.New(), fs: fs, state: &handlerState{plan: &plantypes.BuildPlan{}}}
			job, execution := &models.AppRunnerJob{}, &models.AppRunnerJobExecution{ID: uuid.NewString()}
			require.NoError(t, h.Initialize(ctx, job, execution))
			t.Cleanup(func() { require.NoError(t, h.Cleanup(ctx, job, execution)) })
			require.NoError(t, afero.WriteFile(h.state.workspace.FS(), "main.tf", []byte("terraform { backend s3 {} }"), 0600))
			content, err := h.state.source.ReadFile("main.tf")
			require.NoError(t, err)
			require.Equal(t, "terraform { backend s3 {} }", string(content))
			outputs, err := h.Outputs(ctx)
			require.NoError(t, err)
			files := outputs["files"].([]ociarchive.FileRef)
			require.Len(t, files, 1)
			require.Equal(t, "main.tf", files[0].RelPath)
			require.NoError(t, h.state.arch.Pack(ctx, zap.NewNop(), h.state.source, files))
			_, err = h.state.source.ReadFile("../neighbor/main.tf")
			require.Error(t, err)
			source := h.state.source
			require.NoError(t, h.Cleanup(ctx, job, execution))
			_, err = source.ReadFile("main.tf")
			require.Error(t, err)
		})
	}
}
