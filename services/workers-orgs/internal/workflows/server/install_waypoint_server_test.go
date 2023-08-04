package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/pkg/deprecated/helm"
	"github.com/powertoolsdev/mono/pkg/generics"
	waypointhelm "github.com/powertoolsdev/mono/pkg/waypoint/helm"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	"helm.sh/helm/v3/pkg/release"

	"k8s.io/client-go/rest"
)

type testHelmInstaller struct {
	fn func(context.Context, *helm.InstallConfig) (*release.Release, error)
}

func (t testHelmInstaller) Install(ctx context.Context, cfg *helm.InstallConfig) (*release.Release, error) {
	return t.fn(ctx, cfg)
}

func TestInstallWaypointServer(t *testing.T) {
	tests := map[string]struct {
		kubeconfig    *rest.Config
		requestFn     func() InstallWaypointServerRequest
		errExpected   error
		helmInstallFn func(*testing.T) testHelmInstaller
	}{
		"errors without namespace": {
			requestFn: func() InstallWaypointServerRequest {
				req := generics.GetFakeObj[InstallWaypointServerRequest]()
				req.Namespace = ""
				return req
			},
			errExpected: fmt.Errorf("Namespace"),
		},

		"errors without release name": {
			requestFn: func() InstallWaypointServerRequest {
				req := generics.GetFakeObj[InstallWaypointServerRequest]()
				req.ReleaseName = ""
				return req
			},
			errExpected: fmt.Errorf("ReleaseName"),
		},

		"errors without chart": {
			requestFn: func() InstallWaypointServerRequest {
				req := generics.GetFakeObj[InstallWaypointServerRequest]()
				req.Chart = nil
				return req
			},
			errExpected: fmt.Errorf("Chart"),
		},

		"errors without chart name": {
			requestFn: func() InstallWaypointServerRequest {
				req := generics.GetFakeObj[InstallWaypointServerRequest]()
				req.Chart.Name = ""
				return req
			},
			errExpected: fmt.Errorf("Name"),
		},

		"errors without chart version": {
			requestFn: func() InstallWaypointServerRequest {
				req := generics.GetFakeObj[InstallWaypointServerRequest]()
				req.Chart.Version = ""
				return req
			},
			errExpected: fmt.Errorf("Version"),
		},

		"uses api": {
			requestFn: func() InstallWaypointServerRequest {
				req := generics.GetFakeObj[InstallWaypointServerRequest]()
				req.ReleaseName = "test-release"
				return req
			},
			helmInstallFn: func(t *testing.T) testHelmInstaller {
				return testHelmInstaller{
					fn: func(ctx context.Context, cfg *helm.InstallConfig) (*release.Release, error) {
						assert.NotNil(t, cfg)
						assert.Equal(t, "test-release", cfg.ReleaseName)
						return &release.Release{Name: cfg.ReleaseName}, nil
					},
				}
			},
		},

		"uses the correct values": {
			requestFn: func() InstallWaypointServerRequest {
				req := generics.GetFakeObj[InstallWaypointServerRequest]()
				return req
			},
			helmInstallFn: func(t *testing.T) testHelmInstaller {
				return testHelmInstaller{
					fn: func(ctx context.Context, cfg *helm.InstallConfig) (*release.Release, error) {
						assert.NotNil(t, cfg)

						var vals waypointhelm.Values
						err := mapstructure.Decode(cfg.Values, &vals)
						assert.Nil(t, err)
						assert.True(t, vals.Server.Enabled)
						assert.False(t, vals.Bootstrap.ServiceAccount.Create)

						// assert that the image values are set correctly
						assert.NotEmpty(t, vals.Server.Image.Repository)
						assert.NotEmpty(t, vals.Server.Image.Tag)
						return &release.Release{Name: cfg.ReleaseName}, nil
					},
				}
			},
		},

		"wraps errors": {
			requestFn: func() InstallWaypointServerRequest {
				req := generics.GetFakeObj[InstallWaypointServerRequest]()
				return req
			},
			errExpected: errOops,
			helmInstallFn: func(t *testing.T) testHelmInstaller {
				return testHelmInstaller{
					fn: func(ctx context.Context, cfg *helm.InstallConfig) (*release.Release, error) {
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

			if test.helmInstallFn != nil {
				a.helmInstaller = test.helmInstallFn(t)
			}
			req := test.requestFn()
			enc, err := env.ExecuteActivity(a.InstallWaypointServer, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			resp := InstallWaypointServerResponse{}
			err = enc.Get(&resp)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}
