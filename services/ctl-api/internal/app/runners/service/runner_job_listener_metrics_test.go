package service

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap"
)

func TestPostgresRunnerJobListenerMetrics(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	cfg, err := internal.NewConfig()
	require.NoError(t, err)
	admin, err := psql.NewPrimaryListenerConn(ctx, cfg)
	require.NoError(t, err)
	defer admin.Close(context.Background())
	dbName := "listener_metrics_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	_, err = admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{dbName}.Sanitize())
	require.NoError(t, err)
	defer func() {
		_, err := admin.Exec(context.Background(), "DROP DATABASE "+pgx.Identifier{dbName}.Sanitize()+" WITH (FORCE)")
		require.NoError(t, err)
	}()
	cfg.DBName = dbName
	conn, err := psql.NewPrimaryListenerConn(ctx, cfg)
	require.NoError(t, err)
	defer conn.Close(context.Background())
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer provider.Shutdown(context.Background())
	mw, err := metrics.New(validator.New(), metrics.WithDisable(true), metrics.WithLogger(zap.NewNop()))
	require.NoError(t, err)
	rl := &RunnerJobNotifyListener{cfg: cfg, l: zap.NewNop(), mw: mw, registry: NewRunnerJobWakeRegistry(), metrics: newRunnerJobListenerMetrics(provider), done: make(chan struct{})}
	listenerCtx, stop := context.WithCancel(ctx)
	go rl.run(listenerCtx)
	defer func() { stop(); <-rl.done }()
	collect := func() metricdata.ResourceMetrics {
		var data metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &data))
		return data
	}
	failures := func() int64 {
		return metricByName(t, collect(), "nuon.runner.job_tail.listener.failures").Data.(metricdata.Sum[int64]).DataPoints[0].Value
	}
	require.Eventually(t, func() bool { return metricGauge(t, collect(), "nuon.runner.job_tail.listener.connected") == 1 }, 5*time.Second, 10*time.Millisecond)
	wake, unsubscribe := rl.registry.Subscribe("runner-test")
	defer unsubscribe()
	for _, payload := range []string{"not-json", `{}`, `{"runner_id":"runner-test"}`} {
		_, err = conn.Exec(ctx, "SELECT pg_notify($1, $2)", runnerJobNotifyChannel, payload)
		require.NoError(t, err)
	}
	select {
	case <-wake:
	case <-time.After(5 * time.Second):
		t.Fatal("notification did not wake subscriber")
	}
	assertOutcomePopulation(t, collect(), "nuon.runner.job_tail.listener.notifications", []string{"valid", "invalid"}, map[string]int64{"valid": 1, "invalid": 2})
	require.Zero(t, failures())
	var pid uint32
	require.NoError(t, conn.QueryRow(ctx, "SELECT pid FROM pg_stat_activity WHERE datname=$1 AND pid <> pg_backend_pid()", dbName).Scan(&pid))
	var terminated bool
	require.NoError(t, conn.QueryRow(ctx, "SELECT pg_terminate_backend($1)", pid).Scan(&terminated))
	require.True(t, terminated)
	require.Eventually(t, func() bool { return failures() == 1 && rl.metrics.connected.Load() == 1 }, 5*time.Second, 10*time.Millisecond)
	var reconnectedPID uint32
	require.NoError(t, conn.QueryRow(ctx, "SELECT pid FROM pg_stat_activity WHERE datname=$1 AND pid <> pg_backend_pid()", dbName).Scan(&reconnectedPID))
	require.NotEqual(t, pid, reconnectedPID)
	if !testing.Short() {
		// Exercise the real session-age timeout without altering production timing.
		require.Eventually(t, func() bool {
			var currentPID uint32
			if err := conn.QueryRow(ctx, "SELECT pid FROM pg_stat_activity WHERE datname=$1 AND pid <> pg_backend_pid()", dbName).Scan(&currentPID); err != nil {
				return false
			}
			return currentPID != reconnectedPID && rl.metrics.connected.Load() == 1
		}, listenerSessionMaxAge+35*time.Second, time.Second)
		require.EqualValues(t, 1, failures())
	}
	stop()
	select {
	case <-rl.done:
	case <-time.After(5 * time.Second):
		t.Fatal("listener did not stop")
	}
	require.Zero(t, metricGauge(t, collect(), "nuon.runner.job_tail.listener.connected"))
	require.EqualValues(t, 1, failures())
}
