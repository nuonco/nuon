package activities

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/stretchr/testify/assert"
)

func Test_AddConnectionsToPlan(t *testing.T) {
	componentID := domains.NewComponentID()
	installID := domains.NewInstallID()
	instance1 := generics.GetFakeObj[models.Instance]()
	instance1.ComponentID = componentID
	instance1.Deploys = []*models.Deploy{generics.GetFakeObj[*models.Deploy]()}

	instance2 := generics.GetFakeObj[models.Instance]()
	instance2.Deploys = []*models.Deploy{generics.GetFakeObj[*models.Deploy]()}

	tests := map[string]struct {
		instanceRepo          func(*gomock.Controller) *repos.MockInstanceRepo
		expectedConnectionIDs []string
		errExpected           error
	}{
		"no other components": {
			instanceRepo: func(ctl *gomock.Controller) *repos.MockInstanceRepo {
				connections := []*models.Instance{}
				mockRepo := repos.NewMockInstanceRepo(ctl)

				mockRepo.EXPECT().ListByInstall(gomock.Any(), installID).Return(connections, nil)
				return mockRepo
			},
			expectedConnectionIDs: []string{},
		},
		"only self, no connections": {
			instanceRepo: func(ctl *gomock.Controller) *repos.MockInstanceRepo {
				connections := []*models.Instance{&instance1}
				mockRepo := repos.NewMockInstanceRepo(ctl)

				mockRepo.EXPECT().ListByInstall(gomock.Any(), installID).Return(connections, nil)
				return mockRepo
			},
			expectedConnectionIDs: []string{},
		},
		"one other connected component": {
			instanceRepo: func(ctl *gomock.Controller) *repos.MockInstanceRepo {
				connections := []*models.Instance{&instance1, &instance2}
				mockRepo := repos.NewMockInstanceRepo(ctl)

				mockRepo.EXPECT().ListByInstall(gomock.Any(), installID).Return(connections, nil)
				return mockRepo
			},
			expectedConnectionIDs: []string{instance2.ComponentID},
		},
		"repo error": {
			instanceRepo: func(ctl *gomock.Controller) *repos.MockInstanceRepo {
				mockRepo := repos.NewMockInstanceRepo(ctl)
				mockRepo.EXPECT().ListByInstall(gomock.Any(), installID).Return(nil, assert.AnError)
				return mockRepo
			},
			errExpected: assert.AnError,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			act := &activities{instanceRepo: test.instanceRepo(mockCtl)}

			actual, err := act.AddConnectionsToPlan(context.Background(), componentID, installID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			actualIDs := []string{}
			for _, con := range actual.Instances {
				actualIDs = append(actualIDs, con.ComponentId)
				assert.Equal(t, instance2.InstallID, con.InstallId)
				assert.Equal(t, instance2.Deploys[0].ID, con.DeployId)
				assert.Equal(t, instance2.ComponentID, con.ComponentId)
				assert.Equal(t, instance2.Component.Name, con.ComponentName)
			}
			assert.NoError(t, err)
			assert.EqualValues(t, test.expectedConnectionIDs, actualIDs)
		})
	}
}
