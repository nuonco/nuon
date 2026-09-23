package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func TestRepairActivityContext(t *testing.T) {
	em := &app.QueueEmitter{CreatedByID: "accacme", OrgID: "orgacme"}

	ctx := repairActivityContext(context.Background(), em)
	acct, err := cctx.AccountIDFromContext(ctx)
	require.NoError(t, err)
	require.Equal(t, "accacme", acct)
	org, err := cctx.OrgIDFromContext(ctx)
	require.NoError(t, err)
	require.Equal(t, "orgacme", org)

	// existing values win
	ctx = cctx.SetAccountIDContext(context.Background(), "accother")
	ctx = repairActivityContext(ctx, em)
	acct, _ = cctx.AccountIDFromContext(ctx)
	require.Equal(t, "accother", acct)
}
