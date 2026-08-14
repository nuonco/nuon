package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// configConnectionIDFor returns the seeded app config's connection for a
// component, which a ComponentBuild needs as its parent.
func (s *InstallsServiceTestSuite) configConnectionIDFor(componentID string) string {
	for _, ccc := range s.testAppConfig.ComponentConfigConnections {
		if ccc.ComponentID == componentID {
			return ccc.ID
		}
	}
	s.T().Fatalf("no component config connection for component %s", componentID)
	return ""
}

// seedDeployedHelmComponent gives a helm component the deploy history a recovery
// needs: the plan reads the release name, namespace and storage driver off the
// build the component was last deployed with.
func (s *InstallsServiceTestSuite) seedDeployedHelmComponent() (*app.Install, *app.Component, *app.InstallDeploy) {
	install := s.createTestInstall()
	helmComp := s.getSeededComponent(app.ComponentTypeHelmChart)
	installComp := s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, helmComp.ID)
	build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), s.configConnectionIDFor(helmComp.ID))
	deploy := s.deps.Seeder.CreateInstallDeploy(s.ctx, s.T(), installComp.ID, build.ID)

	return install, helmComp, deploy
}

func (s *InstallsServiceTestSuite) recoverPath(installID, componentID string) string {
	return fmt.Sprintf("/v1/installs/%s/components/%s/recover-helm-release", installID, componentID)
}

func (s *InstallsServiceTestSuite) TestRecoverInstallComponentHelmReleaseSuccess() {
	install, helmComp, _ := s.seedDeployedHelmComponent()

	rr := s.makeRequest(http.MethodPost, s.recoverPath(install.ID, helmComp.ID), nil)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var found bool
	for _, c := range tests.GetQueueSignals(s.T(), s.deps.DB) {
		if string(c.Type) == "execute-workflow" {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "expected the workflow to be enqueued")

	// The recovery gets its own deploy row so it has a log stream and an audit
	// trail, and it must be typed as a recovery rather than an apply.
	var recovery app.InstallDeploy
	res := s.deps.DB.
		Where(app.InstallDeploy{Type: app.InstallDeployTypeRecover}).
		Order("created_at DESC").
		First(&recovery)
	require.NoError(s.T(), res.Error)
	assert.Equal(s.T(), app.InstallDeployStatusQueued, recovery.Status)
}

// A recovery targets a Helm release, so there is nothing to do for a component
// that does not have one.
func (s *InstallsServiceTestSuite) TestRecoverInstallComponentHelmReleaseRejectsNonHelmComponent() {
	install := s.createTestInstall()
	tfComp := s.getSeededComponent(app.ComponentTypeTerraformModule)
	installComp := s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, tfComp.ID)
	build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), s.configConnectionIDFor(tfComp.ID))
	s.deps.Seeder.CreateInstallDeploy(s.ctx, s.T(), installComp.ID, build.ID)

	rr := s.makeRequest(http.MethodPost, s.recoverPath(install.ID, tfComp.ID), nil)
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

// Nothing was ever released, so there is no pending release to recover — telling
// the caller to deploy is more useful than starting a workflow that no-ops.
func (s *InstallsServiceTestSuite) TestRecoverInstallComponentHelmReleaseRejectsNeverDeployed() {
	install := s.createTestInstall()
	helmComp := s.getSeededComponent(app.ComponentTypeHelmChart)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, helmComp.ID)

	rr := s.makeRequest(http.MethodPost, s.recoverPath(install.ID, helmComp.ID), nil)
	assert.Equal(s.T(), http.StatusConflict, rr.Code)
}

// The safety guard: rolling a release back while Helm is genuinely mid-operation
// can corrupt it, and nothing else serializes this against a live deploy.
func (s *InstallsServiceTestSuite) TestRecoverInstallComponentHelmReleaseRejectsWhileJobRunning() {
	install, helmComp, deploy := s.seedDeployedHelmComponent()

	job := s.deps.Seeder.CreateRunnerJob(s.ctx, s.T(), deploy.ID, "install_deploys")
	require.NoError(s.T(), s.deps.DB.Model(job).Update("status", app.RunnerJobStatusInProgress).Error)

	rr := s.makeRequest(http.MethodPost, s.recoverPath(install.ID, helmComp.ID), nil)
	assert.Equal(s.T(), http.StatusConflict, rr.Code)
}

// A finished job must not block recovery — a deploy that failed and left the
// release stuck is the whole reason this endpoint exists.
func (s *InstallsServiceTestSuite) TestRecoverInstallComponentHelmReleaseAllowedAfterFailedJob() {
	install, helmComp, deploy := s.seedDeployedHelmComponent()

	job := s.deps.Seeder.CreateRunnerJob(s.ctx, s.T(), deploy.ID, "install_deploys")
	require.NoError(s.T(), s.deps.DB.Model(job).Update("status", app.RunnerJobStatusFailed).Error)

	rr := s.makeRequest(http.MethodPost, s.recoverPath(install.ID, helmComp.ID), nil)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	assert.Equal(s.T(), http.StatusCreated, rr.Code)
}

func (s *InstallsServiceTestSuite) TestRecoverInstallComponentHelmReleaseComponentNotFound() {
	install := s.createTestInstall()

	rr := s.makeRequest(http.MethodPost, s.recoverPath(install.ID, "cmp_nonexistent_00000000"), nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
