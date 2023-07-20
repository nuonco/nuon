package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/helm/waypoint"
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

func TestInstallWaypoint(t *testing.T) {
	errInstallWaypoint := fmt.Errorf("install-waypoint-err")

	tests := map[string]struct {
		kubeconfig    *rest.Config
		requestFn     func() InstallWaypointRequest
		errExpected   error
		helmInstallFn func(*testing.T) testHelmInstaller
	}{
		"errors without namespace": {
			requestFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				req.Namespace = ""
				return req
			},
			errExpected: fmt.Errorf("Namespace"),
		},

		"errors without release name": {
			requestFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				req.ReleaseName = ""
				return req
			},
			errExpected: fmt.Errorf("ReleaseName"),
		},

		"errors without chart": {
			requestFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				req.Chart = nil
				return req
			},
			errExpected: fmt.Errorf("Chart"),
		},

		"errors without chart name": {
			requestFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				req.Chart.Name = ""
				return req
			},
			errExpected: fmt.Errorf("Name"),
		},

		"errors without chart version": {
			requestFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				req.Chart.Version = ""
				return req
			},
			errExpected: fmt.Errorf("Version"),
		},

		"uses api": {
			requestFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				req.ReleaseName = "test-release"
				req.Chart.Name = "waypoint"
				return req
			},
			helmInstallFn: func(t *testing.T) testHelmInstaller {
				return testHelmInstaller{
					fn: func(ctx context.Context, cfg *helm.InstallConfig) (*release.Release, error) {
						assert.NotNil(t, cfg)
						assert.Equal(t, "test-release", cfg.ReleaseName)
						assert.Equal(t, "waypoint", cfg.Chart.Name)
						return &release.Release{Name: cfg.ReleaseName}, nil
					},
				}
			},
		},

		"configures values correctly": {
			requestFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				return req
			},
			helmInstallFn: func(t *testing.T) testHelmInstaller {
				return testHelmInstaller{
					fn: func(ctx context.Context, cfg *helm.InstallConfig) (*release.Release, error) {
						assert.NotNil(t, cfg)
						var vals waypoint.Values
						err := mapstructure.Decode(cfg.Values, &vals)
						assert.Nil(t, err)
						assert.True(t, vals.Runner.Enabled)
						assert.False(t, vals.Server.Enabled)

						assert.NotEmpty(t, vals.Runner.Image.Repository)
						assert.NotEmpty(t, vals.Runner.Image.Tag)
						return &release.Release{Name: cfg.ReleaseName}, nil
					},
				}
			},
		},

		"wraps errors": {
			requestFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				return req
			},
			errExpected: errInstallWaypoint,
			helmInstallFn: func(t *testing.T) testHelmInstaller {
				return testHelmInstaller{
					fn: func(ctx context.Context, cfg *helm.InstallConfig) (*release.Release, error) {
						return nil, errInstallWaypoint
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
			enc, err := env.ExecuteActivity(a.InstallWaypoint, test.requestFn())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			resp := InstallWaypointResponse{}
			err = enc.Get(&resp)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

func Test_waypointRunnerValues(t *testing.T) {
	tests := map[string]struct {
		requestFn   func() InstallWaypointRequest
		errExpected error
		assertFn    func(*testing.T, map[string]interface{}, InstallWaypointRequest)
	}{
		"happy path": {
			requestFn: func() InstallWaypointRequest {
				return generics.GetFakeObj[InstallWaypointRequest]()
			},
			assertFn: func(t *testing.T, v map[string]interface{}, req InstallWaypointRequest) {
				var vals waypoint.Values
				err := mapstructure.Decode(v, &vals)
				assert.Nil(t, err)

				assert.True(t, vals.Runner.Enabled)
				assert.True(t, vals.Runner.Server.TLS)
				assert.True(t, vals.Runner.Server.TLSSkipVerify)
				assert.Equal(t, req.RunnerConfig.Cookie, vals.Runner.Server.Cookie)
				assert.Equal(t, req.RunnerConfig.ServerAddr, vals.Runner.Server.Addr)
				assert.Equal(t, req.RunnerConfig.ID, vals.Runner.ID)

				assert.False(t, vals.Server.Enabled)
				assert.False(t, vals.UI.Service.Enabled)
				assert.False(t, vals.Bootstrap.ServiceAccount.Create)

				odrAnnots := vals.Runner.Odr.ServiceAccount.Annotations
				expectedKey := "eks.amazonaws.com/role-arn"
				assert.NotEmpty(t, odrAnnots)
				assert.Equal(t, req.RunnerConfig.OdrIAMRoleArn, odrAnnots[expectedKey])
			},
		},
		"errors with invalid request": {
			requestFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				req.Namespace = ""
				return req
			},
			errExpected: fmt.Errorf("Namespace"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := test.requestFn()
			vals, err := getWaypointRunnerValues(req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.Nil(t, err)
			assert.NotNil(t, vals)
			test.assertFn(t, vals, req)
		})
	}
}
