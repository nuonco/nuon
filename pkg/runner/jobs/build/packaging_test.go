package build_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	"github.com/nuonco/nuon/pkg/runner/jobs/build/pulumi"
	"github.com/nuonco/nuon/pkg/runner/jobs/build/sandbox"
	"github.com/nuonco/nuon/pkg/runner/jobs/build/terraform"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	ocicopy "github.com/nuonco/nuon/pkg/runner/oci/copy"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type packagingClient struct {
	nuonrunner.Client
	plan   *models.PlantypesCompositePlan
	result *models.ServiceCreateRunnerJobExecutionResultRequest
}

func (c *packagingClient) GetJobCompositePlan(context.Context, string) (*models.PlantypesCompositePlan, error) {
	return c.plan, nil
}

func (c *packagingClient) CreateJobExecutionResult(_ context.Context, _, _ string, result *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	c.result = result
	return &models.AppRunnerJobExecutionResult{}, nil
}

type packagingCopy struct {
	ocicopy.Copier
	archive ociarchive.Archive
}

func (c *packagingCopy) CopyFromStore(ctx context.Context, source oras.ReadOnlyTarget, tag string, _ *configs.OCIRegistryRepository, _ string) (*v1.Descriptor, error) {
	descriptor, err := oras.Copy(ctx, source, tag, c.archive.Ref().(oras.Target), "latest", oras.DefaultCopyOptions)
	return &descriptor, err
}

func TestBuildPackagingKeepsPinnedSource(t *testing.T) {
	t.Setenv("IS_NUONCTL", "false")
	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, "component"), 0o700))
	want := "terraform { backend s3 {} }\n# owned source\n"
	require.NoError(t, os.WriteFile(filepath.Join(repo, "component/main.tf"), []byte(want), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "not-in-component.txt"), []byte("not packaged"), 0o600))
	for _, args := range [][]string{{"init", "-b", "main"}, {"add", "."}, {"commit", "-m", "fixture"}} {
		cmd := exec.Command("git", append([]string{"-c", "user.name=Example", "-c", "user.email=example@example.com", "-c", "commit.gpgsign=false"}, args...)...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s", output)
	}
	for _, backend := range []string{"disk", "memory"} {
		for _, kind := range []string{"terraform", "pulumi", "sandbox", "terraform-vendor", "sandbox-vendor"} {
			vendor := strings.HasSuffix(kind, "-vendor")
			if vendor && backend == "disk" {
				continue
			}
			t.Run(backend+"/"+kind, func(t *testing.T) {
				ctx := pkgctx.SetLogger(context.Background(), zap.NewNop())
				plan := &plantypes.CompositePlan{BuildPlan: &plantypes.BuildPlan{
					Src:                &plantypes.GitSource{URL: repo, Ref: "main", Path: "component"},
					Dst:                &configs.OCIRegistryRepository{},
					TerraformBuildPlan: &plantypes.TerraformBuildPlan{VendorProviders: vendor},
				}}
				data, err := json.Marshal(plan)
				require.NoError(t, err)
				client := &packagingClient{plan: &models.PlantypesCompositePlan{}}
				require.NoError(t, json.Unmarshal(data, client.plan))
				unpacked := ociarchive.New()
				require.NoError(t, unpacked.Initialize(ctx))
				t.Cleanup(func() { require.NoError(t, unpacked.Cleanup(ctx)) })
				copier := &packagingCopy{archive: unpacked}
				var fs afero.Fs
				if backend == "memory" {
					fs = afero.NewMemMapFs()
				}
				var handler jobs.JobHandler
				switch strings.TrimSuffix(kind, "-vendor") {
				case "terraform":
					handler, err = terraform.New(terraform.Params{V: validator.New(), APIClient: client, OCICopy: copier, FS: fs})
				case "pulumi":
					handler, err = pulumi.New(pulumi.Params{V: validator.New(), APIClient: client, OCICopy: copier, FS: fs})
				case "sandbox":
					handler, err = sandbox.New(sandbox.HandlerParams{V: validator.New(), APIClient: client, OCICopy: copier, FS: fs})
				}
				require.NoError(t, err)
				job := &models.AppRunnerJob{ID: "job", OwnerID: "artifact"}
				execution := &models.AppRunnerJobExecution{ID: "packaging-" + uuid.NewString()}
				t.Cleanup(func() { require.NoError(t, handler.Cleanup(ctx, job, execution)) })
				require.NoError(t, handler.Fetch(ctx, job, execution))
				require.NoError(t, handler.Validate(ctx, job, execution))
				require.NoError(t, handler.Initialize(ctx, job, execution))
				root := filepath.Join("/tmp", "workspace-"+execution.ID)
				if backend == "disk" {
					require.NoError(t, os.Rename(root, root+"-moved"))
					t.Cleanup(func() { require.NoError(t, os.RemoveAll(root+"-moved")) })
				} else {
					_, err := os.Stat(root)
					require.True(t, os.IsNotExist(err))
				}
				require.NoError(t, os.MkdirAll(filepath.Join(root, "component"), 0o700))
				t.Cleanup(func() { require.NoError(t, os.RemoveAll(root)) })
				require.NoError(t, os.WriteFile(filepath.Join(root, "component/main.tf"), []byte("host canary"), 0o600))
				if vendor {
					require.ErrorContains(t, handler.Exec(ctx, job, execution), "provider vendoring requires a disk-backed workspace")
					contents, err := os.ReadFile(filepath.Join(root, "component/main.tf"))
					require.NoError(t, err)
					require.Equal(t, "host canary", string(contents))
					entries, err := os.ReadDir(filepath.Join(root, "component"))
					require.NoError(t, err)
					require.Len(t, entries, 1)
					require.Nil(t, client.result)
					return
				}
				require.NoError(t, handler.Exec(ctx, job, execution))
				require.NotNil(t, client.result)
				require.True(t, client.result.Success)
				contents, err := afero.ReadFile(unpacked.FS(), "main.tf")
				require.NoError(t, err)
				require.Equal(t, want, string(contents))
				_, err = unpacked.FS().Stat("not-in-component.txt")
				require.True(t, os.IsNotExist(err))
				outputs, err := handler.Outputs(ctx)
				require.NoError(t, err)
				files := outputs["files"].([]ociarchive.FileRef)
				require.Len(t, files, 1)
				require.Equal(t, "main.tf", files[0].RelPath)
				require.NoError(t, handler.Cleanup(ctx, job, execution))
			})
		}
	}
}
