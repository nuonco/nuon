package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestAdminService_UpsertSandboxVersion(t *testing.T) {
	errUpsertSandboxVersion := fmt.Errorf("failed to upsert sandbox version")
	sandboxVersion := generics.GetFakeObj[*models.SandboxVersion]()

	tests := map[string]struct {
		inputFn     func() models.SandboxVersionInput
		repoFn      func(*gomock.Controller) *repos.MockAdminRepo
		errExpected error
	}{
		"create a new sandbox version": {
			inputFn: func() models.SandboxVersionInput {
				inp := generics.GetFakeObj[models.SandboxVersionInput]()
				inp.ID = gomock.Nil().String()
				inp.SandboxName = "test-sandbox"
				inp.SandboxVersion = "1.0"
				inp.TfVersion = "1.0"
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockAdminRepo {
				repo := repos.NewMockAdminRepo(ctl)
				repo.EXPECT().UpsertSandboxVersion(gomock.Any(), gomock.Any()).Return(sandboxVersion, nil)
				return repo
			},
		},
		"upsert not found": {
			inputFn: func() models.SandboxVersionInput {
				inp := generics.GetFakeObj[models.SandboxVersionInput]()
				inp.ID = sandboxVersion.ID.String()
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockAdminRepo {
				repo := repos.NewMockAdminRepo(ctl)
				repo.EXPECT().UpsertSandboxVersion(gomock.Any(), gomock.Any()).Return(nil, errUpsertSandboxVersion)
				return repo
			},
			errExpected: errUpsertSandboxVersion,
		},
		"upsert happy path": {
			inputFn: func() models.SandboxVersionInput {
				inp := generics.GetFakeObj[models.SandboxVersionInput]()
				inp.ID = sandboxVersion.ID.String()
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockAdminRepo {
				repo := repos.NewMockAdminRepo(ctl)
				repo.EXPECT().UpsertSandboxVersion(gomock.Any(), gomock.Any()).Return(sandboxVersion, nil)
				return repo
			},
			errExpected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			sandboxVersionInput := test.inputFn()
			repo := test.repoFn(mockCtl)
			svc := &adminService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}

			returnedSandbox, err := svc.UpsertSandboxVersion(context.Background(), sandboxVersionInput)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedSandbox)
		})
	}
}
