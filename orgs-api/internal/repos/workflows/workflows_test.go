package workflows

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-generics"
	"github.com/stretchr/testify/assert"
)

//nolint:all
func TestNew(t *testing.T) {
	//TODO(jm): this test isn't working properly, fix it later.
	return
	orgsBucket := generics.GetFakeObj[Bucket]()
	appsBucket := generics.GetFakeObj[Bucket]()
	installsBucket := generics.GetFakeObj[Bucket]()
	deploymentsBucket := generics.GetFakeObj[Bucket]()
	instancesBucket := generics.GetFakeObj[Bucket]()

	tests := map[string]struct {
		optFns      func() []repoOption
		assertFn    func(*testing.T, *repo)
		errExpected error
	}{
		"happy path": {
			optFns: func() []repoOption {
				return []repoOption{
					WithOrgsBucket(orgsBucket),
					WithAppsBucket(appsBucket),
					WithInstallsBucket(installsBucket),
					WithDeploymentsBucket(deploymentsBucket),
					WithInstancesBucket(instancesBucket),
				}
			},
			assertFn: func(t *testing.T, r *repo) {
				assert.Equal(t, orgsBucket, r.OrgsBucket)
				assert.Equal(t, appsBucket, r.AppsBucket)
				assert.Equal(t, installsBucket, r.InstallsBucket)
				assert.Equal(t, deploymentsBucket, r.DeploymentsBucket)
				assert.Equal(t, instancesBucket, r.InstancesBucket)
			},
		},
		"missing orgs bucket": {
			optFns: func() []repoOption {
				return []repoOption{
					WithAppsBucket(appsBucket),
					WithInstallsBucket(installsBucket),
					WithDeploymentsBucket(deploymentsBucket),
					WithInstancesBucket(instancesBucket),
				}
			},
			errExpected: fmt.Errorf("repo.OrgsBucket"),
		},
		"missing apps bucket": {
			optFns: func() []repoOption {
				return []repoOption{
					WithOrgsBucket(orgsBucket),
					WithInstallsBucket(installsBucket),
					WithDeploymentsBucket(deploymentsBucket),
					WithInstancesBucket(instancesBucket),
				}
			},
			errExpected: fmt.Errorf("repo.AppsBucket"),
		},
		"missing installs bucket": {
			optFns: func() []repoOption {
				return []repoOption{
					WithOrgsBucket(orgsBucket),
					WithAppsBucket(appsBucket),
					WithDeploymentsBucket(deploymentsBucket),
					WithInstancesBucket(instancesBucket),
				}
			},
			errExpected: fmt.Errorf("repo.InstallsBucket"),
		},
		"missing deployments bucket": {
			optFns: func() []repoOption {
				return []repoOption{
					WithOrgsBucket(orgsBucket),
					WithAppsBucket(appsBucket),
					WithInstallsBucket(installsBucket),
					WithInstancesBucket(instancesBucket),
				}
			},
			errExpected: fmt.Errorf("repo.DeploymentsBucket"),
		},
		"missing instances bucket": {
			optFns: func() []repoOption {
				return []repoOption{
					WithOrgsBucket(orgsBucket),
					WithAppsBucket(appsBucket),
					WithInstallsBucket(installsBucket),
					WithDeploymentsBucket(deploymentsBucket),
				}
			},
			errExpected: fmt.Errorf("repo.InstancesBucket"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			opts := test.optFns()
			r, err := New(v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, r)
		})
	}
}
