package kubernetes_manifest

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	ocicopy "github.com/nuonco/nuon/pkg/runner/oci/copy"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type manifestResultClient struct {
	nuonrunner.Client
	reported bool
}

func (c *manifestResultClient) CreateJobExecutionResult(_ context.Context, _, _ string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	c.reported = req.Success
	return &models.AppRunnerJobExecutionResult{}, nil
}

type manifestCopy struct {
	ocicopy.Copier
	inspect func(context.Context, oras.ReadOnlyTarget, string) *ocispec.Descriptor
}

func (c *manifestCopy) CopyFromStore(ctx context.Context, store oras.ReadOnlyTarget, tag string, _ *configs.OCIRegistryRepository, _ string) (*ocispec.Descriptor, error) {
	return c.inspect(ctx, store, tag), nil
}

func TestExecPackagesPinnedManifest(t *testing.T) {
	for _, backend := range []string{"default disk", "memory", "hardlink"} {
		for _, sourceType := range []string{"inline", "kustomize"} {
			t.Run(backend+"/"+sourceType, func(t *testing.T) {
				ctx := pkgctx.SetLogger(context.Background(), zap.NewNop())
				job := &models.AppRunnerJob{ID: "job", OwnerID: "build"}
				execution := &models.AppRunnerJobExecution{ID: uuid.NewString()}
				manifest := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: pinned-source\n"
				client := &manifestResultClient{}
				copied := false
				copier := &manifestCopy{inspect: func(ctx context.Context, store oras.ReadOnlyTarget, tag string) *ocispec.Descriptor {
					descriptor, err := store.Resolve(ctx, tag)
					require.NoError(t, err)
					reader, err := store.Fetch(ctx, descriptor)
					require.NoError(t, err)
					var image ocispec.Manifest
					require.NoError(t, json.NewDecoder(reader).Decode(&image))
					require.NoError(t, reader.Close())
					require.Len(t, image.Layers, 1)
					layer, err := store.Fetch(ctx, image.Layers[0])
					require.NoError(t, err)
					defer layer.Close()
					compressed, err := gzip.NewReader(layer)
					require.NoError(t, err)
					defer compressed.Close()
					archive := tar.NewReader(compressed)
					header, err := archive.Next()
					require.NoError(t, err)
					require.Equal(t, defaultManifestFilename, header.Name)
					data, err := io.ReadAll(archive)
					require.NoError(t, err)
					require.Equal(t, manifest, string(data))
					_, err = archive.Next()
					require.ErrorIs(t, err, io.EOF)
					copied = true
					return &descriptor
				}}
				var fs afero.Fs
				if backend == "memory" {
					fs = afero.NewMemMapFs()
				}
				h, err := New(HandlerParams{V: validator.New(), APIClient: client, OCICopy: copier, FS: fs})
				require.NoError(t, err)
				h.state = &handlerState{cfg: &plantypes.KubernetesManifestBuildPlan{SourceType: "inline", InlineManifest: manifest}}
				require.NoError(t, h.Initialize(ctx, job, execution))
				t.Cleanup(func() { require.NoError(t, h.Cleanup(ctx, job, execution)) })
				if backend == "memory" {
					_, err := os.Stat(h.state.workspace.Root())
					require.True(t, os.IsNotExist(err), "injected workspace must not fall through to disk")
				} else {
					info, err := os.Stat(h.state.workspace.Root())
					require.NoError(t, err)
					require.True(t, info.IsDir())
				}
				if sourceType == "kustomize" {
					kustomizeFixture(t, h.state.workspace.FS(), map[string]string{
						"overlay/kustomization.yaml": "resources: [resource.yaml]\n",
						"overlay/resource.yaml":      manifest,
					})
					h.state.cfg.SourceType = "kustomize"
					h.state.cfg.KustomizePath = "overlay"
				}
				output := h.state.arch.BasePath()
				neighbor := t.TempDir()
				canaryPath := filepath.Join(neighbor, defaultManifestFilename)
				require.NoError(t, os.WriteFile(canaryPath, []byte("neighbor-canary"), 0600))
				if backend == "hardlink" {
					require.NoError(t, os.Link(canaryPath, filepath.Join(output, defaultManifestFilename)))
					require.ErrorIs(t, h.Exec(ctx, job, execution), os.ErrExist)
					require.False(t, copied)
					require.False(t, client.reported)
					canary, err := os.ReadFile(canaryPath)
					require.NoError(t, err)
					require.Equal(t, "neighbor-canary", string(canary))
					return
				}
				require.NoError(t, os.Rename(output, output+"-moved"))
				require.NoError(t, os.Symlink(neighbor, output))
				t.Cleanup(func() {
					require.NoError(t, os.Remove(output))
					require.NoError(t, os.Rename(output+"-moved", output))
				})
				require.NoError(t, h.Exec(ctx, job, execution))
				require.True(t, copied)
				require.True(t, client.reported)
				canary, err := os.ReadFile(canaryPath)
				require.NoError(t, err)
				require.Equal(t, "neighbor-canary", string(canary))
			})
		}
	}
}
