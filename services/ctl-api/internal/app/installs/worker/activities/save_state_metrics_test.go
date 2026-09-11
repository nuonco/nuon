package activities

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	typesstate "github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func TestSaveStateRecordsCreateError(t *testing.T) {
	wantErr := errors.New("create failed")
	a, reader, shutdown := newSaveStateMetricsHarness(t, func(tx *gorm.DB) { tx.AddError(wantErr) })
	defer shutdown()

	got, err := a.SaveState(cctx.SetOrgIDContext(context.Background(), "org-id"), validSaveStateRequest())
	require.Nil(t, got)
	require.ErrorIs(t, err, wantErr)
	require.Equal(t, int64(1), saveStateOperationCount(t, reader, "save", "error"))
}

func TestSaveStateRecordsSuccess(t *testing.T) {
	a, reader, shutdown := newSaveStateMetricsHarness(t, func(tx *gorm.DB) { tx.RowsAffected = 1 })
	defer shutdown()

	got, err := a.SaveState(cctx.SetOrgIDContext(context.Background(), "org-id"), validSaveStateRequest())
	require.NoError(t, err)
	require.Equal(t, "install-id", got.InstallID)
	require.Equal(t, int64(1), saveStateOperationCount(t, reader, "save", "success"))
}

func validSaveStateRequest() *SaveStateRequest {
	return &SaveStateRequest{
		State:           &typesstate.State{ID: "install-id"},
		InstallID:       "install-id",
		TriggeredByID:   "trigger-id",
		TriggeredByType: "workflow",
		GeneratedBy:     app.InstallStateGenerateSourceStateManager,
	}
}

func newSaveStateMetricsHarness(t *testing.T, create func(*gorm.DB)) (*Activities, *sdkmetric.ManualReader, func()) {
	t.Helper()
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused"}), &gorm.Config{
		DisableAutomaticPing:   true,
		DryRun:                 true,
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)
	require.NoError(t, db.Callback().Create().Replace("gorm:before_create", func(*gorm.DB) {}))
	require.NoError(t, db.Callback().Create().Replace("gorm:create", create))
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	return &Activities{db: db, stateMetrics: pkgstate.NewMetrics(provider)}, reader,
		func() { require.NoError(t, provider.Shutdown(context.Background())) }
}

func saveStateOperationCount(t *testing.T, reader *sdkmetric.ManualReader, operation, outcome string) int64 {
	t.Helper()
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != "nuon.install.state.operations" {
				continue
			}
			for _, point := range m.Data.(metricdata.Sum[int64]).DataPoints {
				op, _ := point.Attributes.Value("operation")
				result, _ := point.Attributes.Value("outcome")
				if op.AsString() == operation && result.AsString() == outcome {
					return point.Value
				}
			}
		}
	}
	return 0
}
