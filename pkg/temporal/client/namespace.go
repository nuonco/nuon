package temporal

import (
	"context"
	"fmt"

	enumspb "go.temporal.io/api/enums/v1"

	"go.temporal.io/api/workflowservice/v1"
	tclient "go.temporal.io/sdk/client"
	converter "go.temporal.io/sdk/converter"
	"go.uber.org/zap"

	"github.com/pkg/errors"
)

func (t *temporal) GetNamespaceClient(namespace string) (tclient.Client, error) {
	defaultClient, err := t.getClient()
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	opts := t.getOpts()
	opts.Namespace = namespace

	client, err := tclient.NewClientFromExisting(defaultClient, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to get client in namespace %s: %w", namespace, err)
	}

	return client, nil
}

func (t *temporal) ExecuteWorkflowInNamespace(ctx context.Context,
	namespace string,
	options tclient.StartWorkflowOptions,
	workflow interface{},
	args ...interface{},
) (tclient.WorkflowRun, error) {
	client, err := t.GetNamespaceClient(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get namespace client")
	}

	return client.ExecuteWorkflow(ctx, options, workflow, args...)
}

func (t *temporal) GetWorkflowStatusInNamespace(ctx context.Context,
	namespace string,
	workflowID string,
	runID string,
) (enumspb.WorkflowExecutionStatus, error) {
	client, err := t.GetNamespaceClient(namespace)
	if err != nil {
		return enumspb.WorkflowExecutionStatus(0), errors.Wrap(err, "unable to get namespace client")
	}

	exec, err := client.DescribeWorkflowExecution(ctx, workflowID, runID)
	if err != nil {
		return enumspb.WorkflowExecutionStatus(0), errors.Wrap(err, "unable to get execution")
	}

	return exec.WorkflowExecutionInfo.Status, nil
}

func (t *temporal) DescribeWorkflowExecutionInNamespace(ctx context.Context,
	namespace string,
	workflowID string,
	runID string,
) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	client, err := t.GetNamespaceClient(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get namespace client")
	}

	return client.DescribeWorkflowExecution(ctx, workflowID, runID)
}

func (t *temporal) GetWorkflowInNamespace(ctx context.Context,
	namespace string,
	workflowID string,
	runID string,
) (tclient.WorkflowRun, error) {
	client, err := t.GetNamespaceClient(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get namespace client")
	}

	return client.GetWorkflow(ctx, workflowID, runID), nil
}

func (t *temporal) CancelWorkflowInNamespace(ctx context.Context,
	namespace string,
	workflowID string,
	runID string,
) error {
	client, err := t.GetNamespaceClient(namespace)
	if err != nil {
		return errors.Wrap(err, "unable to get namespace client")
	}

	if err := client.CancelWorkflow(ctx, workflowID, runID); err != nil {
		return fmt.Errorf("unable to cancel workflow: %w", err)
	}

	return nil
}

func (t *temporal) SignalWorkflowInNamespace(ctx context.Context,
	namespace string,
	workflowID string,
	runID string,
	signalName string,
	signalArg interface{},
) error {
	client, err := t.GetNamespaceClient(namespace)
	if err != nil {
		t.Logger.Error("unable to get namespace client", zap.Error(err))
		return nil
	}

	return client.SignalWorkflow(ctx,
		workflowID,
		runID,
		signalName,
		signalArg)
}

func (t *temporal) SignalWithStartWorkflowInNamespace(ctx context.Context,
	namespace string,
	workflowID string,
	signalName string,
	signalArg interface{},
	options tclient.StartWorkflowOptions,
	workflow interface{},
	workflowArgs interface{},
) (tclient.WorkflowRun, error) {
	client, err := t.GetNamespaceClient(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get namespace client")
	}

	run, err := client.SignalWithStartWorkflow(ctx,
		workflowID,
		signalName,
		signalArg,
		options,
		workflow,
		workflowArgs)
	return run, err
}

func (t *temporal) UpdateWorkflowInNamespace(ctx context.Context,
	namespace string,
	opts tclient.UpdateWorkflowOptions,
) (tclient.WorkflowUpdateHandle, error) {
	client, err := t.GetNamespaceClient(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get namespace client")
	}

	return client.UpdateWorkflow(ctx,
		opts)
}

func (t *temporal) UpdateWithStartWorkflowInNamespace(ctx context.Context,
	namespace string,
	opts tclient.UpdateWithStartWorkflowOptions,
) (tclient.WorkflowUpdateHandle, error) {
	client, err := t.GetNamespaceClient(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get namespace client")
	}

	return client.UpdateWithStartWorkflow(ctx, opts)
}

func (t *temporal) GetWorkflowUpdateHandleInNamespace(
	namespace string,
	ref tclient.GetWorkflowUpdateHandleOptions,
) tclient.WorkflowUpdateHandle {
	client, err := t.GetNamespaceClient(namespace)
	if err != nil {
		t.Logger.Error("unable to get namespace client", zap.String("namespace", namespace), zap.Error(err))
		return nil
	}

	return client.GetWorkflowUpdateHandle(ref)
}

func (t *temporal) QueryWorkflowInNamespace(ctx context.Context,
	namespace string,
	workflowID string,
	workflowRunID string,
	queryType string,
	args ...any,
) (converter.EncodedValue, error) {
	client, err := t.GetNamespaceClient(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get namespace client")
	}

	return client.QueryWorkflow(ctx,
		workflowID,
		workflowRunID,
		queryType,
		args...)
}

func (t *temporal) QueryWorkflowWithOptionsInNamespace(ctx context.Context,
	namespace string,
	req *tclient.QueryWorkflowWithOptionsRequest,
) (*tclient.QueryWorkflowWithOptionsResponse, error) {
	client, err := t.GetNamespaceClient(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get namespace client")
	}

	return client.QueryWorkflowWithOptions(ctx,
		req)
}
