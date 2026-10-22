package installs

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

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
		svc := newContentsService(step, helmContents)

		out := captureStdout(t, func() {
			err := svc.WorkflowStepContents(context.Background(), "wf1", "step1", true)
			require.NoError(t, err)
		})
		require.Contains(t, out, `"helm_content_diff"`)
		require.Contains(t, out, `"my-app"`)
	})

	t.Run("diff json", func(t *testing.T) {
		svc := newContentsService(step, helmContents)

		out := captureStdout(t, func() {
			err := svc.WorkflowStepDiff(context.Background(), "wf1", "step1", true)
			require.NoError(t, err)
		})
		require.Contains(t, out, `"plan_type":"helm_approval"`)
		require.Contains(t, out, `"release":"my-app"`)
		require.Contains(t, out, `"add":1`)
	})

	t.Run("diff table", func(t *testing.T) {
		svc := newContentsService(step, helmContents)

		out := captureStdout(t, func() {
			err := svc.WorkflowStepDiff(context.Background(), "wf1", "step1", false)
			require.NoError(t, err)
		})
		require.Contains(t, out, "Deployment")
		require.Contains(t, out, "my-app")
	})

	t.Run("missing approval", func(t *testing.T) {
		svc := newContentsService(&models.AppWorkflowStep{ID: "step1"}, nil)

		err := svc.WorkflowStepContents(context.Background(), "wf1", "step1", true)
		require.ErrorContains(t, err, "has no approval contents")
	})

	t.Run("unsupported diff", func(t *testing.T) {
		svc := newContentsService(step, map[string]any{"installs": []any{}})

		err := svc.WorkflowStepDiff(context.Background(), "wf1", "step1", true)
		require.ErrorContains(t, err, "steps contents")
	})
}

type contentsTestClient struct {
	nuon.Client
	step     *models.AppWorkflowStep
	contents any
}

func (c contentsTestClient) GetWorkflowStep(context.Context, string, string) (*models.AppWorkflowStep, error) {
	return c.step, nil
}

func (c contentsTestClient) GetWorkflowStepApprovalContents(context.Context, string, string, string) (any, error) {
	return c.contents, nil
}

func newContentsService(step *models.AppWorkflowStep, contents any) *Service {
	return New(contentsTestClient{step: step, contents: contents}, &config.Config{})
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
