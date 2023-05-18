package deleteorg

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/services/api/internal/workflows"
	"github.com/stretchr/testify/assert"
)

func Test_ActivityTriggerOrgJob(t *testing.T) {
	err := errors.New("error")
	orgID, _ := shortid.NewNanoID("org")

	tests := map[string]struct {
		mockMgr     func(*gomock.Controller) *workflows.MockOrgWorkflowManager
		errExpected error
	}{
		"happy path": {
			mockMgr: func(ctl *gomock.Controller) *workflows.MockOrgWorkflowManager {
				wkflowmgr := workflows.NewMockOrgWorkflowManager(ctl)
				wkflowmgr.EXPECT().Deprovision(gomock.Any(), orgID).Return("12345", nil)
				return wkflowmgr
			},
		},
		"mgr err": {
			mockMgr: func(ctl *gomock.Controller) *workflows.MockOrgWorkflowManager {
				wkflowmgr := workflows.NewMockOrgWorkflowManager(ctl)
				wkflowmgr.EXPECT().Deprovision(gomock.Any(), orgID).Return("12345", err)
				return wkflowmgr
			},
			errExpected: err,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			act := &activities{
				mgr: test.mockMgr(mockCtl),
			}

			_, err := act.TriggerOrgDeprovision(context.Background(), orgID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
