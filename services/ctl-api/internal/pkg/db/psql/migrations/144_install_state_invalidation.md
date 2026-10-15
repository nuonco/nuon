# Install-state publication rollout

This is a coordinated upgrade, not a binary-only rolling upgrade or rollback.
Old binaries write snapshot metadata directly; new binaries use `install_state_heads`.
Do not run old and new state writers together.

## Deploy

1. Pause ingress and drain ctl-api API processes, workers, and other source writers.
   Preserve Temporal histories: legacy `SaveState` activity requests remain supported
   and regenerate from current sources rather than saving their cached payload.
2. Run migrations from one new-binary process before starting the remaining processes.
   Source writers must use PostgreSQL's default `READ COMMITTED` isolation; the
   source barrier relies on fresh snapshots for the trigger's target lookup.
   Check `SHOW default_transaction_isolation` and do not enable repeatable-read
   source transactions with this protocol.
   Migration 144 installs the coordination functions. Each 145 migration installs
   one table's triggers in its own transaction with `lock_timeout = '2s'`.
   Successful stages remain committed; a blocked stage rolls back and fails startup.
   Remove the blocker, then retry migrations. Do not raise the timeout or resume on
   partial migration success.
3. Confirm that every 144/145 stage is `applied`, then start the new processes and
   resume traffic. No head backfill is required: legacy snapshots bootstrap lazily.
4. Compare write latency, lock waits, PostgreSQL deadlocks, state regeneration rate,
   publication conflicts and Temporal activity errors with the pre-deploy baseline.
   Stop the rollout if write latency or lock waits rise materially. Investigate any
   `40P01` or `55P03` error before proceeding. Production load measurements are still
   required; the integration suite is not a write-throughput benchmark.

Runner-job status is not an invalidation input. Execution status only invalidates
when finished/non-finished priority changes. Actions expose their run status, so
those status changes must invalidate. Secrets targets match only secrets-sync jobs.
Transition tables still contain PostgreSQL's row images, but our materialized
projection carries IDs/FKs and eligibility fields, not output or template payloads.

Head invalidation uses one sorted upsert pass for both missing and existing rows.
It does not guarantee deadlock freedom for arbitrary multi-statement transactions
that visit installs in opposite orders. A `40P01` abort must be retried at the whole
source transaction boundary, including its source mutation; never retry just the
trigger or publish after a failed write. This change does not add blanket retries
to API mutations or transactions with external side effects.

## Roll back

1. Pause ingress and stop **all** new/old readers, workers and source writers.
2. Back up the database. Verify the intended database connection, then run:

   ```sh
   psql "$DATABASE_URL" -v install_state_writers_stopped=on \
     -f services/ctl-api/internal/pkg/db/psql/migrations/rollback_install_state_publication.sql
   ```

   The script refuses current blob-only snapshots: restore their inline JSON from
   the blobs first. It drops only the state-invalidation triggers, one autocommit
   statement at a time under a two-second lock timeout. It marks live snapshots
   stale in the old format, detaches their blob pointers where inline JSON exists
   (so old read-repair cannot re-read older bytes), clears heads and removes only
   this feature's migration records/functions. It keeps snapshots, stored blobs,
   source records, the head table and migration 112's job-notification trigger.
3. If any step fails, keep traffic paused and fix the blocker before rerunning the
   script. Its statements are repeatable; never resume a partially rolled-back DB.
4. Start the previous binary and verify repeated reads after an input update.
   A later upgrade reruns 144/145 and lazily bootstraps the stale legacy snapshots.

The script does not give the old binary the new publication guarantees; it only
converts coordination metadata back to the format that binary understands.
