package helpers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func TestGetInstallStateRecordsLookupError(t *testing.T) {
	wantErr := errors.New("lookup failed")
	h, reader, shutdown := newGetStateMetricsHarness(t, func(tx *gorm.DB) {
		tx.AddError(wantErr)
	})
	defer shutdown()

	got, err := h.GetInstallState(context.Background(), "install-id", false, false)
	require.Nil(t, got)
	require.ErrorIs(t, err, wantErr)
	require.Equal(t, int64(1), getStateOperationCount(t, reader, "get", "error"))
}

func TestGetInstallStateRecordsCachedSuccess(t *testing.T) {
	h, reader, shutdown := newGetStateMetricsHarness(t, func(tx *gorm.DB) {
		switch dest := tx.Statement.Dest.(type) {
		case *app.InstallState:
			dest.InstallID = "install-id"
			dest.State = &state.State{ID: "install-id", Name: "payments"}
		case *app.Install:
			dest.ID = "install-id"
			dest.Labels = map[string]string{"environment": "test"}
		default:
			t.Fatalf("unexpected query destination %T", dest)
		}
		tx.RowsAffected = 1
	})
	defer shutdown()

	got, err := h.GetInstallState(context.Background(), "install-id", false, false)
	require.NoError(t, err)
	require.Equal(t, "payments", got.Name)
	require.Equal(t, map[string]string{"environment": "test"}, got.Labels)
	require.Equal(t, int64(1), getStateOperationCount(t, reader, "get", "success"))
}

func newGetStateMetricsHarness(t *testing.T, query func(*gorm.DB)) (*Helpers, *sdkmetric.ManualReader, func()) {
	t.Helper()
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused"}), &gorm.Config{
		DisableAutomaticPing: true,
		DryRun:               true,
	})
	require.NoError(t, err)
	require.NoError(t, db.Callback().Query().Replace("gorm:query", query))
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	return &Helpers{
		db:           db,
		cfg:          &internal.Config{},
		l:            zap.NewNop(),
		stateMetrics: pkgstate.NewMetrics(provider),
	}, reader, func() { require.NoError(t, provider.Shutdown(context.Background())) }
}

func getStateOperationCount(t *testing.T, reader *sdkmetric.ManualReader, operation, outcome string) int64 {
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
