package workflowmetrics

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	deploymentRefreshInterval = 5 * time.Minute
	deploymentSnapshotTTL     = 15 * time.Minute
	// Leave the SDK's overflow slot unused rather than exporting partial totals.
	maxDeploymentSeries = 1999
)

type deploymentComponent struct {
	org, app, component string
}

type deploymentBucket struct {
	deploymentComponent
	state string
}

type deploymentSnapshot struct {
	attempts map[deploymentBucket]int64
	applies  map[deploymentComponent]int64
	latest   map[deploymentBucket]int64
}

// Materialize candidates so outer joins cannot repeat per-target latest lookups.
// Exclude plan-only runs before LIMIT so a newer plan cannot hide a deployment.
const deploymentSnapshotQuery = `
WITH components AS MATERIALIZED (
  SELECT ic.id, ic.install_id, c.org_id, c.app_id, c.id AS component_id
  FROM install_components ic
  JOIN installs i ON i.id = ic.install_id AND i.deleted_at = 0
  JOIN components c ON c.id = ic.component_id AND c.deleted_at = 0
    AND c.org_id = ic.org_id AND c.app_id = i.app_id
  JOIN apps a ON a.id = c.app_id AND a.deleted_at = 0
  JOIN orgs o ON o.id = c.org_id AND o.deleted_at = 0
  WHERE ic.deleted_at = 0 AND i.org_id = ic.org_id
), candidates AS MATERIALIZED (
  SELECT 'attempts' AS family, d.install_component_id, d.install_workflow_id,
    d.status, d.status_v2
  FROM install_deploys d
  WHERE d.deleted_at = 0 AND d.type = 'apply'
    AND d.created_at >= $1 AND d.created_at < $2
  UNION ALL
  SELECT 'applies', d.install_component_id, d.install_workflow_id, d.status, d.status_v2
  FROM install_deploys d
  WHERE d.deleted_at = 0 AND d.type = 'apply'
    AND d.applied_at >= $1 AND d.applied_at < $2
  UNION ALL
  SELECT 'latest', d.install_component_id, d.install_workflow_id, d.status, d.status_v2
  FROM components c
  CROSS JOIN LATERAL (
    SELECT d.install_component_id, d.install_workflow_id, d.status, d.status_v2
    FROM install_deploys d
    JOIN install_workflows w ON w.id = d.install_workflow_id
      AND w.deleted_at = 0 AND w.plan_only IS NOT TRUE
      AND w.owner_type = 'installs' AND w.owner_id = c.install_id
    WHERE d.install_component_id = c.id AND d.deleted_at = 0 AND d.type = 'apply'
      AND d.created_at < $2
    ORDER BY d.created_at DESC, d.id DESC
    LIMIT 1
  ) d
), classified AS (
  SELECT d.family, c.org_id, c.app_id, c.component_id,
    CASE
      WHEN d.family = 'applies' THEN ''
      WHEN d.status = 'health-failed'
        AND COALESCE(NULLIF(d.status_v2->>'status', ''), 'error') IN ('error', 'health-failed')
        THEN 'health_failed'
      ELSE CASE COALESCE(NULLIF(d.status_v2->>'status', ''), d.status)
        WHEN 'active' THEN 'applied'
        WHEN 'error' THEN 'error'
        WHEN 'health-failed' THEN 'health_failed'
        WHEN 'retried' THEN 'retried'
        WHEN 'cancelled' THEN 'cancelled'
        WHEN 'executing' THEN 'executing'
        WHEN 'pending-approval' THEN 'awaiting_approval'
        WHEN 'approval-denied' THEN 'skipped'
        WHEN 'noop' THEN 'skipped'
        WHEN 'auto-skipped' THEN 'skipped'
        WHEN 'pending' THEN 'pending'
        WHEN 'queued' THEN 'pending'
        WHEN 'planning' THEN 'pending'
        WHEN 'syncing' THEN 'pending'
        ELSE 'unknown'
      END
    END AS state
  FROM candidates d
  JOIN components c ON c.id = d.install_component_id
  JOIN install_workflows w ON w.id = d.install_workflow_id
    AND w.deleted_at = 0 AND w.plan_only IS NOT TRUE
    AND w.owner_type = 'installs' AND w.owner_id = c.install_id
)
SELECT family, org_id, app_id, component_id, state, count(*)
FROM classified
GROUP BY family, org_id, app_id, component_id, state
LIMIT $3`

