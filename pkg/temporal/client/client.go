package temporal

import (
	"context"
	"fmt"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	tclient "go.temporal.io/sdk/client"
	sdktally "go.temporal.io/sdk/contrib/tally"
	converter "go.temporal.io/sdk/converter"

	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_client.go -source=client.go -package=temporal
type Client interface {
	tclient.Client

	GetNamespaceClient(namespace string) (tclient.Client, error)

	// ExecuteWorkflowInNamespace is a wrapper that will execute a workflow in a different namespace
	ExecuteWorkflowInNamespace(ctx context.Context,
		namespace string,
		options tclient.StartWorkflowOptions,
		workflow interface{},
		args ...interface{}) (tclient.WorkflowRun, error)

	// GetWorkflowInNamespace is a wrapper that will get a workflow in a different namespace
	GetWorkflowInNamespace(ctx context.Context,
		namespace string,
		workflowID string,
		runID string) (tclient.WorkflowRun, error)

	// DescribeWorkflowExecutionInNamespace is a wrapper that will get a workflow in a different namespace
	DescribeWorkflowExecutionInNamespace(ctx context.Context,
		namespace string,
		workflowID string,
		runID string) (*workflowservice.DescribeWorkflowExecutionResponse, error)

	// DescribeWorkflowExecutionInNamespace is a wrapper that will get a workflow in a different namespace
	GetWorkflowStatusInNamespace(ctx context.Context,
		namespace string,
		workflowID string,
		runID string) (enumspb.WorkflowExecutionStatus, error)

	// CancelWorkflowInNamespace is a wrapper that will get a workflow in a different namespace
	CancelWorkflowInNamespace(ctx context.Context,
		namespace string,
		workflowID string,
		runID string) error

	// SignalWorkflowInNamespace is a wrapper that will signal a workflow in a different namespace
	SignalWorkflowInNamespace(ctx context.Context,
		namespace string,
		workflowID string,
		runID string,
		signalName string,
		signalArg interface{}) error

	// SignalWithStartWorkflowInNamespace is a wrapper that will signal and start a workflow in a different
	// namespace
	SignalWithStartWorkflowInNamespace(ctx context.Context,
		namespace string,
		workflowID string,
		signalName string,
		signalArg interface{},
		options tclient.StartWorkflowOptions,
		workflow interface{},
		workflowArgs interface{}) (tclient.WorkflowRun, error)

	UpdateWorkflowInNamespace(ctx context.Context,
		namespace string,
		opts tclient.UpdateWorkflowOptions,
	) (tclient.WorkflowUpdateHandle, error)

	UpdateWithStartWorkflowInNamespace(ctx context.Context,
		namespace string,
		opts tclient.UpdateWithStartWorkflowOptions) (tclient.WorkflowUpdateHandle, error)

	GetWorkflowUpdateHandleInNamespace(namespace string, ref tclient.GetWorkflowUpdateHandleOptions) tclient.WorkflowUpdateHandle

	QueryWorkflowInNamespace(ctx context.Context,
		namespace string,
		workflowID string,
		runID string,
		queryType string,
		args ...interface{}) (converter.EncodedValue, error)

	QueryWorkflowWithOptionsInNamespace(ctx context.Context,
		namespace string,
		request *tclient.QueryWorkflowWithOptionsRequest) (*tclient.QueryWorkflowWithOptionsResponse, error)
}

func (t *temporal) getOpts() tclient.Options {
	opts := tclient.Options{
		HostPort:           t.Addr,
		Logger:             temporalzap.NewLogger(t.Logger),
		DataConverter:      t.Converter,
		ContextPropagators: t.propagators,
	}
	if t.tallyScope != nil {
		opts.MetricsHandler = sdktally.NewMetricsHandler(t.tallyScope)
	}

	return opts
}

// getClient returns a temporal client from memory, or creates a new one and caches it
func (t *temporal) getClient() (tclient.Client, error) {
	t.clientOnce.Do(func() {
		opts := t.getOpts()
		opts.Namespace = t.Namespace
		opts.DataConverter = t.Converter

		tc, err := tclient.Dial(opts)
		if err != nil {
			t.clientErr = fmt.Errorf("unable to dial temporal: %w", err)
		} else {
			t.Client = tc
		}
	})

	return t.Client, t.clientErr
}
