package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/go-generics"
	"github.com/stretchr/testify/assert"
)

func TestDeploymentService_GetComponentDeployments(t *testing.T) {
	errGetComponentDeployments := fmt.Errorf("error getting component deployments")
	componentID := uuid.New()
	deployment := generics.GetFakeObj[*models.Deployment]()

	tests := map[string]struct {
		componentIDs []string
		repoFn       func(*gomock.Controller) *repos.MockDeploymentRepo
		errExpected  error
		assertFn     func(*testing.T, *models.Deployment)
	}{
		"happy path": {
			componentIDs: []string{componentID.String()},
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				deployments := []*models.Deployment{deployment}
				repo.EXPECT().ListByComponents(gomock.Any(), []uuid.UUID{componentID}, &models.ConnectionOptions{}).Return(deployments, nil, nil)
				return repo
			},
		},
		"invalid-id": {
			componentIDs: []string{"foo"},
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				return repo
			},
			errExpected: InvalidIDErr{},
		},
		"error": {
			componentIDs: []string{componentID.String()},
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				repo.EXPECT().ListByComponents(gomock.Any(), []uuid.UUID{componentID}, &models.ConnectionOptions{}).Return(nil, nil, errGetComponentDeployments)
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

func TestDeploymentService_GetDeployment(t *testing.T) {
	errGetDeployment := fmt.Errorf("error getting app")
	deploymentID := uuid.New()
	app := generics.GetFakeObj[*models.Deployment]()

	tests := map[string]struct {
		deploymentID string
		repoFn       func(*gomock.Controller) *repos.MockDeploymentRepo
		errExpected  error
		assertFn     func(*testing.T, *models.Deployment)
	}{
		"happy path": {
			deploymentID: deploymentID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), deploymentID).Return(app, nil)
				return repo
			},
		},
		"invalid-id": {
			deploymentID: "foo",
			repoFn: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				repo := repos.NewMockDeploymentRepo(ctl)
				return repo
			},
			errExpected: fmt.Errorf("is not a valid uuid"),
		},
		"error": {
			deploymentID: deploymentID.String(),
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
