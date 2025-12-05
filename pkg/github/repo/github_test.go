package github

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/kube"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	installID := "987654321"
	appKeyID := uuid.NewString()
	appKeySecretName := uuid.NewString()
	appKeySecretNamespace := uuid.NewString()
	clusterInfo := generics.GetFakeObj[*kube.ClusterInfo]()
	repo := "powertoolsdev/empty"

	tests := map[string]struct {
		optFns      func() []Option
		assertFn    func(*testing.T, *gh)
		errExpected error
	}{
		"happy path": {
			optFns: func() []Option {
				return []Option{
					WithInstallID(installID),
					WithAppKeyID(appKeyID),
					WithAppKeySecretName(appKeySecretName),
					WithAppKeySecretNamespace(appKeySecretNamespace),
					WithAppKeyClusterInfo(clusterInfo),
					WithRepo(repo),
				}
			},
			assertFn: func(t *testing.T, ecr *gh) {
				assert.Equal(t, "powertoolsdev", ecr.RepoOwner)
				assert.Equal(t, "empty", ecr.RepoName)
				assert.Equal(t, int64(987654321), ecr.InstallID)
				assert.Equal(t, appKeyID, ecr.AppKeyID)
				assert.Equal(t, appKeySecretName, ecr.AppKeySecretName)
				assert.Equal(t, appKeySecretNamespace, ecr.AppKeySecretNamespace)
				assert.Equal(t, clusterInfo, ecr.AppKeyClusterInfo)
			},
		},
		"missing install id": {
			optFns: func() []Option {
				return []Option{
					WithAppKeyID(appKeyID),
					WithAppKeySecretName(appKeySecretName),
					WithAppKeySecretNamespace(appKeySecretNamespace),
					WithAppKeyClusterInfo(clusterInfo),
					WithRepo(repo),
				}
			},
			errExpected: fmt.Errorf("InstallID"),
		},
		"missing cluster info": {
			optFns: func() []Option {
				return []Option{
					WithInstallID(installID),
					WithAppKeyID(appKeyID),
					WithAppKeySecretName(appKeySecretName),
					WithAppKeySecretNamespace(appKeySecretNamespace),
					WithRepo(repo),
				}
			},
			errExpected: fmt.Errorf("ClusterInfo"),
		},
		"invalid install id": {
			optFns: func() []Option {
				return []Option{
					WithAppKeyID(appKeyID),
					WithInstallID(uuid.NewString()),
					WithAppKeySecretName(appKeySecretName),
					WithAppKeySecretNamespace(appKeySecretNamespace),
					WithAppKeyClusterInfo(clusterInfo),
					WithRepo(repo),
				}
			},
			errExpected: fmt.Errorf("invalid github install id"),
		},
		"missing app key id": {
			optFns: func() []Option {
				return []Option{
					WithInstallID(installID),
					WithAppKeySecretName(appKeySecretName),
					WithAppKeySecretNamespace(appKeySecretNamespace),
					WithAppKeyClusterInfo(clusterInfo),
					WithRepo(repo),
				}
			},
			errExpected: fmt.Errorf("AppKeyID"),
		},
		"missing app key secret name": {
			optFns: func() []Option {
				return []Option{
					WithInstallID(installID),
					WithAppKeyID(appKeyID),
					WithAppKeySecretNamespace(appKeySecretNamespace),
					WithAppKeyClusterInfo(clusterInfo),
					WithRepo(repo),
				}
			},
			errExpected: fmt.Errorf("AppKeySecretName"),
		},
		"missing app key secret namespace": {
			optFns: func() []Option {
				return []Option{
					WithInstallID(installID),
					WithAppKeyID(appKeyID),
					WithAppKeyClusterInfo(clusterInfo),
					WithRepo(repo),
				}
			},
			errExpected: fmt.Errorf("AppKeySecretName"),
		},
		"missing repo": {
			optFns: func() []Option {
				return []Option{
					WithInstallID(installID),
					WithAppKeyID(appKeyID),
					WithAppKeySecretName(appKeySecretName),
					WithAppKeyClusterInfo(clusterInfo),
					WithAppKeySecretNamespace(appKeySecretNamespace),
				}
			},
			errExpected: fmt.Errorf("Repo"),
		},
		"invalid repo": {
			optFns: func() []Option {
				return []Option{
					WithInstallID(installID),
					WithAppKeyID(appKeyID),
					WithAppKeySecretName(appKeySecretName),
					WithAppKeyClusterInfo(clusterInfo),
					WithAppKeySecretNamespace(appKeySecretNamespace),
					WithRepo("not-a-valid-repo"),
				}
			},
			errExpected: fmt.Errorf("invalid github repo"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			opts := test.optFns()
			ecr, err := New(v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, ecr)
		})
	}
}
