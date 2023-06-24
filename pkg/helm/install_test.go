package helm

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var errOops = errors.New("oops")

type testChartLoader struct {
	fn func(string, action.ChartPathOptions) (*chart.Chart, error)
}

func (t testChartLoader) load(s string, a action.ChartPathOptions) (*chart.Chart, error) {
	return t.fn(s, a)
}

func Test_Install(t *testing.T) {
	tests := map[string]struct {
		loader      func(*testing.T) testChartLoader
		errExpected error
		values      map[string]interface{}
	}{
		"succeeds with valid chart": {
			loader: func(t *testing.T) testChartLoader {
				return testChartLoader{
					fn: func(s string, i action.ChartPathOptions) (*chart.Chart, error) {
						assert.Equal(t, t.Name(), s)
						assert.NotNil(t, i)
						return buildChart(withName(t.Name()), withVersion(i.Version)), nil
					},
				}
			},
		},
		"wraps loader error": {
			loader: func(t *testing.T) testChartLoader {
				return testChartLoader{
					fn: func(s string, i action.ChartPathOptions) (*chart.Chart, error) {
						assert.Equal(t, t.Name(), s)
						assert.NotNil(t, i)
						return nil, errOops
					},
				}
			},
			errExpected: errOops,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			w := helmInstallRunner{chartLoader: test.loader(t)}
			cfg := actionConfigFixture(t)
			c := action.NewInstall(cfg)

			c.Atomic = true
			c.Namespace = "test-ns"
			c.RepoURL = "url"
			c.Version = "v1.2.3"
			c.ReleaseName = "test-release"

			release, err := w.install(context.Background(), c, t.Name(), test.values)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, release)
			assert.Equal(t, c.ReleaseName, release.Name)
		})
	}
}

var manifestWithHook = `kind: ConfigMap
metadata:
name: test-cm
annotations:
"helm.sh/hook": post-install,pre-delete,post-upgrade
data:
name: value`

type chartOptions struct {
	*chart.Chart
}

type chartOption func(*chartOptions)

func buildChart(opts ...chartOption) *chart.Chart {
	c := &chartOptions{
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				APIVersion: "v1",
				Name:       "hello",
				Version:    "0.1.0",
			},
			// This adds a basic template and hooks.
			Templates: []*chart.File{
				{Name: "templates/hello", Data: []byte("hello: world")},
				{Name: "templates/hooks", Data: []byte(manifestWithHook)},
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c.Chart
}

func withName(name string) chartOption {
	return func(opts *chartOptions) {
		opts.Metadata.Name = name
	}
}

func withVersion(version string) chartOption {
	return func(opts *chartOptions) {
		opts.Metadata.Version = version
	}
}

func actionConfigFixture(t *testing.T) *action.Configuration {
	t.Helper()

	registryClient, err := registry.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	return &action.Configuration{
		Releases:       storage.Init(driver.NewMemory()),
		KubeClient:     &fake.FailingKubeClient{PrintingKubeClient: fake.PrintingKubeClient{Out: io.Discard}},
		Capabilities:   chartutil.DefaultCapabilities,
		RegistryClient: registryClient,
		Log: func(format string, v ...interface{}) {
			t.Helper()
			t.Logf(format, v...)
		},
	}
}

type testInstaller struct {
	fn func(ctx context.Context, client *action.Install, chart string, values map[string]interface{}) (*release.Release, error)
}

func (t testInstaller) install(ctx context.Context, client *action.Install, chart string, values map[string]interface{}) (*release.Release, error) {
	return t.fn(ctx, client, chart, values)
}

func TestInstall(t *testing.T) {
	tests := map[string]struct {
		errExpected     error
		installRunnerFn func(*testing.T, *InstallConfig) testInstaller
	}{

		"passes helm client": {
			installRunnerFn: func(t *testing.T, cfg *InstallConfig) testInstaller {
				return testInstaller{
					fn: func(ctx context.Context, client *action.Install, chartName string, values map[string]interface{}) (*release.Release, error) {
						assert.NotNil(t, client)
						return &release.Release{Name: client.ReleaseName}, nil
					},
				}
			},
		},

		"sets values on client": {
			installRunnerFn: func(t *testing.T, cfg *InstallConfig) testInstaller {
				return testInstaller{
					fn: func(ctx context.Context, client *action.Install, chartName string, values map[string]interface{}) (*release.Release, error) {
						assert.Equal(t, cfg.Chart.URL, client.RepoURL)
						assert.Equal(t, cfg.Chart.Version, client.Version)
						assert.Equal(t, cfg.ReleaseName, client.ReleaseName)
						assert.Equal(t, cfg.Atomic, client.Atomic)
						assert.Equal(t, cfg.CreateNamespace, client.CreateNamespace)
						return &release.Release{Name: cfg.ReleaseName}, nil
					},
				}
			},
		},

		"wraps errors": {
			errExpected: errOops,
			installRunnerFn: func(t *testing.T, cfg *InstallConfig) testInstaller {
				return testInstaller{
					fn: func(ctx context.Context, client *action.Install, chartName string, values map[string]interface{}) (*release.Release, error) {
						return nil, errOops
					},
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			i := &installer{}
			cfg := generics.GetFakeObj[InstallConfig]()
			cfg.Kubeconfig = &rest.Config{}

			if test.installRunnerFn != nil {
				cfg.installer = test.installRunnerFn(t, &cfg)
			}

			resp, err := i.Install(context.Background(), &cfg)
			if test.errExpected != nil {
				assert.ErrorIs(t, err, test.errExpected)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}
