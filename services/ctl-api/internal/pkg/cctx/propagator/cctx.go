package propagator

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

const (
	propagationHeader string = "cctx"
)

type Payload struct {
	OrgID     string `json:"org_id"`
	AccountID string `json:"account_id"`
}

func (s *propagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	acctID, err := cctx.AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return err
	}

	payload, err := converter.GetDefaultDataConverter().ToPayload(Payload{
		OrgID:     orgID,
		AccountID: acctID,
	})
	if err != nil {
		return err
	}

	writer.Set(propagationHeader, payload)
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
func (s *propagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	acctID, err := cctx.AccountIDFromWorkflowContext(ctx)
	if err != nil {
		return err
	}

	orgID, err := cctx.OrgIDFromWorkflowContext(ctx)
	if err != nil {
		return err
	}

	payload, err := converter.GetDefaultDataConverter().ToPayload(Payload{
		OrgID:     orgID,
		AccountID: acctID,
	})
	if err != nil {
		return err
	}
	writer.Set(propagationHeader, payload)
	return nil
}

func (s *propagator) getPayload(reader workflow.HeaderReader) (*Payload, error) {
	value, ok := reader.Get(propagationHeader)
	if !ok {
		return nil, fmt.Errorf("no propagation key (%s) set for cctx", propagationHeader)
	}

	var payload Payload
	if err := converter.GetDefaultDataConverter().FromPayload(value, &payload); err != nil {
		return nil, errors.Wrap(err, "unable to convert payload")
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

	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into context
func (s *propagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	payload, err := s.getPayload(reader)
	if err != nil {
		return nil, err
	}

	ctx = cctx.SetAccountIDWorkflowContext(ctx, payload.AccountID)
	ctx = cctx.SetOrgIDWorkflowContext(ctx, payload.OrgID)

	return ctx, nil
}
