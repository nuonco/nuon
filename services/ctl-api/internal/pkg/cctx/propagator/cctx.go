package propagator

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

var _ workflow.ContextPropagator = (*propagator)(nil)

const (
	propagationHeader string = "cctx"
)

func (s *propagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	pl, err := FetchPayload(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to fetch payload from context")
	}

	payload, err := s.dataConverter.ToPayload(pl)
	if err != nil {
		return errors.Wrap(err, "unable to convert payload")
	}

	writer.Set(propagationHeader, payload)
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
func (s *propagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	acctID, err := cctx.AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return err
	}

	traceID := cctx.TraceIDFromContext(ctx)
	if traceID == "" {
		u7 := uuid.Must(uuid.NewV7())
		traceID = u7.String()
		cctx.SetTraceIDWorkflowContext(ctx, traceID)
	}
	logStream, _ := cctx.GetLogStreamWorkflow(ctx)

	// Get a workflow-context-aware copy of the data converter so that
	// DataConverterWithoutDeadlockDetection can pause the deadlock detector
	// during serialization. Without this, the converter's context field is nil
	// and PauseDeadlockDetector is a no-op, which allows json-iterator's
	// first-time reflection cache build to trigger TMPRL1101 panics.
	dc := withWorkflowContext(ctx, s.dataConverter)

	payload, err := dc.ToPayload(Payload{
		OrgID:     orgID,
		AccountID: acctID,
		TraceID:   traceID,
		LogStream: logStream,
	})
	if err != nil {
		return err
	}
	writer.Set(propagationHeader, payload)
	return nil
}

func (s *propagator) getPayload(reader workflow.HeaderReader) (*Payload, error) {
	return s.getPayloadFromDC(s.dataConverter, reader)
}

func (s *propagator) getPayloadWithWorkflowContext(ctx workflow.Context, reader workflow.HeaderReader) (*Payload, error) {
	dc := withWorkflowContext(ctx, s.dataConverter)
	return s.getPayloadFromDC(dc, reader)
}

// withWorkflowContext returns a DataConverter that pauses the deadlock detector
// during serialization calls. If the converter doesn't implement ContextAware
// it is returned as-is.
func withWorkflowContext(ctx workflow.Context, dc converter.DataConverter) converter.DataConverter {
	if ca, ok := dc.(workflow.ContextAware); ok {
		return ca.WithWorkflowContext(ctx)
	}
	return dc
}

func (s *propagator) getPayloadFromDC(dc converter.DataConverter, reader workflow.HeaderReader) (*Payload, error) {
	value, ok := reader.Get(propagationHeader)
	if !ok {
		return nil, fmt.Errorf("no propagation key (%s) set for cctx", propagationHeader)
	}

	var payload Payload
	if err := dc.FromPayload(value, &payload); err != nil {
		return nil, errors.Wrap(err, "unable to convert payload")
	}

	if payload.TraceID == "" {
		u7 := uuid.Must(uuid.NewV7())
		payload.TraceID = u7.String()
	}

	return &payload, nil
}

// Extract extracts values from headers and puts them into context
func (s *propagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	payload, err := s.getPayload(reader)
	if err != nil {
		return nil, err
	}

	ctx = cctx.SetAccountIDContext(ctx, payload.AccountID)
	ctx = cctx.SetOrgIDContext(ctx, payload.OrgID)
	ctx = cctx.SetTraceIDContext(ctx, payload.TraceID)

	if payload.LogStream != nil {
		ctx = cctx.SetLogStreamContext(ctx, payload.LogStream)
	}

	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into context
func (s *propagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	payload, err := s.getPayloadWithWorkflowContext(ctx, reader)
	if err != nil {
		return nil, err
	}

	ctx = cctx.SetAccountIDWorkflowContext(ctx, payload.AccountID)
	ctx = cctx.SetOrgIDWorkflowContext(ctx, payload.OrgID)
	ctx = cctx.SetTraceIDWorkflowContext(ctx, payload.TraceID)

	if payload.LogStream != nil {
		ctx = cctx.SetLogStreamWorkflowContext(ctx, payload.LogStream)
	}

	return ctx, nil
}
