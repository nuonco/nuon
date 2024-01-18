package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
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

func TestInstallWaypoint(t *testing.T) {
	errInstallWaypoint := fmt.Errorf("error installing waypoint")

	tests := map[string]struct {
		kubeconfig    *rest.Config
		reqFn         func() InstallWaypointRequest
		errExpected   error
		helmInstallFn func(*testing.T) testHelmInstaller
	}{
		"errors without namespace": {
			reqFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				req.Namespace = ""
				return req
			},
			errExpected: fmt.Errorf("Namespace"),
		},

		"errors without chart": {
			reqFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				req.Chart = nil
				return req
			},
			errExpected: fmt.Errorf("Chart"),
		},

		"errors without chart name": {
			reqFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				req.Chart.Name = ""
				return req
			},
			errExpected: fmt.Errorf("Name"),
		},

		"errors without chart version": {
			reqFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				req.Chart.Version = ""
				return req
			},
			errExpected: fmt.Errorf("Version"),
		},

		"uses api": {
			reqFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
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

		"configures values correctly": {
			reqFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				return req
			},
			helmInstallFn: func(t *testing.T) testHelmInstaller {
				return testHelmInstaller{
					fn: func(ctx context.Context, cfg *helm.InstallConfig) (*release.Release, error) {
						assert.NotNil(t, cfg)
						var vals waypointhelm.Values
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
			reqFn: func() InstallWaypointRequest {
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
			enc, err := env.ExecuteActivity(a.InstallWaypoint, test.reqFn())
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
	cookie := uuid.NewString()
	addr := fmt.Sprintf("%s.stage.nuon.co", uuid.NewString())

	tests := map[string]struct {
		reqFn       func() InstallWaypointRequest
		errExpected error
		assertFn    func(*testing.T, InstallWaypointRequest, map[string]interface{})
	}{
		"happy path": {
			reqFn: func() InstallWaypointRequest {
				req := generics.GetFakeObj[InstallWaypointRequest]()
				req.RunnerConfig.Cookie = cookie
				req.RunnerConfig.ServerAddr = addr
				return req
			},
			assertFn: func(t *testing.T, req InstallWaypointRequest, v map[string]interface{}) {
				var vals waypointhelm.Values
				err := mapstructure.Decode(v, &vals)
				assert.Nil(t, err)

				// assert runner server ocnnection
				assert.True(t, vals.Runner.Server.TLS)
				assert.True(t, vals.Runner.Server.TLSSkipVerify)
				assert.Equal(t, cookie, vals.Runner.Server.Cookie)
				assert.Equal(t, addr, vals.Runner.Server.Addr)

				assert.False(t, vals.Server.Enabled)
				assert.False(t, vals.UI.Service.Enabled)
				assert.False(t, vals.Bootstrap.ServiceAccount.Create)

				// assert odr setup
				assert.True(t, vals.Runner.Odr.ServiceAccount.Create)
				assert.Equal(t, runnerOdrServiceAccountName(req.InstallID), vals.Runner.Odr.ServiceAccount.Name)

				// assert runner setup
				assert.True(t, vals.Runner.Enabled)
				assert.True(t, vals.Runner.ServiceAccount.Create)
				assert.Equal(t, runnerServiceAccountName(req.InstallID), vals.Runner.ServiceAccount.Name)
			},
		},
		"errors with invalid request": {
			reqFn: func() InstallWaypointRequest {
				return InstallWaypointRequest{}
			},
			errExpected: fmt.Errorf("Namespace"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := test.reqFn()
			vals, err := getWaypointRunnerValues(req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.Nil(t, err)
			assert.NotNil(t, vals)
			test.assertFn(t, req, vals)
		})
	}
}
