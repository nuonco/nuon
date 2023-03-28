package installs

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/generics"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/repos/temporal"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/repos/workflows"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func Test_commands_Deprovision(t *testing.T) {
	installID := shortid.New()

	tests := map[string]struct {
		repos       func(*testing.T, *gomock.Controller) (workflows.Repo, temporal.Repo)
		errExpected error
	}{
		"happy path": {
			repos: func(t *testing.T, mockCtl *gomock.Controller) (workflows.Repo, temporal.Repo) {
				pReq := generics.GetFakeObj[*installsv1.ProvisionRequest]()

				wkflowsRepo := workflows.NewMockRepo(mockCtl)
				wkflowsRepo.EXPECT().GetInstallProvisionRequest(gomock.Any(), installID).
					Return(pReq, nil)

				temporalRepo := temporal.NewMockRepo(mockCtl)
				temporalRepo.EXPECT().TriggerInstallDeprovision(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *installsv1.DeprovisionRequest) error {
					assert.Equal(t, pReq.OrgId, req.OrgId)
					assert.Equal(t, pReq.AppId, req.AppId)
					assert.Equal(t, pReq.InstallId, req.InstallId)
					assert.True(t, proto.Equal(pReq.SandboxSettings, req.SandboxSettings))
					assert.True(t, proto.Equal(pReq.AccountSettings, req.AccountSettings))
					assert.NoError(t, req.Validate())
					return nil
				})

				return wkflowsRepo, temporalRepo
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)

			workflowsRepo, temporalRepo := test.repos(t, mockCtl)
			cmd := &commands{
				Workflows: workflowsRepo,
				Temporal:  temporalRepo,
			}

			err := cmd.Deprovision(ctx, installID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
