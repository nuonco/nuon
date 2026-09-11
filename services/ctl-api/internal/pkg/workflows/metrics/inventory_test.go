package workflowmetrics

import (
	"context"
	"testing"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestInventoryObservation(t *testing.T) {
	r, reader := testReporter(t)
	require.Empty(t, collect(t, reader))
	r.inventoryCollected, r.inventoryActive = time.Now(), true
	r.inventory = map[inventoryBucket]value{
		{"org", "active"}:              {count: 7},
		{"queue", "awaiting_dispatch"}: {count: 3, oldest: time.Unix(100, 0)},
	}
	observed := collect(t, reader)
	counts := observed["nuon.inventory.current"].Data.(metricdata.Gauge[int64]).DataPoints
	require.Len(t, counts, 19)
	var total int64
	for _, point := range counts {
		require.Equal(t, 2, point.Attributes.Len())
		total += point.Value
	}
	require.EqualValues(t, 7, total)
	oldest := observed["nuon.queue.oldest_created_at"].Data.(metricdata.Gauge[float64]).DataPoints
	require.Len(t, oldest, 1)
	require.Equal(t, float64(100), oldest[0].Value)
	r.inventory = nil
	observed = collect(t, reader)
	require.NotContains(t, observed, "nuon.queue.oldest_created_at")
	for _, point := range observed["nuon.queue.current"].Data.(metricdata.Gauge[int64]).DataPoints {
		require.Zero(t, point.Value)
	}
	r.inventoryCollected = time.Now().Add(-inventorySnapshotTTL - time.Second)
	require.Len(t, collect(t, reader), 1)
	r.inventoryCollected = time.Now()
	r.release(nil)
	require.Len(t, collect(t, reader), 1)
}

func TestPostgresInventory(t *testing.T) {
	// INTEGRATION: session-local fixtures use the deployed PostgreSQL column types.
	tests.SkipIfNotIntegration(t)
	ctx := context.Background()
	cfg, err := internal.NewConfig()
	require.NoError(t, err)
	conn, err := psql.NewPrimaryListenerConn(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close(ctx)) })
	_, err = conn.Exec(ctx, `
CREATE TEMP TABLE orgs AS SELECT id, status, status_v2, deleted_at FROM public.orgs WITH NO DATA;
CREATE TEMP TABLE apps AS SELECT id, org_id, status, status_v2, deleted_at FROM public.apps WITH NO DATA;
CREATE TEMP TABLE installs AS SELECT id, org_id, app_id, lifecycle_phase, deleted_at FROM public.installs WITH NO DATA;
CREATE TEMP TABLE queues AS SELECT id, deleted_at FROM public.queues WITH NO DATA;
CREATE TEMP TABLE queue_signals AS SELECT id, org_id, queue_id, status, enqueued, created_at, deleted_at FROM public.queue_signals WITH NO DATA;
INSERT INTO orgs VALUES ('org-a','error','{"status":"active"}',0),('org-b','active','{}',0),('org-deleted','active','{}',1);
INSERT INTO apps VALUES ('app-a','org-a','active','{"status":"error"}',0),('app-b','org-b','future','{}',0),('app-hidden','org-deleted','active','{}',0),('app-deleted','org-a','active','{}',1);
INSERT INTO installs VALUES ('install-a','org-a','app-a','{"phase":"provisioned"}',0),('install-b','org-a','app-a','{}',0),('install-hidden','org-a','app-deleted','{}',0),('install-wrong-org','org-b','app-a','{}',0),('install-deleted','org-a','app-a','{}',1);
INSERT INTO queues VALUES ('queue-a',0),('queue-deleted',1);
INSERT INTO queue_signals VALUES
('signal-a','org-a','queue-a','{"status":"queued"}',false,'2000-01-01',0),
('signal-b','org-a','queue-a','{"status":"queued"}',false,'2001-01-01',0),
('signal-c','org-a','queue-a','{"status":"queued"}',true,'2002-01-01',0),
('signal-d','org-a','queue-a','{"status":"in-progress"}',false,'2003-01-01',0),
('terminal','org-a','queue-a','{"status":"error"}',false,'1999-01-01',0),
('deleted','org-a','queue-a','{"status":"queued"}',false,'1999-01-01',1),
('deleted-parent','org-a','queue-deleted','{"status":"queued"}',false,'1999-01-01',0),
('deleted-org','org-deleted','queue-a','{"status":"queued"}',false,'1999-01-01',0);`)
	require.NoError(t, err)
	values, err := readInventory(ctx, conn)
	require.NoError(t, err)
	want := map[inventoryBucket]value{
		{"org", "active"}:              {count: 2},
		{"app", "error"}:               {count: 1},
		{"app", "unknown"}:             {count: 1},
		{"install", "provisioned"}:     {count: 1},
		{"install", "unknown"}:         {count: 1},
		{"queue", "awaiting_dispatch"}: {2, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"queue", "queued"}:            {1, time.Date(2002, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"queue", "executing"}:         {1, time.Date(2003, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	require.Len(t, values, len(want))
	for key, expected := range want {
		require.Contains(t, values, key)
		require.Equal(t, expected.count, values[key].count)
		require.True(t, expected.oldest.Equal(values[key].oldest), "oldest timestamp for %v", key)
	}
	r, _ := testReporter(t)
	require.NoError(t, r.refreshInventoryIfDue(ctx, conn))
	stamp := r.inventoryCollected
	_, err = conn.Exec(ctx, "ALTER TABLE pg_temp.orgs RENAME COLUMN status TO unavailable")
	require.NoError(t, err)
	r.inventoryAttempt = time.Time{}
	require.Error(t, r.refreshInventoryIfDue(ctx, conn))
	require.Equal(t, stamp, r.inventoryCollected)
	require.NoError(t, r.refreshInventoryIfDue(ctx, nil), "failed refresh must not retry every 30 seconds")
	_, err = conn.Exec(ctx, "ALTER TABLE pg_temp.orgs RENAME COLUMN unavailable TO status; UPDATE pg_temp.orgs SET deleted_at=1")
	require.NoError(t, err)
	values, err = readInventory(ctx, conn)
	require.NoError(t, err)
	require.Empty(t, values)
}
