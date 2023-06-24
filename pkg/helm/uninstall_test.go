package helm

import (
	"context"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	"k8s.io/client-go/rest"
)

// TODO(jdt): this isn't testing much currently!
func Test_uninstall(t *testing.T) {
	tests := map[string]struct {
		config      *UninstallConfig
		expected    *release.UninstallReleaseResponse
		errExpected error
	}{
		"doesn't fail if chart isn't loaded": {},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			h := helmUninstaller{}
			cfg := actionConfigFixture(t)
			c := action.NewUninstall(cfg)

			resp, err := h.uninstall(context.Background(), c, "test")
			if test.errExpected != nil {
				assert.ErrorIs(t, err, test.errExpected)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected, resp)
		})
	}
}

type testUninstaller struct {
	fn func(ctx context.Context, client *action.Uninstall, name string) (*release.UninstallReleaseResponse, error)
}

func (t testUninstaller) uninstall(ctx context.Context, client *action.Uninstall, releaseName string) (*release.UninstallReleaseResponse, error) {
	return t.fn(ctx, client, releaseName)
}

func TestUninstall(t *testing.T) {
	tests := map[string]struct {
		errExpected       error
		uninstallRunnerFn func(*testing.T, *UninstallConfig) testUninstaller
	}{
		"passes helm client": {
			uninstallRunnerFn: func(t *testing.T, cfg *UninstallConfig) testUninstaller {
				return testUninstaller{
					fn: func(ctx context.Context, client *action.Uninstall, name string) (*release.UninstallReleaseResponse, error) {
						assert.NotNil(t, client)
						return &release.UninstallReleaseResponse{}, nil
					},
				}
			},
		},

		"passes release name": {
			uninstallRunnerFn: func(t *testing.T, cfg *UninstallConfig) testUninstaller {
				return testUninstaller{
					fn: func(ctx context.Context, client *action.Uninstall, name string) (*release.UninstallReleaseResponse, error) {
						assert.Equal(t, cfg.ReleaseName, name)
						return &release.UninstallReleaseResponse{}, nil
					},
				}
			},
		},

		"wraps errors": {
			errExpected: errOops,
			uninstallRunnerFn: func(t *testing.T, cfg *UninstallConfig) testUninstaller {
				return testUninstaller{
					fn: func(ctx context.Context, client *action.Uninstall, name string) (*release.UninstallReleaseResponse, error) {
						return nil, errOops
					},
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			u := &uninstaller{}
			cfg := generics.GetFakeObj[UninstallConfig]()
			cfg.Kubeconfig = &rest.Config{}

			if test.uninstallRunnerFn != nil {
				cfg.uninstaller = test.uninstallRunnerFn(t, &cfg)
			}
			resp, err := u.Uninstall(context.Background(), &cfg)
			if test.errExpected != nil {
				assert.ErrorIs(t, err, test.errExpected)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}
