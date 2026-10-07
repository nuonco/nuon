package containerimage

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"

	"github.com/nuonco/nuon/pkg/oci/signature"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const signedDigest = "sha256:96d02f455d5a73b817c0602910748609cf8471b1cc9522f78c75cedb1f67d072"

type fixedResolver struct{}

func (fixedResolver) Resolve(context.Context, *configs.OCIRegistryRepository, string) (*ocispec.Descriptor, error) {
	return &ocispec.Descriptor{Digest: signedDigest, MediaType: ocispec.MediaTypeImageIndex}, nil
}

func (fixedResolver) Tags(context.Context, *configs.OCIRegistryRepository) ([]string, error) {
	return nil, nil
}

type recordingCopier struct{ copied bool }

func (c *recordingCopier) Copy(context.Context, *configs.OCIRegistryRepository, string, *configs.OCIRegistryRepository, string) (*ocispec.Descriptor, error) {
	c.copied = true
	return &ocispec.Descriptor{}, nil
}

func (c *recordingCopier) CopyFromStore(context.Context, oras.ReadOnlyTarget, string, *configs.OCIRegistryRepository, string) (*ocispec.Descriptor, error) {
	return nil, nil
}

func (c *recordingCopier) CopyFromLocalRegistry(context.Context, string, *configs.OCIRegistryRepository, string) (*ocispec.Descriptor, error) {
	return nil, nil
}

func runSignedImageBuild(t *testing.T, subject string) (*models.ServiceCreateRunnerJobExecutionResultRequest, bool, error) {
	ctrl := gomock.NewController(t)
	client := nuonrunner.NewMockClient(ctrl)

	var result *models.ServiceCreateRunnerJobExecutionResultRequest
	client.EXPECT().
		CreateJobExecutionResult(gomock.Any(), "job-1", "exec-1", gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
			result = req
			return &models.AppRunnerJobExecutionResult{}, nil
		})

	copier := &recordingCopier{}
	h := &handler{
		apiClient:  client,
		ociCopy:    copier,
		ociResolve: fixedResolver{},
		state: &handlerState{
			jobID:          "job-1",
			jobExecutionID: "exec-1",
			resultTag:      "bld-1",
			plan:           &plantypes.BuildPlan{},
			regCfg:         &configs.OCIRegistryRepository{},
			cfg: &plantypes.ContainerImagePullPlan{
				Image: "cgr.dev/chainguard/static",
				Tag:   "latest",
				RepoCfg: &configs.OCIRegistryRepository{
					RegistryType: configs.OCIRegistryTypePublicOCI,
					Repository:   "cgr.dev/chainguard/static",
					OCIAuth:      &configs.OCIRegistryAuth{},
				},
				Verification: &signature.Verification{
					RequireSignature: true,
					Authorities: []signature.Authority{{
						Type:    signature.AuthorityTypeKeyless,
						Issuer:  "https://token.actions.githubusercontent.com",
						Subject: subject,
					}},
				},
			},
		},
	}

	ctx := pkgctx.SetLogger(context.Background(), zap.NewNop())
	err := h.Exec(ctx, &models.AppRunnerJob{ID: "job-1"}, &models.AppRunnerJobExecution{ID: "exec-1"})
	return result, copier.copied, err
}

func TestExecReportsSignatureOutcome(t *testing.T) {
	if os.Getenv("NUON_INTEGRATION") == "" {
		t.Skip("set NUON_INTEGRATION to run registry verification")
	}

	t.Run("trusted authority", func(t *testing.T) {
		result, copied, err := runSignedImageBuild(t, "https://github.com/chainguard-images/images/.github/workflows/release.yaml@refs/heads/main")
		require.NoError(t, err)
		require.True(t, copied)
		require.True(t, result.Success)
		require.Equal(t, signedDigest, result.SourceDigest)
	})

	t.Run("untrusted authority", func(t *testing.T) {
		result, copied, err := runSignedImageBuild(t, "https://github.com/acme/images/.github/workflows/release.yaml@refs/heads/main")
		require.Error(t, err)
		require.False(t, copied)
		require.False(t, result.Success)
		require.Equal(t, signature.VerifyStep, result.ErrorMetadata["step"])
	})
}
