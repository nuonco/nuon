package createorg

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/stretchr/testify/assert"
)

func Test_ActivityTriggerOrgJob(t *testing.T) {
	err := errors.New("error")
	orgID, _ := shortid.NewNanoID("org")

	tests := map[string]struct {
		mockWfc     func(*gomock.Controller) workflowsclient.Client
		errExpected error
	}{
		"happy path": {
			mockWfc: func(ctl *gomock.Controller) workflowsclient.Client {
				mockWfc := workflowsclient.NewMockClient(ctl)
				mockWfc.EXPECT().TriggerOrgSignup(gomock.Any(), gomock.Any()).Return("123456", nil)
				return mockWfc
			},
		},
		"workflow err": {
			mockWfc: func(ctl *gomock.Controller) workflowsclient.Client {
				mockWfc := workflowsclient.NewMockClient(ctl)
				mockWfc.EXPECT().TriggerOrgSignup(gomock.Any(), gomock.Any()).Return("", err)
				return mockWfc
			},
			errExpected: err,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			act := &activities{
				wfc: test.mockWfc(mockCtl),
			}

			_, err := act.TriggerOrgProvision(context.Background(), orgID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
