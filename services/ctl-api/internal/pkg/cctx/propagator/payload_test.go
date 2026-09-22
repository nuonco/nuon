package propagator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"

	"github.com/nuonco/nuon/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// oldPayload is the Payload shape shipped before temporaljson tags were added.
type oldPayload struct {
	OrgID     string         `json:"org_id"`
	AccountID string         `json:"account_id"`
	TraceID   string         `json:"trace_id"`
	LogStream *app.LogStream `json:"log_stream,omitempty"`
}

type headerMap map[string]*commonpb.Payload

func (h headerMap) Get(key string) (*commonpb.Payload, bool) { p, ok := h[key]; return p, ok }
func (h headerMap) ForEachKey(fn func(string, *commonpb.Payload) error) error {
	for k, v := range h {
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}
func (h headerMap) Set(key string, value *commonpb.Payload) { h[key] = value }

func newTestPropagator() *propagator {
	return &propagator{
		dataConverter: converter.NewCompositeDataConverter(
			converter.NewNilPayloadConverter(),
			converter.NewByteSlicePayloadConverter(),
			dataconverter.NewJSONConverter(),
		),
	}
}

func TestExtract_LegacyHeaderKeys(t *testing.T) {
	p := newTestPropagator()

	pl, err := p.dataConverter.ToPayload(oldPayload{
		OrgID:     "orgacme",
		AccountID: "accacme",
		TraceID:   "trace-1",
		LogStream: &app.LogStream{ID: "lgs1"},
	})
	require.NoError(t, err)
	hdr := headerMap{propagationHeader: pl}

	ctx, err := p.Extract(context.Background(), hdr)
	require.NoError(t, err)

	acct, err := cctx.AccountIDFromContext(ctx)
	require.NoError(t, err)
	require.Equal(t, "accacme", acct)
	org, err := cctx.OrgIDFromContext(ctx)
	require.NoError(t, err)
	require.Equal(t, "orgacme", org)
	require.Equal(t, "trace-1", cctx.TraceIDFromContext(ctx))
	ls, err := cctx.GetLogStreamContext(ctx)
	require.NoError(t, err)
	require.Equal(t, "lgs1", ls.ID)
}

func TestExtract_CurrentHeaderKeys(t *testing.T) {
	p := newTestPropagator()

	pl, err := p.dataConverter.ToPayload(Payload{
		OrgID:     "orgacme",
		AccountID: "accacme",
		TraceID:   "trace-1",
		QueueID:   "queacme",
	})
	require.NoError(t, err)
	hdr := headerMap{propagationHeader: pl}

	ctx, err := p.Extract(context.Background(), hdr)
	require.NoError(t, err)

	acct, err := cctx.AccountIDFromContext(ctx)
	require.NoError(t, err)
	require.Equal(t, "accacme", acct)
	require.Equal(t, "queacme", cctx.QueueIDFromContext(ctx))
}
