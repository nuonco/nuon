package imagemetadata_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"oras.land/oras-go/v2/registry/remote/retry"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/errs"
	"github.com/nuonco/nuon/pkg/runner/jobs/sync/imagemetadata"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type planClient struct {
	nuonrunner.Client
	plan *models.PlantypesCompositePlan
}

func (c *planClient) GetJobCompositePlan(context.Context, string) (*models.PlantypesCompositePlan, error) {
	return c.plan, nil
}

type registryTransport func(*http.Request) (*http.Response, error)

func (f registryTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMetadataUsesAuthorizedRegistry(t *testing.T) {
	for name, tc := range map[string]struct{ host, repository string }{
		"ecr": {"123456789012.dkr.ecr.us-west-2.amazonaws.com", "acme/api"},
		"gar": {"us-central1-docker.pkg.dev", "acme-project/acme-repo/api"},
		"acr": {"acme.azurecr.io", "acme/api"},
	} {
		t.Run(name, func(t *testing.T) {
			var requestedHost, requestedPath string
			old := retry.DefaultClient
			retry.DefaultClient = &http.Client{Transport: registryTransport(func(r *http.Request) (*http.Response, error) {
				requestedHost, requestedPath = r.URL.Host, r.URL.Path
				return &http.Response{StatusCode: http.StatusForbidden, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
			})}
			t.Cleanup(func() { retry.DefaultClient = old })

			tag := "release-42"
			plan := &models.PlantypesFetchImageMetadataPlan{Tag: &tag}
			plan.Registry.RegistryType = models.ConfigsOCIRegistryTypePrivateOci
			plan.Registry.Repository = tc.repository
			plan.Registry.LoginServer = tc.host
			plan.Registry.Ociauth = &models.ConfigsOCIRegistryAuth{Username: "reader", Password: "test-password"}
			client := &planClient{plan: &models.PlantypesCompositePlan{FetchImageMetadataPlan: plan}}
			h, err := imagemetadata.New(imagemetadata.HandlerParams{
				V: validator.New(), APIClient: client,
				ErrRecorder: errs.NewRecorder(errs.Params{L: zap.NewNop(), LC: fxtest.NewLifecycle(t)}),
			})
			require.NoError(t, err)
			ctx := pkgctx.SetLogger(context.Background(), zap.NewNop())
			job, execution := &models.AppRunnerJob{}, &models.AppRunnerJobExecution{}
			require.NoError(t, h.Fetch(ctx, job, execution))
			err = h.Exec(ctx, job, execution)
			require.Equal(t, tc.host, requestedHost)
			require.Equal(t, "/v2/"+tc.repository+"/manifests/"+tag, requestedPath)
			require.ErrorContains(t, err, "403")
		})
	}
}
