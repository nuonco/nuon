package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestInstanceService_GetInstancesByInstall(t *testing.T) {
	errGetInstancesByInstall := fmt.Errorf("error getting install's instances")
	instance := generics.GetFakeObj[*models.Instance]()
	installID := instance.InstallID

	tests := map[string]struct {
		installID   string
		repoFn      func(*gomock.Controller) *repos.MockInstanceRepo
		errExpected error
		assertFn    func(*testing.T, *models.Instance)
	}{
		"happy path": {
			installID: installID,
			repoFn: func(ctl *gomock.Controller) *repos.MockInstanceRepo {
				repo := repos.NewMockInstanceRepo(ctl)
				repo.EXPECT().ListByInstall(gomock.Any(), installID).Return([]*models.Instance{instance}, nil)
				return repo
			},
		},
		"error": {
			installID: installID,
			repoFn: func(ctl *gomock.Controller) *repos.MockInstanceRepo {
				repo := repos.NewMockInstanceRepo(ctl)
				repo.EXPECT().ListByInstall(gomock.Any(), installID).Return(nil, errGetInstancesByInstall)
				return repo
			},
			errExpected: errGetInstancesByInstall,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &instanceService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}
			instances, err := svc.GetInstancesByInstall(context.Background(), test.installID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, instance, instances[0])
		})
	}
}
