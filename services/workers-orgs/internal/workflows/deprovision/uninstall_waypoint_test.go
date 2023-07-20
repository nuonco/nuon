package deprovision

import (
	"context"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	"helm.sh/helm/v3/pkg/release"

	"k8s.io/client-go/rest"
)

type testHelmUninstaller struct {
	fn func(context.Context, *helm.UninstallConfig) (*release.UninstallReleaseResponse, error)
}

func (t testHelmUninstaller) Uninstall(ctx context.Context, cfg *helm.UninstallConfig) (*release.UninstallReleaseResponse, error) {
	return t.fn(ctx, cfg)
}

func TestUninstallWaypoint(t *testing.T) {
	tests := map[string]struct {
		kubeconfig      *rest.Config
		request         UninstallWaypointRequest
		errExpected     error
		helmUninstallFn func(*testing.T) testHelmUninstaller
	}{
		"uses api": {
			request: UninstallWaypointRequest{
				Namespace:   "test",
				ReleaseName: "test-release",
				ClusterInfo: generics.GetFakeObj[kube.ClusterInfo](),
			},
			helmUninstallFn: func(t *testing.T) testHelmUninstaller {
				return testHelmUninstaller{
					fn: func(ctx context.Context, cfg *helm.UninstallConfig) (*release.UninstallReleaseResponse, error) {
						assert.NotNil(t, cfg)
						assert.Equal(t, "test-release", cfg.ReleaseName)
						assert.Equal(t, "test", cfg.Namespace)
						return &release.UninstallReleaseResponse{}, nil
					},
				}
			},
		},

		"wraps errors": {
			request: UninstallWaypointRequest{
				Namespace:   "test",
				ReleaseName: "test-release",
				ClusterInfo: generics.GetFakeObj[kube.ClusterInfo](),
			},
			errExpected: errOops,
			helmUninstallFn: func(t *testing.T) testHelmUninstaller {
				return testHelmUninstaller{
					fn: func(ctx context.Context, cfg *helm.UninstallConfig) (*release.UninstallReleaseResponse, error) {
						return nil, errOops
					},
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()
			a := &Activities{}
			a.Kubeconfig = &rest.Config{}
			env.RegisterActivity(a)

			if test.helmUninstallFn != nil {
				a.helmUninstaller = test.helmUninstallFn(t)
			}
			enc, err := env.ExecuteActivity(a.UninstallWaypoint, test.request)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			resp := UninstallWaypointResponse{}
			err = enc.Get(&resp)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}
