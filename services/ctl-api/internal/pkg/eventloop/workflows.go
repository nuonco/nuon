package eventloop

import (
	"context"
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
)

func (e *evClient) GetWorkflowStatus(ctx context.Context, namespace string, workflowId string) (enumsv1.WorkflowExecutionStatus, error) {
	// get workflow
	nsClient, err := e.client.GetNamespaceClient(namespace)
	if err != nil {
		return enumsv1.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED, err
	}

	// get running executions: expected to be <= 1
	request := workflowservice.ListWorkflowExecutionsRequest{
		Namespace: namespace,
		Query:     fmt.Sprintf("WorkflowId = '%s' AND ExecutionStatus='Running'", workflowId),
	}

	workflows, err := nsClient.ListWorkflow(
		ctx, &request,
	)
	if err != nil {
		e.l.Error(fmt.Sprintf("[evClient.GetWorkflowStatus] %v", err))
		return enumsv1.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED, err
	}

	if len(workflows.Executions) == 0 {
		return enumsv1.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED, nil
	}
	// type is WorkflowExecutionInfo
	workflow := workflows.Executions[0]
	return workflow.Status, nil
}

func (e evClient) GetWorkflowCount(ctx context.Context, namespace string, workflowId string) (int64, error) {
	nsClient, err := e.client.GetNamespaceClient(namespace)
	// get total count of executions: use value in metrics
	wfcRequest := workflowservice.CountWorkflowExecutionsRequest{
		Namespace: namespace,
	}
	wfCount, err := nsClient.CountWorkflow(ctx, &wfcRequest)
	if err != nil {
		e.l.Error(fmt.Sprintf("[evClient.GetWorkflowCount] %v", err))
		return 0, err
	}
	return wfCount.Count, nil
}
