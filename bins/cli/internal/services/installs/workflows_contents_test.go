package installs

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestWorkflowStepContentsAndDiff(t *testing.T) {
	helmContents := map[string]any{
		"op": "install",
		"helm_content_diff": []any{
			map[string]any{
				"kind":      "Deployment",
				"name":      "my-app",
				"namespace": "default",
				"entries": []any{
					map[string]any{"type": 1, "path": "spec.replicas", "applied": "3"},
				},
			},
		},
	}
	step := &models.AppWorkflowStep{
		ID:       "step1",
		Approval: &models.AppWorkflowStepApproval{ID: "appr1"},
	}

	t.Run("contents json", func(t *testing.T) {
		svc, api := newContentsService(t)
		api.EXPECT().GetWorkflowStep(gomock.Any(), "wf1", "step1").Return(step, nil)
		api.EXPECT().GetWorkflowStepApprovalContents(gomock.Any(), "wf1", "step1", "appr1").Return(helmContents, nil)

		out := captureStdout(t, func() {
			err := svc.WorkflowStepContents(context.Background(), "wf1", "step1", true)
			require.NoError(t, err)
		})
		require.Contains(t, out, `"helm_content_diff"`)
		require.Contains(t, out, `"my-app"`)
	})

	t.Run("diff json", func(t *testing.T) {
		svc, api := newContentsService(t)
		api.EXPECT().GetWorkflowStep(gomock.Any(), "wf1", "step1").Return(step, nil)
		api.EXPECT().GetWorkflowStepApprovalContents(gomock.Any(), "wf1", "step1", "appr1").Return(helmContents, nil)

		out := captureStdout(t, func() {
			err := svc.WorkflowStepDiff(context.Background(), "wf1", "step1", true)
			require.NoError(t, err)
		})
		require.Contains(t, out, `"plan_type":"helm_approval"`)
		require.Contains(t, out, `"release":"my-app"`)
		require.Contains(t, out, `"add":1`)
	})

	t.Run("diff table", func(t *testing.T) {
		svc, api := newContentsService(t)
		api.EXPECT().GetWorkflowStep(gomock.Any(), "wf1", "step1").Return(step, nil)
		api.EXPECT().GetWorkflowStepApprovalContents(gomock.Any(), "wf1", "step1", "appr1").Return(helmContents, nil)

		out := captureStdout(t, func() {
			err := svc.WorkflowStepDiff(context.Background(), "wf1", "step1", false)
			require.NoError(t, err)
		})
		require.Contains(t, out, "Deployment")
		require.Contains(t, out, "my-app")
	})

	t.Run("missing approval", func(t *testing.T) {
		svc, api := newContentsService(t)
		api.EXPECT().GetWorkflowStep(gomock.Any(), "wf1", "step1").Return(&models.AppWorkflowStep{ID: "step1"}, nil)

		err := svc.WorkflowStepContents(context.Background(), "wf1", "step1", true)
		require.ErrorContains(t, err, "has no approval contents")
	})

	t.Run("unsupported diff", func(t *testing.T) {
		svc, api := newContentsService(t)
		api.EXPECT().GetWorkflowStep(gomock.Any(), "wf1", "step1").Return(step, nil)
		api.EXPECT().GetWorkflowStepApprovalContents(gomock.Any(), "wf1", "step1", "appr1").Return(map[string]any{"installs": []any{}}, nil)

		err := svc.WorkflowStepDiff(context.Background(), "wf1", "step1", true)
		require.ErrorContains(t, err, "steps contents")
	})
}

func newContentsService(t *testing.T) (*Service, *nuon.MockClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	api := nuon.NewMockClient(ctrl)
	return New(api, &config.Config{}), api
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	original := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = original
		_ = writer.Close()
	}()

	fn()
	require.NoError(t, writer.Close())
	os.Stdout = original

	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	require.NoError(t, err)
	return buf.String()
}
