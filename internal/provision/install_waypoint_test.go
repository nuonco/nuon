package provision

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/go-helm"
	"github.com/powertoolsdev/go-helm/waypoint"
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
	tests := map[string]struct {
		kubeconfig    *rest.Config
		request       InstallWaypointRequest
		errExpected   error
		helmInstallFn func(*testing.T) testHelmInstaller
	}{
		"errors without namespace": {
			request:     InstallWaypointRequest{},
			errExpected: ErrInvalidNamespaceName,
		},

		"errors without release name": {
			request: InstallWaypointRequest{
				Namespace: "test",
			},
			errExpected: ErrInvalidReleaseName,
		},

		"errors without chart": {
			request: InstallWaypointRequest{
				Namespace:   "test",
				ReleaseName: "test-release",
			},
			errExpected: ErrInvalidChart,
		},

		"errors without chart name": {
			request: InstallWaypointRequest{
				Namespace:   "test",
				ReleaseName: "test-release",
				Chart:       &helm.Chart{},
			},
			errExpected: ErrInvalidChart,
		},

		"errors without chart url": {
			request: InstallWaypointRequest{
				Namespace:   "test",
				ReleaseName: "test-release",
				Chart:       &helm.Chart{Name: "test-chart"},
			},
			errExpected: ErrInvalidChart,
		},

		"errors without chart version": {
			request: InstallWaypointRequest{
				Namespace:   "test",
				ReleaseName: "test-release",
				Chart:       &helm.Chart{Name: "waypoint", URL: "https://helm.releases.hashicorp.com"},
			},
			errExpected: ErrInvalidChart,
		},

		"uses api": {
			request: InstallWaypointRequest{
				Namespace:   "test",
				ReleaseName: "test-release",
				Chart:       &waypoint.DefaultChart,
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
			request: InstallWaypointRequest{
				Namespace:   "test",
				ReleaseName: "test-release",
				Chart:       &waypoint.DefaultChart,
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
						return &release.Release{Name: cfg.ReleaseName}, nil
					},
				}
			},
		},

		"wraps errors": {
			request: InstallWaypointRequest{
				Namespace:   "test",
				ReleaseName: "test-release",
				Chart:       &waypoint.DefaultChart,
			},
			errExpected: ErrInvalidNamespaceName,
			helmInstallFn: func(t *testing.T) testHelmInstaller {
				return testHelmInstaller{
					fn: func(ctx context.Context, cfg *helm.InstallConfig) (*release.Release, error) {
						return nil, ErrInvalidNamespaceName
					},
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()
			a := &ProvisionActivities{}
			a.Kubeconfig = &rest.Config{}
			env.RegisterActivity(a)

			if test.helmInstallFn != nil {
				a.helmInstaller = test.helmInstallFn(t)
			}
			enc, err := env.ExecuteActivity(a.InstallWaypoint, test.request)
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
	installID := uuid.NewString()
	cookie := uuid.NewString()
	addr := fmt.Sprintf("%s.stage.nuon.co", uuid.NewString())

	tests := map[string]struct {
		req         InstallWaypointRequest
		errExpected error
		assertFn    func(*testing.T, map[string]interface{})
	}{
		"happy path": {
			req: InstallWaypointRequest{
				Namespace:   installID,
				ReleaseName: "v10.0.1",
				Chart:       &waypoint.DefaultChart,
				RunnerConfig: RunnerConfig{
					ID:         installID,
					Cookie:     cookie,
					ServerAddr: addr,
				},
			},
			assertFn: func(t *testing.T, v map[string]interface{}) {
				var vals waypoint.Values
				err := mapstructure.Decode(v, &vals)
				assert.Nil(t, err)

				assert.True(t, vals.Runner.Enabled)
				assert.True(t, vals.Runner.Server.TLS)
				assert.True(t, vals.Runner.Server.TLSSkipVerify)
				assert.Equal(t, cookie, vals.Runner.Server.Cookie)
				assert.Equal(t, addr, vals.Runner.Server.Addr)

				assert.False(t, vals.Server.Enabled)
				assert.False(t, vals.UI.Service.Enabled)
				assert.False(t, vals.Bootstrap.ServiceAccount.Create)
			},
		},
		"errors with invalid request": {
			req:         InstallWaypointRequest{},
			errExpected: ErrInvalidNamespaceName,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			vals, err := getWaypointRunnerValues(test.req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.Nil(t, err)
			assert.NotNil(t, vals)
			test.assertFn(t, vals)
		})
	}
}