func readDeploymentSnapshot(ctx context.Context, conn *pgx.Conn, now time.Time) (deploymentSnapshot, error) {
	result := deploymentSnapshot{
		attempts: make(map[deploymentBucket]int64),
		applies:  make(map[deploymentComponent]int64),
		latest:   make(map[deploymentBucket]int64),
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = tx.Rollback(cleanupCtx)
	}()
	if _, err := tx.Exec(ctx, `SET LOCAL statement_timeout = '5s'; SET LOCAL lock_timeout = '100ms';
SET LOCAL max_parallel_workers_per_gather = 0; SET LOCAL work_mem = '4MB'`); err != nil {
		return result, err
	}
	rows, err := tx.Query(ctx, deploymentSnapshotQuery, now.Add(-24*time.Hour), now, 3*maxDeploymentSeries+1)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var family string
		var key deploymentBucket
		var count int64
		if err := rows.Scan(&family, &key.org, &key.app, &key.component, &key.state, &count); err != nil {
			return result, err
		}
		if _, ok := result.applies[key.deploymentComponent]; !ok {
			result.applies[key.deploymentComponent] = 0
		}
		switch family {
		case "attempts":
			result.attempts[key] = count
		case "applies":
			result.applies[key.deploymentComponent] = count
		case "latest":
			result.latest[key] = count
		}
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	if err := result.retainZeros(deploymentSnapshot{}); err != nil {
		return result, err
	}
	return result, tx.Commit(ctx)
}

func (s *deploymentSnapshot) retainZeros(previous deploymentSnapshot) error {
	for _, counts := range []map[deploymentBucket]int64{s.attempts, s.latest, previous.attempts, previous.latest} {
		for key := range counts {
			if _, exists := s.applies[key.deploymentComponent]; !exists {
				continue
			}
			if _, exists := s.attempts[key]; !exists {
				s.attempts[key] = 0
			}
			if _, exists := s.latest[key]; !exists {
				s.latest[key] = 0
			}
		}
	}
	if len(s.attempts) > maxDeploymentSeries || len(s.applies) > maxDeploymentSeries || len(s.latest) > maxDeploymentSeries {
		return fmt.Errorf("deployment snapshot exceeds %d series per metric", maxDeploymentSeries)
	}
	return nil
}

func (key deploymentBucket) attributes() metric.ObserveOption {
	return metric.WithAttributes(attribute.String("org.id", key.org), attribute.String("app.id", key.app),
		attribute.String("component.id", key.component), attribute.String("deployment.state", key.state))
}

func (r *reporter) refreshDeploymentsIfDue(ctx context.Context, conn *pgx.Conn) error {
	if time.Since(r.deploymentAttempt) < deploymentRefreshInterval {
		return nil
	}
	r.deploymentAttempt = time.Now()
	// Server timeout leaves time to roll back without discarding the leader session.
	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	return r.refreshDeployments(ctx, conn)
}

func (r *reporter) refreshDeployments(ctx context.Context, conn *pgx.Conn) error {
	started := time.Now()
	values, err := readDeploymentSnapshot(ctx, conn, started)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := values.retainZeros(r.deployments); err != nil {
		return err
	}
	// Completion time lets consumers reject old series retained by backend
	// lookback, including observations emitted while this query was running.
	r.deployments, r.deploymentsCollected, r.deploymentsActive = values, time.Now(), true
	return nil
}
