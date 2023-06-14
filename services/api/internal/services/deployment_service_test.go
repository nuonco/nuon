package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestDeploymentService_GetComponentDeployments(t *testing.T) {
	errGetComponentDeployments := fmt.Errorf("error getting component deployments")
	componentID := domains.NewComponentID()
	deployment := generics.GetFakeObj[*models.Deployment]()

	tests := map[string]struct {
		componentIDs []string
		repoFn       func(*gomock.Controller) *repos.MockDeploymentRepo
		errExpected  error
		assertFn     func(*testing.T, *models.Deployment)
	}{
		"happy path": {
			componentIDs: []string{componentID},
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				deployments := []*models.Deployment{deployment}
				repo.EXPECT().
					ListByComponents(gomock.Any(), []string{componentID}, &models.ConnectionOptions{}).
					Return(deployments, nil, nil)
				return repo
			},
		},
		"error": {
			componentIDs: []string{componentID},
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				repo.EXPECT().
					ListByComponents(gomock.Any(), []string{componentID}, &models.ConnectionOptions{}).
					Return(nil, nil, errGetComponentDeployments)
				return repo
			},
			errExpected: errGetComponentDeployments,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &deploymentService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}
			deployments, _, err := svc.GetComponentDeployments(context.Background(), test.componentIDs, &models.ConnectionOptions{})
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, deployment, deployments[0])
		})
	}
}

func TestDeploymentService_GetAppDeployments(t *testing.T) {
	errGetAppDeployments := fmt.Errorf("error getting app deployments")
	appID := domains.NewAppID()
	deployment := generics.GetFakeObj[*models.Deployment]()

	tests := map[string]struct {
		appIDs      []string
		repoFn      func(*gomock.Controller) *repos.MockDeploymentRepo
		errExpected error
		assertFn    func(*testing.T, *models.Deployment)
	}{
		"happy path": {
			appIDs: []string{appID},
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				deployments := []*models.Deployment{deployment}
				repo.EXPECT().
					ListByApps(gomock.Any(), []string{appID}, &models.ConnectionOptions{}).
					Return(deployments, nil, nil)
				return repo
			},
		},
		"error": {
			appIDs: []string{appID},
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				repo.EXPECT().
					ListByApps(gomock.Any(), []string{appID}, &models.ConnectionOptions{}).
					Return(nil, nil, errGetAppDeployments)
				return repo
			},
			errExpected: errGetAppDeployments,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &deploymentService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}
			deployments, _, err := svc.GetAppDeployments(context.Background(), test.appIDs, &models.ConnectionOptions{})
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, deployment, deployments[0])
		})
	}
}

func TestDeploymentService_GetInstallDeployments(t *testing.T) {
	errGetInstallDeployments := fmt.Errorf("error getting install deployments")
	installID := domains.NewInstallID()
	deployment := generics.GetFakeObj[*models.Deployment]()

	tests := map[string]struct {
		installIDs  []string
		repoFn      func(*gomock.Controller) *repos.MockDeploymentRepo
		errExpected error
		assertFn    func(*testing.T, *models.Deployment)
	}{
		"happy path": {
			installIDs: []string{installID},
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				deployments := []*models.Deployment{deployment}
				repo.EXPECT().
					ListByInstalls(gomock.Any(), []string{installID}, &models.ConnectionOptions{}).
					Return(deployments, nil, nil)
				return repo
			},
		},
		"error": {
			installIDs: []string{installID},
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				repo.EXPECT().
					ListByInstalls(gomock.Any(), []string{installID}, &models.ConnectionOptions{}).
					Return(nil, nil, errGetInstallDeployments)
				return repo
			},
			errExpected: errGetInstallDeployments,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &deploymentService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}
			deployments, _, err := svc.GetInstallDeployments(context.Background(), test.installIDs, &models.ConnectionOptions{})
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, deployment, deployments[0])
		})
	}
}

func TestDeploymentService_GetDeployment(t *testing.T) {
	errGetDeployment := fmt.Errorf("error getting app")
	deploymentID := domains.NewDeploymentID()
	app := generics.GetFakeObj[*models.Deployment]()

	tests := map[string]struct {
		deploymentID string
		repoFn       func(*gomock.Controller) *repos.MockDeploymentRepo
		errExpected  error
		assertFn     func(*testing.T, *models.Deployment)
	}{
		"happy path": {
			deploymentID: deploymentID,
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), deploymentID).Return(app, nil)
				return repo
			},
		},
		"error": {
			deploymentID: deploymentID,
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), deploymentID).Return(nil, errGetDeployment)
				return repo
			},
			errExpected: errGetDeployment,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &deploymentService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}
			returnedDeployment, err := svc.GetDeployment(context.Background(), test.deploymentID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedDeployment)
		})
	}
}
