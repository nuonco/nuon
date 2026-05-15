package activities

import (
	"context"
	"fmt"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
)

type TerminateOrgWorkflowsRequest struct {
	OrgID string `json:"org_id" validate:"required"`
}

type TerminateOrgWorkflowsResponse struct {
	Terminated int      `json:"terminated"`
	Errors     []string `json:"errors,omitempty"`
}

var orgWorkflowNamespaces = []string{
	"general",
	"installs",
	"runners",
	"orgs",
	"components",
	"apps",
	"actions",
	"vcs",
	"onboardings",
}

// @temporal-gen-v2 activity
// @by-field OrgID
// @schedule-to-close-timeout 300s
// @start-to-close-timeout 300s
func (a *Activities) TerminateOrgWorkflows(ctx context.Context, req TerminateOrgWorkflowsRequest) (*TerminateOrgWorkflowsResponse, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("org_id", req.OrgID))

	resp := &TerminateOrgWorkflowsResponse{}

	for _, ns := range orgWorkflowNamespaces {
		client, err := a.tClient.GetNamespaceClient(ns)
		if err != nil {
			l.Warn("unable to get namespace client", zap.String("namespace", ns), zap.Error(err))
			continue
		}

		var token []byte
		for {
			wfs, err := client.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
				Namespace:     ns,
				PageSize:      100,
				NextPageToken: token,
				Query:         "ExecutionStatus='Running'",
			})
			if err != nil {
				l.Warn("unable to list workflows", zap.String("namespace", ns), zap.Error(err))
				break
			}

			for _, wf := range wfs.Executions {
				orgID := memoStringValue(wf.Memo, "org-id")
				ownerID := memoStringValue(wf.Memo, "owner-id")
				if orgID != req.OrgID && ownerID != req.OrgID {
					continue
				}

				l.Info("terminating workflow",
					zap.String("namespace", ns),
					zap.String("workflow_id", wf.Execution.WorkflowId),
				)

				if err := client.TerminateWorkflow(ctx, wf.Execution.WorkflowId, wf.Execution.RunId, fmt.Sprintf("admin org cleanup for %s", req.OrgID)); err != nil {
					l.Warn("unable to terminate workflow",
						zap.String("workflow_id", wf.Execution.WorkflowId),
						zap.Error(err))
					resp.Errors = append(resp.Errors, wf.Execution.WorkflowId)
					continue
				}
				resp.Terminated++
			}

			token = wfs.NextPageToken
			if len(token) == 0 {
				break
			}
		}
	}

	l.Info("org workflow termination complete",
		zap.Int("terminated", resp.Terminated),
		zap.Int("errors", len(resp.Errors)))

	return resp, nil
}

// memoStringValue extracts a string value from a Temporal workflow Memo by key.
func memoStringValue(memo *commonpb.Memo, key string) string {
	if memo == nil {
		return ""
	}
	payload, ok := memo.Fields[key]
	if !ok {
		return ""
	}
	var val string
	if err := converter.GetDefaultDataConverter().FromPayload(payload, &val); err != nil {
		return ""
	}
	return val
}
